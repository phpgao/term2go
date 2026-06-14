// example/notification — event subscriptions
//
// Demonstrates:
//   - Subscribing to new session events
//   - Subscribing to keystroke events
//
// Usage: go run ./example/notification
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phpgao/term2go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := term2go.Run(ctx, "notification-example", func(caller term2go.Caller) error {
		fmt.Println("Waiting for events (10 seconds)...")
		fmt.Println("Please create a new session or press keys in terminal")
		fmt.Println()

		// Just wait and show that we're connected
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		done := ctx.Done()
		for {
			select {
			case <-done:
				fmt.Println("\nTimeout, exiting")
				return nil
			case <-ticker.C:
				fmt.Print(".")
			}
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
