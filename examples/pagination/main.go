package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ramaris "github.com/ramaris-app/go-sdk"
)

func main() {
	key := os.Getenv("RAMARIS_API_KEY")
	if key == "" {
		log.Fatal("RAMARIS_API_KEY environment variable is required")
	}

	client := ramaris.NewClient(key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	page := 1
	total := 0

	for {
		resp, err := client.ListStrategies(ctx, &ramaris.ListOptions{Page: page, PageSize: 50})
		if err != nil {
			log.Fatalf("List strategies page %d failed: %v", page, err)
		}

		for _, s := range resp.Data {
			total++
			fmt.Printf("%d. %s (%s)\n", total, s.Name, s.ShareID)
		}

		if page >= resp.Pagination.TotalPages {
			break
		}
		page++
	}

	fmt.Printf("\nTotal strategies: %d\n", total)

	// Show rate limit info
	if rl := client.RateLimit(); rl != nil {
		fmt.Printf("Rate limit: %d/%d remaining (resets at %d)\n", rl.Remaining, rl.Limit, rl.Reset)
	}
}
