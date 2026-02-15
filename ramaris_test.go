package ramaris

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Error type tests ---

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:       "NOT_FOUND",
		Message:    "strategy not found",
		StatusCode: 404,
	}

	want := "ramaris: NOT_FOUND: strategy not found"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_ImplementsError(t *testing.T) {
	var _ error = (*Error)(nil)
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{
		Code:       "RATE_LIMITED",
		Message:    "too many requests",
		StatusCode: 429,
		RetryAfter: 30,
	}

	want := "ramaris: RATE_LIMITED: too many requests (retry after 30s)"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestRateLimitError_ImplementsError(t *testing.T) {
	var _ error = (*RateLimitError)(nil)
}

// --- JSON unmarshal tests ---

func TestStrategyListItem_Unmarshal(t *testing.T) {
	raw := `{
		"id": 1,
		"shareId": "abc123",
		"name": "Top Wallets",
		"description": "Tracks best performers",
		"roiPercent": 42.5,
		"lastActivityAt": "2025-01-15T10:30:00Z",
		"createdAt": "2024-12-01T00:00:00Z",
		"creator": {"nickname": "alice"},
		"stats": {"walletsTracked": 10, "totalSwaps": 250}
	}`

	var s StrategyListItem
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if s.ID != 1 {
		t.Errorf("ID = %d, want 1", s.ID)
	}
	if s.ShareID != "abc123" {
		t.Errorf("ShareID = %q, want %q", s.ShareID, "abc123")
	}
	if s.Name != "Top Wallets" {
		t.Errorf("Name = %q, want %q", s.Name, "Top Wallets")
	}
	if s.Description == nil || *s.Description != "Tracks best performers" {
		t.Errorf("Description = %v, want %q", s.Description, "Tracks best performers")
	}
	if s.ROIPercent == nil || *s.ROIPercent != 42.5 {
		t.Errorf("ROIPercent = %v, want 42.5", s.ROIPercent)
	}
	if s.Creator.Nickname != "alice" {
		t.Errorf("Creator.Nickname = %q, want %q", s.Creator.Nickname, "alice")
	}
	if s.Stats.WalletsTracked != 10 {
		t.Errorf("Stats.WalletsTracked = %d, want 10", s.Stats.WalletsTracked)
	}
}

func TestStrategyListItem_Unmarshal_NullFields(t *testing.T) {
	raw := `{
		"id": 2,
		"shareId": "def456",
		"name": "New Strategy",
		"description": null,
		"roiPercent": null,
		"lastActivityAt": null,
		"createdAt": "2025-01-01T00:00:00Z",
		"creator": {"nickname": "bob"},
		"stats": {"walletsTracked": 0, "totalSwaps": 0}
	}`

	var s StrategyListItem
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if s.Description != nil {
		t.Errorf("Description = %v, want nil", s.Description)
	}
	if s.ROIPercent != nil {
		t.Errorf("ROIPercent = %v, want nil", s.ROIPercent)
	}
	if s.LastActivityAt != nil {
		t.Errorf("LastActivityAt = %v, want nil", s.LastActivityAt)
	}
}

func TestStrategy_Unmarshal(t *testing.T) {
	raw := `{
		"id": 1,
		"shareId": "abc123",
		"name": "Top Wallets",
		"description": null,
		"roiPercent": 42.5,
		"lastActivityAt": null,
		"createdAt": "2024-12-01T00:00:00Z",
		"creator": {"nickname": "alice"},
		"stats": {"walletsTracked": 10, "totalSwaps": 250, "totalNotifications": 5},
		"status": "ACTIVE",
		"tags": ["base", "defi"]
	}`

	var s Strategy
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if s.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", s.Status, "ACTIVE")
	}
	if s.Stats.TotalNotifications != 5 {
		t.Errorf("Stats.TotalNotifications = %d, want 5", s.Stats.TotalNotifications)
	}
	if len(s.Tags) != 2 || s.Tags[0] != "base" {
		t.Errorf("Tags = %v, want [base defi]", s.Tags)
	}
}

