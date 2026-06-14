// example/prompt — custom prompt
//
// Demonstrates:
//   - Sending text to terminal
//   - Waiting for user input
//
// Usage: go run ./example/prompt
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

	err := term2go.Run(ctx, "prompt-example", func(caller term2go.Caller) error {
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

		_ = newSession.SetName(ctx, "example-prompt")
		_ = newSession.SetBadge(ctx, "PROMPT")

		fmt.Println("Please answer the prompt question in terminal...")
		fmt.Println()

		// 3. Send prompt text
		prompt := "What is your name? "
		if err = newSession.SendText(ctx, prompt); err != nil {
			return fmt.Errorf("failed to send prompt: %w", err)
		}

		// 4. Wait for user input (simple wait, in real app should listen for input)
		time.Sleep(5 * time.Second)

		// 5. Read current content
		buffer, err := newSession.GetBuffer(ctx, &iterm2.LineRange{
			ScreenContentsOnly: proto.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to read buffer: %w", err)
		}

		fmt.Printf("\n=== Terminal Content ===\n")
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
