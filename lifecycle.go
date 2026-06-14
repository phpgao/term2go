package term2go

import (
	"context"
	"log"
	"sync"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// EachSessionOnce
// ============================================================================

// EachSessionOnce calls fn exactly once for every session — including those
// that already exist and those created in the future. It subscribes to new
// session notifications on the connection so the callback fires automatically
// when a new session appears.
//
// Already-seen session IDs are tracked internally so fn is never called more
// than once for the same session.
//
// Errors returned by fn are logged and do not interrupt processing.
//
// Deprecated: Use EachSessionOnceCtx instead for proper cancellation support.
func EachSessionOnce(conn *Connection, fn func(session *Session) error) {
	EachSessionOnceCtx(context.Background(), conn, fn)
}

// EachSessionOnceCtx is like EachSessionOnce but accepts a context for cancellation.
// Pass ctx.Done() to stop processing new sessions.
func EachSessionOnceCtx(ctx context.Context, conn *Connection, fn func(session *Session) error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var mu sync.Mutex
	seen := make(map[string]bool)

	// Visit all currently-existing sessions.
	app, err := GetApp(ctx, conn.caller())
	if err != nil {
		log.Printf("term2go: EachSessionOnce: GetApp failed: %v", err)
		// Continue anyway — we can still subscribe for future sessions.
	} else {
		for _, w := range app.Windows {
			for _, t := range w.Tabs {
				for _, s := range t.Root.Sessions() {
					mu.Lock()
					if seen[s.ID] {
						mu.Unlock()
						continue
					}
					seen[s.ID] = true
					mu.Unlock()
					if err := fn(s); err != nil {
						log.Printf("term2go: EachSessionOnce: callback error for session %s: %v", s.ID, err)
					}
				}
			}
		}
	}

	// Subscribe to new sessions. The handler fires in the connection's
	// dispatch goroutine, so we must be thread-safe.
	var subErr error
	_, subErr = SubscribeNewSession(ctx, conn, conn, func(caller Caller, n *iterm2.NewSessionNotification) {
		// Check if context has been cancelled
		if ctx.Err() != nil {
			return
		}
		sessionID := n.GetSessionId()
		mu.Lock()
		if seen[sessionID] {
			mu.Unlock()
			return
		}
		seen[sessionID] = true
		mu.Unlock()

		s := &Session{caller: caller, ID: sessionID}
		if err = fn(s); err != nil {
			log.Printf("term2go: EachSessionOnce: callback error for session %s: %v", sessionID, err)
		}
	})
	if subErr != nil {
		log.Printf("term2go: EachSessionOnce: subscribe failed: %v", subErr)
	}
}