func TestWallet_Unmarshal(t *testing.T) {
	raw := `{
		"id": 456,
		"winRate": 0.65,
		"realizedPnL": 1234.56,
		"createdAt": "2025-01-01T00:00:00Z",
		"stats": {"totalSwaps": 100, "openPositions": 3, "followers": 42},
		"tags": ["whale"],
		"status": "ACTIVE",
		"topTokens": [
			{"symbol": "DEGEN", "realizedProfitUsd": 500.0, "tradeCount": 10}
		]
	}`

	var w Wallet
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if w.ID != 456 {
		t.Errorf("ID = %d, want 456", w.ID)
	}
	if w.WinRate == nil || *w.WinRate != 0.65 {
		t.Errorf("WinRate = %v, want 0.65", w.WinRate)
	}
	if w.Stats.Followers != 42 {
		t.Errorf("Stats.Followers = %d, want 42", w.Stats.Followers)
	}
	if len(w.TopTokens) != 1 || w.TopTokens[0].Symbol != "DEGEN" {
		t.Errorf("TopTokens = %v, want [{DEGEN ...}]", w.TopTokens)
	}
}

func TestUserProfile_Unmarshal(t *testing.T) {
	raw := `{
		"id": "user_123",
		"nickname": "alice",
		"name": null,
		"email": "alice@example.com",
		"createdAt": "2024-06-01T00:00:00Z",
		"isFounder": true,
		"stats": {"strategiesCreated": 3, "walletsFollowed": 5, "strategiesFollowed": 2}
	}`

	var p UserProfile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if p.ID != "user_123" {
		t.Errorf("ID = %q, want %q", p.ID, "user_123")
	}
	if p.Nickname == nil || *p.Nickname != "alice" {
		t.Errorf("Nickname = %v, want %q", p.Nickname, "alice")
	}
	if p.Name != nil {
		t.Errorf("Name = %v, want nil", p.Name)
	}
	if !p.IsFounder {
		t.Error("IsFounder = false, want true")
	}
}

func TestSubscription_Unmarshal(t *testing.T) {
	raw := `{
		"tier": "PRO",
		"status": "active",
		"currentPeriodEnd": "2025-07-01T00:00:00Z",
		"cancelAtPeriodEnd": false,
		"isFounder": false,
		"createdAt": "2024-01-01T00:00:00Z"
	}`

	var s Subscription
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if s.Tier != "PRO" {
		t.Errorf("Tier = %q, want %q", s.Tier, "PRO")
	}
	if s.CurrentPeriodEnd == nil {
		t.Error("CurrentPeriodEnd = nil, want non-nil")
	}
}

func TestHealthStatus_Unmarshal(t *testing.T) {
	raw := `{
		"status": "ok",
		"version": "1.2.3",
		"timestamp": "2025-01-15T12:00:00Z",
		"user": "user_123",
		"rateLimit": {"limit": 100, "keyPrefix": "rms_abc"}
	}`

	var h HealthStatus
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if h.Status != "ok" {
		t.Errorf("Status = %q, want %q", h.Status, "ok")
	}
	if h.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", h.Version, "1.2.3")
	}
	if h.RateLimit.Limit != 100 {
		t.Errorf("RateLimit.Limit = %d, want 100", h.RateLimit.Limit)
	}
}

func TestPaginatedResponse_Unmarshal(t *testing.T) {
	raw := `{
		"data": [{"id": 1, "shareId": "abc", "name": "Test", "description": null, "roiPercent": null, "lastActivityAt": null, "createdAt": "2025-01-01T00:00:00Z", "creator": {"nickname": "a"}, "stats": {"walletsTracked": 0, "totalSwaps": 0}}],
		"pagination": {"page": 1, "pageSize": 50, "totalItems": 1, "totalPages": 1}
	}`

	var resp ListResponse[StrategyListItem]
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	if resp.Pagination.TotalItems != 1 {
		t.Errorf("Pagination.TotalItems = %d, want 1", resp.Pagination.TotalItems)
	}
	if resp.Pagination.PageSize != 50 {
		t.Errorf("Pagination.PageSize = %d, want 50", resp.Pagination.PageSize)
	}
}

// --- Client tests ---

func TestNewClient(t *testing.T) {
	c := NewClient("rms_test_key")

	if c.apiKey != "rms_test_key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "rms_test_key")
	}
	if c.baseURL != "https://www.ramaris.app/api/v1" {
		t.Errorf("baseURL = %q, want default", c.baseURL)
	}
	if c.httpClient == nil {
		t.Error("httpClient = nil, want non-nil")
	}
}

