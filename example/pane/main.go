// example/pane — split pane operations
//
// Demonstrates:
//   - Horizontal split (Vertical=true)
//   - Vertical split (Vertical=false)
//   - Traversing split structure
//
// Usage: go run ./example/pane
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/phpgao/term2go"
)

func main() {
	ctx := context.Background()

	err := term2go.Run(ctx, "pane-example", func(caller term2go.Caller) error {
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
					newSession = s // Use first session
					break
				}
			}
			break
		}
		if newSession == nil {
			return fmt.Errorf("session not found")
		}
		defer exampleWindow.Close(ctx)

		_ = newSession.SetName(ctx, "example-pane")
		_ = newSession.SetBadge(ctx, "SPL")

		fmt.Printf("Initial session: %s\n", newSession.ID)

		// 3. Horizontal split (top/bottom)
		newSession2, err := newSession.SplitPane(ctx, true, false, "Default")
		if err != nil {
			return fmt.Errorf("horizontal split failed: %w", err)
		}
		fmt.Printf("Horizontal split new session: %s\n", newSession2.ID)

		time.Sleep(500 * time.Millisecond)

		// 4. Vertical split (left/right)
		newSession3, err := newSession2.SplitPane(ctx, false, false, "Default")
		if err != nil {
			return fmt.Errorf("vertical split failed: %w", err)
		}
		fmt.Printf("Vertical split new session: %s\n", newSession3.ID)

		time.Sleep(500 * time.Millisecond)

		// 5. Traverse split structure
		app, err = term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		for _, w := range app.Windows {
			if w.ID != windowID {
				continue
			}
			for _, tab := range w.Tabs {
				fmt.Printf("\nWindow %s split structure:\n", w.ID)
				printSplitterTree(tab.Root, 0)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printSplitterTree(splitter *term2go.Splitter, depth int) {
	if splitter == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	dir := "horizontal"
	if splitter.Vertical {
		dir = "vertical"
	}

	fmt.Printf("%s├─ Splitter (%s)\n", indent, dir)

	for i, child := range splitter.Children {
		if child.Session != nil {
			fmt.Printf("%s│  └─ Session: %s\n", indent, child.Session.ID)
		} else if child.Splitter != nil {
			if i == len(splitter.Children)-1 {
				fmt.Printf("%s│  └─\n", indent)
			} else {
				fmt.Printf("%s│  │\n", indent)
			}
			printSplitterTree(child.Splitter, depth+2)
		}
	}
}
