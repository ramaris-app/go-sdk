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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Check API health
	health, err := client.Health(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("API Status: %s (v%s)\n\n", health.Status, health.Version)

	// List strategies
	strategies, err := client.ListStrategies(ctx, &ramaris.ListOptions{Page: 1, PageSize: 5})
	if err != nil {
		log.Fatalf("List strategies failed: %v", err)
	}
	fmt.Printf("Strategies (page %d of %d):\n", strategies.Pagination.Page, strategies.Pagination.TotalPages)
	for _, s := range strategies.Data {
		roi := "n/a"
		if s.ROIPercent != nil {
			roi = fmt.Sprintf("%.1f%%", *s.ROIPercent)
		}
		fmt.Printf("  - %s (ROI: %s, wallets: %d)\n", s.Name, roi, s.Stats.WalletsTracked)
	}

	// Get a single strategy
	if len(strategies.Data) > 0 {
		fmt.Println()
		detail, err := client.GetStrategy(ctx, strategies.Data[0].ShareID)
		if err != nil {
			log.Fatalf("Get strategy failed: %v", err)
		}
		fmt.Printf("Strategy detail: %s [%s] tags=%v\n", detail.Name, detail.Status, detail.Tags)
	}
}
