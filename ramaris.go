package ramaris

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const defaultBaseURL = "https://www.ramaris.app/api/v1"

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// Client is the Ramaris API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	mu        sync.RWMutex
	rateLimit *RateLimitInfo
}

// NewClient creates a new Ramaris API client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// RateLimit returns the most recent rate limit info, or nil if no request has been made.
func (c *Client) RateLimit() *RateLimitInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.rateLimit == nil {
		return nil
	}
	cp := *c.rateLimit
	return &cp
}

// doRequest performs an HTTP GET request with auth, rate limit tracking, and retry on 429/5xx.
func (c *Client) doRequest(ctx context.Context, path string, opts *ListOptions) (*http.Response, error) {
	reqURL := c.buildURL(path, opts)

	maxRetries := 3
	backoff := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("ramaris: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ramaris: request failed: %w", err)
		}

		c.updateRateLimit(resp.Header)

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Read error body for all error responses
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Parse error envelope
		var errResp errorResponse
		_ = json.Unmarshal(body, &errResp)

		// 429 — rate limited, return immediately with RetryAfter info
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, &RateLimitError{
				Code:       codeOrDefault(errResp.Error.Code, "RATE_LIMITED"),
				Message:    msgOrDefault(errResp.Error.Message, "rate limit exceeded"),
				StatusCode: 429,
				RetryAfter: errResp.Error.RetryAfter,
			}
		}

		// 5xx — server error, retry with backoff
		if resp.StatusCode >= 500 && i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
				continue
			}
		}

		// 4xx (non-429) — return immediately
		return nil, &Error{
			Code:       codeOrDefault(errResp.Error.Code, "UNKNOWN_ERROR"),
			Message:    msgOrDefault(errResp.Error.Message, fmt.Sprintf("HTTP %d", resp.StatusCode)),
			StatusCode: resp.StatusCode,
		}
	}

	return nil, fmt.Errorf("ramaris: max retries exceeded")
}

func (c *Client) buildURL(path string, opts *ListOptions) string {
	u := c.baseURL + path
	if opts == nil {
		return u
	}

	params := url.Values{}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	}
	if len(params) == 0 {
		return u
	}
	return u + "?" + params.Encode()
}

func (c *Client) updateRateLimit(h http.Header) {
	limit := h.Get("X-RateLimit-Limit")
	remaining := h.Get("X-RateLimit-Remaining")
	reset := h.Get("X-RateLimit-Reset")

	if limit == "" || remaining == "" || reset == "" {
		return
	}

	l, _ := strconv.Atoi(limit)
	r, _ := strconv.Atoi(remaining)
	rs, _ := strconv.Atoi(reset)

	c.mu.Lock()
	c.rateLimit = &RateLimitInfo{Limit: l, Remaining: r, Reset: rs}
	c.mu.Unlock()
}

func codeOrDefault(code, def string) string {
	if code != "" {
		return code
	}
	return def
}

func msgOrDefault(msg, def string) string {
	if msg != "" {
		return msg
	}
	return def
}

// --- Endpoint methods ---

// Health checks the API health.
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	resp, err := c.doRequest(ctx, "/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var h HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &h, nil
}

// ListStrategies lists strategies with optional pagination.
func (c *Client) ListStrategies(ctx context.Context, opts *ListOptions) (*ListResponse[StrategyListItem], error) {
	resp, err := c.doRequest(ctx, "/strategies", opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListResponse[StrategyListItem]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &result, nil
}

// GetStrategy gets a single strategy by share ID.
func (c *Client) GetStrategy(ctx context.Context, shareID string) (*Strategy, error) {
	resp, err := c.doRequest(ctx, "/strategies/"+shareID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope singleResponse[Strategy]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &envelope.Data, nil
}

// ListWatchlist lists the authenticated user's watchlist strategies.
func (c *Client) ListWatchlist(ctx context.Context, opts *ListOptions) (*ListResponse[WatchlistStrategy], error) {
	resp, err := c.doRequest(ctx, "/me/strategies/watchlist", opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListResponse[WatchlistStrategy]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &result, nil
}

// ListWallets lists wallets with optional pagination.
func (c *Client) ListWallets(ctx context.Context, opts *ListOptions) (*ListResponse[WalletListItem], error) {
	resp, err := c.doRequest(ctx, "/wallets", opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListResponse[WalletListItem]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &result, nil
}

// GetWallet gets a single wallet by ID.
func (c *Client) GetWallet(ctx context.Context, id int) (*Wallet, error) {
	resp, err := c.doRequest(ctx, "/wallets/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope singleResponse[Wallet]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &envelope.Data, nil
}

// GetProfile gets the authenticated user's profile.
func (c *Client) GetProfile(ctx context.Context) (*UserProfile, error) {
	resp, err := c.doRequest(ctx, "/me/profile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope singleResponse[UserProfile]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &envelope.Data, nil
}

// GetSubscription gets the authenticated user's subscription.
func (c *Client) GetSubscription(ctx context.Context) (*Subscription, error) {
	resp, err := c.doRequest(ctx, "/me/subscription", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope singleResponse[Subscription]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("ramaris: failed to decode response: %w", err)
	}
	return &envelope.Data, nil
}
