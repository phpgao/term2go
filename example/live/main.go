// example/live — live output from continuously running commands
//
// Demonstrates:
//   - Polling GetBuffer in a loop to capture incremental output
//   - Using ScreenContentsOnly for predictable line counts
//   - Dedup via map to avoid re-printing on buffer shifts
//
// Usage: go run ./example/live
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := term2go.Run(ctx, "live-example", func(caller term2go.Caller) error {
		// 1. Create new window
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("failed to create window: %w", err)
		}

		windowID := resp.GetWindowId()
		sessionID := resp.GetSessionId()

		// 2. Wait for shell to initialize
		time.Sleep(1 * time.Second)

		// 3. Find session
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
			return fmt.Errorf("session not found: %s", sessionID)
		}
		defer exampleWindow.Close(ctx)

		_ = newSession.SetName(ctx, "example-live")
		_ = newSession.SetBadge(ctx, "LIVE")

		// 4. Send command
		cmd := "for i in 1 2 3 4 5; do echo \"Count: $i\"; sleep 1; done\n"
		fmt.Printf("Running: %s\n", cmd)

		if err = newSession.SendText(ctx, cmd); err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}

		// 5. Poll buffer
		fmt.Println("=== Live Output ===")
		printed := make(map[string]bool)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nDone")
				return nil
			default:
			}

			time.Sleep(500 * time.Millisecond)

			buffer, err := newSession.GetBuffer(ctx, &iterm2.LineRange{
				ScreenContentsOnly: proto.Bool(true),
			})
			if err != nil {
				continue
			}

			for _, line := range buffer.GetContents() {
				text := strings.TrimRight(line.GetText(), " \t")
				if text == "" {
					continue
				}
				if !printed[text] {
					fmt.Println(text)
					printed[text] = true
				}
			}
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
