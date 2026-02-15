package ramaris

import "time"

// ListOptions configures pagination for list endpoints.
type ListOptions struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
}

// Pagination describes the pagination state of a list response.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// ListResponse is a paginated list of items.
type ListResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// RateLimitInfo holds rate limit state from response headers.
type RateLimitInfo struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}

// StrategyCreator is the creator of a strategy.
type StrategyCreator struct {
	Nickname string `json:"nickname"`
}

// StrategyStats holds aggregate stats for a strategy list item.
type StrategyStats struct {
	WalletsTracked int `json:"walletsTracked"`
	TotalSwaps     int `json:"totalSwaps"`
}

// StrategyListItem is a strategy returned by the list endpoint.
type StrategyListItem struct {
	ID             int             `json:"id"`
	ShareID        string          `json:"shareId"`
	Name           string          `json:"name"`
	Description    *string         `json:"description"`
	ROIPercent     *float64        `json:"roiPercent"`
	LastActivityAt *time.Time      `json:"lastActivityAt"`
	CreatedAt      time.Time       `json:"createdAt"`
	Creator        StrategyCreator `json:"creator"`
	Stats          StrategyStats   `json:"stats"`
}

// StrategyDetailStats extends StrategyStats with notification count.
type StrategyDetailStats struct {
	WalletsTracked     int `json:"walletsTracked"`
	TotalSwaps         int `json:"totalSwaps"`
	TotalNotifications int `json:"totalNotifications"`
}

// Strategy is the full detail of a single strategy.
type Strategy struct {
	ID             int                 `json:"id"`
	ShareID        string              `json:"shareId"`
	Name           string              `json:"name"`
	Description    *string             `json:"description"`
	ROIPercent     *float64            `json:"roiPercent"`
	LastActivityAt *time.Time          `json:"lastActivityAt"`
	CreatedAt      time.Time           `json:"createdAt"`
	Creator        StrategyCreator     `json:"creator"`
	Stats          StrategyDetailStats `json:"stats"`
	Status         string              `json:"status"`
	Tags           []string            `json:"tags"`
}

// WatchlistStrategy is a strategy in the user's watchlist.
type WatchlistStrategy struct {
	ID             int             `json:"id"`
	ShareID        string          `json:"shareId"`
	Name           string          `json:"name"`
	Description    *string         `json:"description"`
	ROIPercent     *float64        `json:"roiPercent"`
	LastActivityAt *time.Time      `json:"lastActivityAt"`
	Creator        StrategyCreator `json:"creator"`
	CopiedAt       time.Time       `json:"copiedAt"`
}

// WalletStats holds aggregate stats for a wallet list item.
type WalletStats struct {
	TotalSwaps    int `json:"totalSwaps"`
	OpenPositions int `json:"openPositions"`
}

// WalletListItem is a wallet returned by the list endpoint.
type WalletListItem struct {
	ID          int         `json:"id"`
	WinRate     *float64    `json:"winRate"`
	RealizedPnL *float64    `json:"realizedPnL"`
	CreatedAt   time.Time   `json:"createdAt"`
	Stats       WalletStats `json:"stats"`
	Tags        []string    `json:"tags"`
}

// WalletDetailStats extends WalletStats with follower count.
type WalletDetailStats struct {
	TotalSwaps    int `json:"totalSwaps"`
	OpenPositions int `json:"openPositions"`
	Followers     int `json:"followers"`
}

// TopToken is a top-performing token in a wallet.
type TopToken struct {
	Symbol            string  `json:"symbol"`
	RealizedProfitUsd float64 `json:"realizedProfitUsd"`
	TradeCount        int     `json:"tradeCount"`
}

// Wallet is the full detail of a single wallet.
type Wallet struct {
	ID          int               `json:"id"`
	WinRate     *float64          `json:"winRate"`
	RealizedPnL *float64          `json:"realizedPnL"`
	CreatedAt   time.Time         `json:"createdAt"`
	Stats       WalletDetailStats `json:"stats"`
	Tags        []string          `json:"tags"`
	Status      string            `json:"status"`
	TopTokens   []TopToken        `json:"topTokens"`
}

// UserProfileStats holds aggregate user stats.
type UserProfileStats struct {
	StrategiesCreated  int `json:"strategiesCreated"`
	WalletsFollowed    int `json:"walletsFollowed"`
	StrategiesFollowed int `json:"strategiesFollowed"`
}

// UserProfile is the authenticated user's profile.
type UserProfile struct {
	ID        string           `json:"id"`
	Nickname  *string          `json:"nickname"`
	Name      *string          `json:"name"`
	Email     string           `json:"email"`
	CreatedAt time.Time        `json:"createdAt"`
	IsFounder bool             `json:"isFounder"`
	Stats     UserProfileStats `json:"stats"`
}

// Subscription is the user's subscription status.
type Subscription struct {
	Tier              string     `json:"tier"`
	Status            string     `json:"status"`
	CurrentPeriodEnd  *time.Time `json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool       `json:"cancelAtPeriodEnd"`
	IsFounder         bool       `json:"isFounder"`
	CreatedAt         *time.Time `json:"createdAt"`
}

// HealthRateLimit describes the rate limit config in the health response.
type HealthRateLimit struct {
	Limit     int    `json:"limit"`
	KeyPrefix string `json:"keyPrefix"`
}

// HealthStatus is the API health check response.
type HealthStatus struct {
	Status    string          `json:"status"`
	Version   string          `json:"version"`
	Timestamp string          `json:"timestamp"`
	User      string          `json:"user"`
	RateLimit HealthRateLimit `json:"rateLimit"`
}

// singleResponse wraps a single resource in a {data: T} envelope.
type singleResponse[T any] struct {
	Data T `json:"data"`
}

// errorResponse is the API error envelope.
type errorResponse struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		RetryAfter int    `json:"retryAfter,omitempty"`
	} `json:"error"`
}
