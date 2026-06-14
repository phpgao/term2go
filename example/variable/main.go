// example/variable — session variable read/write
//
// Demonstrates:
//   - Setting session variables
//   - Reading session variables
//
// Usage: go run ./example/variable
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phpgao/term2go"
)

func main() {
	ctx := context.Background()

	err := term2go.Run(ctx, "variable-example", func(caller term2go.Caller) error {
		// 1. Create new window
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("failed to create window: %w", err)
		}

		windowID := resp.GetWindowId()
		sessionID := resp.GetSessionId()

		// 2. Get app and find session
		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		var newSession *term2go.Session
		var exampleWindow *term2go.Window
		for _, w := range app.Windows {
			if w.ID != windowID {
				continue
			}
			exampleWindow = w
			for _, tab := range w.Tabs {
				for _, s := range tab.Root.Sessions() {
					if s.ID == sessionID {
						newSession = s
						break
					}
				}
			}
			break
		}
		if newSession == nil {
			return fmt.Errorf("session not found")
		}
		defer exampleWindow.Close(ctx)

		// 3. Set session name and badge
		_ = newSession.SetName(ctx, "example-variable")
		_ = newSession.SetBadge(ctx, "VAR")

		// 4. Set variable
		key := "user.example_key"
		value := "hello_from_term2go"
		if err = newSession.SetVariable(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set variable: %w", err)
		}
		fmt.Printf("Set variable: %s = %q\n", key, value)

		// 5. Read variable
		time.Sleep(500 * time.Millisecond) // Wait for variable to take effect
		readValue, err := newSession.GetVariable(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to get variable: %w", err)
		}
		fmt.Printf("Read variable: %s = %q\n", key, readValue)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
