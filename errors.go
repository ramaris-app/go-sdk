package ramaris

import "fmt"

// Error represents an API error response from Ramaris.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("ramaris: %s: %s", e.Code, e.Message)
}

// RateLimitError is returned when the API rate limit is exceeded (HTTP 429).
type RateLimitError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	RetryAfter int    `json:"retryAfter"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("ramaris: %s: %s (retry after %ds)", e.Code, e.Message, e.RetryAfter)
}
