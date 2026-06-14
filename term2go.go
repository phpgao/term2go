// Package term2go provides a Go client library for iTerm2's Python API.
// It connects to the iTerm2 WebSocket interface, implements the RPC protocol,
// and exposes the full session hierarchy (App → Window → Tab → Splitter → Session).
package term2go

import "context"

// Run connects to iTerm2, executes fn with the connection as a Caller,
// and closes the connection when fn returns.
// scriptName identifies this program in iTerm2's scripting console.
func Run(ctx context.Context, scriptName string, fn func(caller Caller) error) error {
	conn, err := Connect(ctx, scriptName)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	return fn(conn)
}