func TestNewClient_WithBaseURL(t *testing.T) {
	c := NewClient("rms_key", WithBaseURL("http://localhost:3000"))

	if c.baseURL != "http://localhost:3000" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:3000")
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient("rms_key", WithHTTPClient(custom))

	if c.httpClient != custom {
		t.Error("httpClient was not set to custom client")
	}
}

func TestClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok","version":"1.0","timestamp":"now","user":"u","rateLimit":{"limit":100,"keyPrefix":"x"}}`)
	}))
	defer srv.Close()

	c := NewClient("rms_secret", WithBaseURL(srv.URL))
	_, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}

	if gotAuth != "Bearer rms_secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer rms_secret")
	}
}

func TestClient_RateLimitHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok","version":"1.0","timestamp":"now","user":"u","rateLimit":{"limit":100,"keyPrefix":"x"}}`)
	}))
	defer srv.Close()

	c := NewClient("rms_key", WithBaseURL(srv.URL))
	_, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}

	rl := c.RateLimit()
	if rl == nil {
		t.Fatal("RateLimit() = nil, want non-nil")
	}
	if rl.Limit != 100 {
		t.Errorf("RateLimit.Limit = %d, want 100", rl.Limit)
	}
	if rl.Remaining != 99 {
		t.Errorf("RateLimit.Remaining = %d, want 99", rl.Remaining)
	}
	if rl.Reset != 1700000000 {
		t.Errorf("RateLimit.Reset = %d, want 1700000000", rl.Reset)
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":{"code":"NOT_FOUND","message":"strategy not found"}}`)
	}))
	defer srv.Close()

	c := NewClient("rms_key", WithBaseURL(srv.URL))
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("Health() error = nil, want error")
	}

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *Error", err)
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "NOT_FOUND")
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestClient_RateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"error":{"code":"RATE_LIMITED","message":"too many requests","retryAfter":60}}`)
	}))
	defer srv.Close()

	c := NewClient("rms_key", WithBaseURL(srv.URL))
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("Health() error = nil, want error")
	}

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("error type = %T, want *RateLimitError", err)
	}
	if rlErr.RetryAfter != 60 {
		t.Errorf("RetryAfter = %d, want 60", rlErr.RetryAfter)
	}
}

func TestClient_Retry5xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, `{"error":{"code":"SERVER_ERROR","message":"bad gateway"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok","version":"1.0","timestamp":"now","user":"u","rateLimit":{"limit":100,"keyPrefix":"x"}}`)
	}))
	defer srv.Close()

	c := NewClient("rms_key", WithBaseURL(srv.URL))
	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if h.Status != "ok" {
		t.Errorf("Status = %q, want %q", h.Status, "ok")
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer srv.Close()

	c := NewClient("rms_key", WithBaseURL(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Health(ctx)
	if err == nil {
		t.Fatal("Health() error = nil, want context error")
	}
}

// --- Endpoint tests ---

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("rms_test", WithBaseURL(srv.URL))
	return srv, c
}

func TestListStrategies(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListOptions
		wantPath string
	}{
		{"nil opts", nil, "/strategies"},
		{"page 2", &ListOptions{Page: 2}, "/strategies?page=2"},
		{"page and size", &ListOptions{Page: 1, PageSize: 10}, "/strategies?page=1&pageSize=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.RequestURI()
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{
					"data": [{"id":1,"shareId":"s1","name":"Alpha","description":null,"roiPercent":null,"lastActivityAt":null,"createdAt":"2025-01-01T00:00:00Z","creator":{"nickname":"a"},"stats":{"walletsTracked":5,"totalSwaps":100}}],
					"pagination": {"page":1,"pageSize":50,"totalItems":1,"totalPages":1}
				}`)
			})

			resp, err := c.ListStrategies(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("ListStrategies() error: %v", err)
			}

			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if len(resp.Data) != 1 {
				t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
			}
			if resp.Data[0].Name != "Alpha" {
				t.Errorf("Data[0].Name = %q, want %q", resp.Data[0].Name, "Alpha")
			}
		})
	}
}

func TestGetStrategy(t *testing.T) {
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":1,"shareId":"abc","name":"My Strategy","description":null,"roiPercent":42.5,"lastActivityAt":null,"createdAt":"2025-01-01T00:00:00Z","creator":{"nickname":"a"},"stats":{"walletsTracked":10,"totalSwaps":50,"totalNotifications":3},"status":"ACTIVE","tags":["defi"]}}`)
	})

	s, err := c.GetStrategy(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetStrategy() error: %v", err)
	}

	if gotPath != "/strategies/abc" {
		t.Errorf("path = %q, want %q", gotPath, "/strategies/abc")
	}
	if s.Name != "My Strategy" {
		t.Errorf("Name = %q, want %q", s.Name, "My Strategy")
	}
	if s.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", s.Status, "ACTIVE")
	}
	if s.Stats.TotalNotifications != 3 {
		t.Errorf("Stats.TotalNotifications = %d, want 3", s.Stats.TotalNotifications)
	}
}

