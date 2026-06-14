// example/query — query operations
//
// Demonstrates:
//   - Getting app info (GetApp)
//   - Traversing windows, tabs, sessions
//
// Usage: go run ./example/query
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/phpgao/term2go"
)

func main() {
	ctx := context.Background()

	err := term2go.Run(ctx, "query-example", func(caller term2go.Caller) error {
		// 1. Get app info
		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		fmt.Printf("=== App Info ===\n")
		fmt.Printf("Window count: %d\n", len(app.Windows))

		// 2. Traverse windows, tabs and sessions
		for i, w := range app.Windows {
			fmt.Printf("\n[Window %d] ID=%s\n", i+1, w.ID)

			for j, tab := range w.Tabs {
				fmt.Printf("  [Tab %d] ID=%s\n", j+1, tab.ID)

				sessions := tab.Root.Sessions()
				fmt.Printf("    Session count: %d\n", len(sessions))

				for k, s := range sessions {
					fmt.Printf("      [Session %d] ID=%s\n", k+1, s.ID)
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
