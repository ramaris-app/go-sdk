# Ramaris Go SDK

Go client library for the [Ramaris](https://www.ramaris.app) API — Base-focused wallet analytics and strategy tracking.

## Install

```bash
go get github.com/ramaris-app/go-sdk
```

Requires Go 1.22+. Zero external dependencies (stdlib only).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    ramaris "github.com/ramaris-app/go-sdk"
)

func main() {
    client := ramaris.NewClient("rms_your_api_key")

    ctx := context.Background()

    // List strategies
    strategies, err := client.ListStrategies(ctx, &ramaris.ListOptions{Page: 1, PageSize: 10})
    if err != nil {
        log.Fatal(err)
    }
    for _, s := range strategies.Data {
        fmt.Printf("%s — %d wallets tracked\n", s.Name, s.Stats.WalletsTracked)
    }
}
```

## Client Options

```go
// Default client
client := ramaris.NewClient("rms_your_api_key")

// Custom base URL (for testing)
client := ramaris.NewClient("rms_key", ramaris.WithBaseURL("http://localhost:3000/api/v1"))

// Custom HTTP client
client := ramaris.NewClient("rms_key", ramaris.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
```

## Endpoints

### Strategies

```go
// List strategies with pagination
resp, err := client.ListStrategies(ctx, &ramaris.ListOptions{Page: 1, PageSize: 50})

// Get a single strategy by share ID
strategy, err := client.GetStrategy(ctx, "shareId")

// List your watchlist
watchlist, err := client.ListWatchlist(ctx, nil) // nil = default pagination
```

### Wallets

```go
// List wallets
resp, err := client.ListWallets(ctx, nil)

// Get a single wallet by ID
wallet, err := client.GetWallet(ctx, 456)
```

### User

```go
// Get your profile
profile, err := client.GetProfile(ctx)

// Get your subscription
sub, err := client.GetSubscription(ctx)
```

### Health

```go
health, err := client.Health(ctx)
```

## Error Handling

```go
import "errors"

strategy, err := client.GetStrategy(ctx, "invalid")
if err != nil {
    var apiErr *ramaris.Error
    if errors.As(err, &apiErr) {
        fmt.Printf("API error: %s (HTTP %d)\n", apiErr.Code, apiErr.StatusCode)
    }

    var rlErr *ramaris.RateLimitError
    if errors.As(err, &rlErr) {
        fmt.Printf("Rate limited, retry after %d seconds\n", rlErr.RetryAfter)
    }
}
```

## Rate Limits

Rate limit info is updated after every successful request:

```go
client.Health(ctx)

if rl := client.RateLimit(); rl != nil {
    fmt.Printf("%d/%d requests remaining (resets at %d)\n", rl.Remaining, rl.Limit, rl.Reset)
}
```

## Retry Behavior

- **5xx errors**: Retried up to 3 times with exponential backoff (500ms, 1s, 2s)
- **429 rate limit**: Returns `*RateLimitError` immediately with `RetryAfter` — caller decides when to retry
- **4xx errors**: Returns `*Error` immediately (no retry)
- All methods respect `context.Context` for cancellation and timeouts

## Testing

```bash
# Unit tests
go test -v -race ./...

# Integration tests (requires API key)
RAMARIS_API_KEY=rms_... go test -tags=integration -v ./...

# Coverage
go test -coverprofile=c.out ./... && go tool cover -func=c.out
```

## License

MIT