func TestListWatchlist(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/strategies/watchlist" {
			t.Errorf("path = %q, want /me/strategies/watchlist", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"data": [{"id":1,"shareId":"w1","name":"Watched","description":null,"roiPercent":null,"lastActivityAt":null,"creator":{"nickname":"b"},"copiedAt":"2025-02-01T00:00:00Z"}],
			"pagination": {"page":1,"pageSize":50,"totalItems":1,"totalPages":1}
		}`)
	})

	resp, err := c.ListWatchlist(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWatchlist() error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0].Name != "Watched" {
		t.Errorf("Data[0].Name = %q, want %q", resp.Data[0].Name, "Watched")
	}
}

func TestListWallets(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"data": [{"id":1,"winRate":0.75,"realizedPnL":500.0,"createdAt":"2025-01-01T00:00:00Z","stats":{"totalSwaps":50,"openPositions":2},"tags":["whale"]}],
			"pagination": {"page":1,"pageSize":50,"totalItems":1,"totalPages":1}
		}`)
	})

	resp, err := c.ListWallets(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWallets() error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	if resp.Data[0].WinRate == nil || *resp.Data[0].WinRate != 0.75 {
		t.Errorf("Data[0].WinRate = %v, want 0.75", resp.Data[0].WinRate)
	}
}

func TestGetWallet(t *testing.T) {
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":456,"winRate":0.65,"realizedPnL":1000.0,"createdAt":"2025-01-01T00:00:00Z","stats":{"totalSwaps":100,"openPositions":3,"followers":42},"tags":[],"status":"ACTIVE","topTokens":[{"symbol":"DEGEN","realizedProfitUsd":500.0,"tradeCount":10}]}}`)
	})

	w, err := c.GetWallet(context.Background(), 456)
	if err != nil {
		t.Fatalf("GetWallet() error: %v", err)
	}

	if gotPath != "/wallets/456" {
		t.Errorf("path = %q, want %q", gotPath, "/wallets/456")
	}
	if w.ID != 456 {
		t.Errorf("ID = %d, want 456", w.ID)
	}
	if len(w.TopTokens) != 1 {
		t.Fatalf("len(TopTokens) = %d, want 1", len(w.TopTokens))
	}
}

func TestGetProfile(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/profile" {
			t.Errorf("path = %q, want /me/profile", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"u1","nickname":"alice","name":null,"email":"a@b.com","createdAt":"2025-01-01T00:00:00Z","isFounder":true,"stats":{"strategiesCreated":1,"walletsFollowed":2,"strategiesFollowed":3}}}`)
	})

	p, err := c.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}
	if p.Email != "a@b.com" {
		t.Errorf("Email = %q, want %q", p.Email, "a@b.com")
	}
	if !p.IsFounder {
		t.Error("IsFounder = false, want true")
	}
}

func TestGetSubscription(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/subscription" {
			t.Errorf("path = %q, want /me/subscription", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"tier":"PRO","status":"active","currentPeriodEnd":"2025-07-01T00:00:00Z","cancelAtPeriodEnd":false,"isFounder":false,"createdAt":"2024-01-01T00:00:00Z"}}`)
	})

	sub, err := c.GetSubscription(context.Background())
	if err != nil {
		t.Fatalf("GetSubscription() error: %v", err)
	}
	if sub.Tier != "PRO" {
		t.Errorf("Tier = %q, want %q", sub.Tier, "PRO")
	}
}

func TestHealth(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok","version":"2.0.0","timestamp":"2025-06-01T00:00:00Z","user":"test","rateLimit":{"limit":100,"keyPrefix":"rms_"}}`)
	})

	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if h.Status != "ok" {
		t.Errorf("Status = %q, want %q", h.Status, "ok")
	}
	if h.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", h.Version, "2.0.0")
	}
}
