// example/property — session property operations
//
// Demonstrates:
//   - Setting session name (title)
//   - Setting session badge
//   - Setting buried state
//   - Setting grid size
//
// Usage: go run ./example/property
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

	err := term2go.Run(ctx, "property-example", func(caller term2go.Caller) error {
		// 1. Create new window
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("failed to create window: %w", err)
		}

		windowID := resp.GetWindowId()

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
					newSession = s
					break
				}
			}
			break
		}
		if newSession == nil {
			return fmt.Errorf("session not found")
		}
		defer exampleWindow.Close(ctx)

		// 3. Identify this example's session
		_ = newSession.SetName(ctx, "example-property")
		_ = newSession.SetBadge(ctx, "PROP")

		// 4. Set session name (demonstrates rename)
		if err = newSession.SetName(ctx, "term2go-example"); err != nil {
			return fmt.Errorf("failed to set name: %w", err)
		}
		fmt.Println("Set session name: term2go-example")

		// 5. Set session badge (demonstrates badge change)
		if err = newSession.SetBadge(ctx, "DEMO"); err != nil {
			return fmt.Errorf("failed to set badge: %w", err)
		}
		fmt.Println("Set session badge: DEMO")

		// 6. Set buried state (hide then unhide)
		if err = newSession.SetBuried(ctx, true); err != nil {
			return fmt.Errorf("failed to set buried: %w", err)
		}
		fmt.Println("Set session buried: true")
		time.Sleep(500 * time.Millisecond)

		if err = newSession.SetBuried(ctx, false); err != nil {
			return fmt.Errorf("failed to unset buried: %w", err)
		}
		fmt.Println("Set session buried: false")

		// 7. Set grid size
		if err = newSession.SetGridSize(ctx, 120, 30); err != nil {
			return fmt.Errorf("failed to set grid size: %w", err)
		}
		fmt.Println("Set grid size: 120x30")

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
