// example/session — basic session operations
//
// Demonstrates:
//   - Connecting to iTerm2
//   - Sending commands
//   - Reading terminal buffer
//
// Usage: go run ./example/session
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

func main() {
	ctx := context.Background()

	err := term2go.Run(ctx, "session-example", func(caller term2go.Caller) error {
		// 1. Create new window and session
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("failed to create window: %w", err)
		}

		windowID := resp.GetWindowId()
		sessionID := resp.GetSessionId()

		// 2. Get app and find the newly created session
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

		// 3. Set session name and badge for identification
		_ = newSession.SetName(ctx, "example-session")
		_ = newSession.SetBadge(ctx, "EX")

		// 4. Send command (the \n acts as pressing Enter)
		cmd := "echo 'Hello from term2go!'; pwd; date\n"
		if err = newSession.SendText(ctx, cmd); err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}

		// 5. Wait for command to complete
		time.Sleep(1 * time.Second)

		// 6. Read screen contents
		buffer, err := newSession.GetBuffer(ctx, &iterm2.LineRange{
			ScreenContentsOnly: proto.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to read buffer: %w", err)
		}

		fmt.Println("=== Screen Contents ===")
		for _, line := range buffer.GetContents() {
			if line.GetText() != "" {
				fmt.Println(line.GetText())
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
