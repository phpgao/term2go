// example/inject — keystroke injection
//
// Demonstrates:
//   - Injecting keystrokes into session
//   - Simulating keyboard input
//
// Usage: go run ./example/inject
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

	err := term2go.Run(ctx, "inject-example", func(caller term2go.Caller) error {
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

		_ = newSession.SetName(ctx, "example-inject")
		_ = newSession.SetBadge(ctx, "INJ")

		// 3. Send command to enter interactive mode
		if err = newSession.SendText(ctx, "cat > /tmp/term2go_inject.txt << 'EOF'\n"); err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}
		time.Sleep(200 * time.Millisecond)

		// 4. Inject text line by line
		lines := []string{
			"This is line 1 from term2go",
			"This is line 2 from term2go",
			"EOF",
		}

		for _, line := range lines {
			if err = newSession.Inject(ctx, []byte(line)); err != nil {
				return fmt.Errorf("failed to inject text: %w", err)
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 5. Press Enter to execute
		if err = newSession.SendText(ctx, "\n"); err != nil {
			return fmt.Errorf("failed to send enter: %w", err)
		}

		time.Sleep(500 * time.Millisecond)

		// 6. Verify result
		if err = newSession.SendText(ctx, "cat /tmp/term2go_inject.txt\n"); err != nil {
			return fmt.Errorf("failed to view result: %w", err)
		}
		time.Sleep(200 * time.Millisecond)

		buffer, err := newSession.GetBuffer(ctx, &iterm2.LineRange{
			ScreenContentsOnly: proto.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to read buffer: %w", err)
		}

		fmt.Println("=== Inject Result ===")
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
