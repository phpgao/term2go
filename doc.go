// Package term2go provides a Go client library for iTerm2's WebSocket API.
//
// It mirrors the official iTerm2 Python API, enabling Go programs to
// control iTerm2 — list windows/tabs/sessions, send text, split panes,
// read terminal content, subscribe to notifications, and more.
//
// Quick start:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "github.com/phpgao/term2go"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    term2go.Run(ctx, "my-app", func(caller term2go.Caller) error {
//	        app, err := term2go.GetApp(caller)
//	        if err != nil {
//	            return err
//	        }
//	        for _, w := range app.Windows {
//	            fmt.Printf("Window: %d tabs\n", len(w.Tabs))
//	        }
//	        return nil
//	    })
//	}
//
// Connection:
//
// The library connects to iTerm2 via WebSocket. It tries the Unix domain
// socket first (~/Library/Application Support/iTerm2/private/socket),
// then falls back to TCP (localhost:1912). Authentication uses the
// ITERM2_COOKIE / ITERM2_KEY environment variables, or obtains them
// automatically via AppleScript.
//
// Object hierarchy:
//
//	App → Window → Tab → Splitter (recursive) → Session
//
//	- App:       top-level container, holds all windows
//	- Window:    an iTerm2 window, holds tabs
//	- Tab:       a tab, holds a Splitter root
//	- Splitter:  a pane-split container (recursive, leaf is Session)
//	- Session:   a terminal session (pane)
//
// RPC:
//
// All 30+ iTerm2 RPC operations are available as package-level functions,
// or through methods on the model objects:
//
//	session.SendText("ls -la\n", false)
//	session.SplitPane(true, false, "Default")
//	name, _ := session.GetVariable("jobName")
//
// Notifications:
//
// Subscribe to iTerm2 events:
//
//	token, _ := term2go.SubscribeNewSession(caller, conn,
//	    func(c Caller, n *iterm2.NewSessionNotification) {
//	        fmt.Println("new session:", n.GetSessionId())
//	    })
//	defer conn.Unsubscribe(token)
//
// Requirements:
//
// The iTerm2 Python API must be enabled in iTerm2's preferences.
package term2go
