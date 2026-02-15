//go:build integration

package ramaris

import (
	"context"
	"os"
	"testing"
	"time"
)

func integrationClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("RAMARIS_API_KEY")
	if key == "" {
		t.Skip("RAMARIS_API_KEY not set, skipping integration test")
	}
	return NewClient(key)
}

func TestIntegration_Health(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h, err := c.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if h.Status == "" {
		t.Error("Status is empty")
	}
	if h.Version == "" {
		t.Error("Version is empty")
	}
}

func TestIntegration_ListStrategies(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.ListStrategies(ctx, &ListOptions{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("ListStrategies() error: %v", err)
	}
	if resp.Pagination.Page != 1 {
		t.Errorf("Pagination.Page = %d, want 1", resp.Pagination.Page)
	}
	if len(resp.Data) == 0 {
		t.Log("Warning: no strategies returned")
	}
}

func TestIntegration_GetStrategy(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First get a strategy ID from the list
	list, err := c.ListStrategies(ctx, &ListOptions{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("ListStrategies() error: %v", err)
	}
	if len(list.Data) == 0 {
		t.Skip("No strategies available")
	}

	s, err := c.GetStrategy(ctx, list.Data[0].ShareID)
	if err != nil {
		t.Fatalf("GetStrategy() error: %v", err)
	}
	if s.ShareID != list.Data[0].ShareID {
		t.Errorf("ShareID = %q, want %q", s.ShareID, list.Data[0].ShareID)
	}
}

func TestIntegration_ListWatchlist(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.ListWatchlist(ctx, nil)
	if err != nil {
		t.Fatalf("ListWatchlist() error: %v", err)
	}
}

func TestIntegration_ListWallets(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.ListWallets(ctx, &ListOptions{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("ListWallets() error: %v", err)
	}
	if resp.Pagination.Page != 1 {
		t.Errorf("Pagination.Page = %d, want 1", resp.Pagination.Page)
	}
}

func TestIntegration_GetProfile(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p, err := c.GetProfile(ctx)
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}
	if p.Email == "" {
		t.Error("Email is empty")
	}
}

func TestIntegration_GetSubscription(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub, err := c.GetSubscription(ctx)
	if err != nil {
		t.Fatalf("GetSubscription() error: %v", err)
	}
	if sub.Tier == "" {
		t.Error("Tier is empty")
	}
}

func TestIntegration_RateLimitTracked(t *testing.T) {
	c := integrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}

	rl := c.RateLimit()
	if rl == nil {
		t.Log("Warning: rate limit headers not returned by API")
		return
	}
	if rl.Limit <= 0 {
		t.Errorf("RateLimit.Limit = %d, want > 0", rl.Limit)
	}
}
