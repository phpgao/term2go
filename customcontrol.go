package term2go

import (
	"context"
	"fmt"
	"regexp"

	iterm2 "github.com/phpgao/term2go/proto"
)

// CustomControlSequenceMonitor watches for custom control sequences matching
// an identity and regex pattern. Corresponds to Python's CustomControlSequenceMonitor.
//
// Usage:
//
//	mon := NewCustomControlSequenceMonitor(conn, "shared-secret", `^open$`, "")
//	mon.Start(ctx, caller)
//	for match := range mon.C {
//	    fmt.Println(match[0])
//	}
//	defer mon.Stop(ctx, caller)
type CustomControlSequenceMonitor struct {
	conn      *Connection
	identity  string
	regex     *regexp.Regexp
	sessionID string
	token     NotificationToken
	C         chan []string // regex match groups
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewCustomControlSequenceMonitor creates a monitor. sessionID can be empty
// to watch all sessions.
func NewCustomControlSequenceMonitor(conn *Connection, identity, pattern, sessionID string) (*CustomControlSequenceMonitor, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile regex: %w", err)
	}
	return &CustomControlSequenceMonitor{
		conn:      conn,
		identity:  identity,
		regex:     re,
		sessionID: sessionID,
		C:         make(chan []string, 16),
	}, nil
}

// Start subscribes to custom escape sequence notifications and begins filtering.
func (m *CustomControlSequenceMonitor) Start(ctx context.Context, caller Caller) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	token, err := SubscribeCustomEscapeSequence(ctx, caller, m.conn,
		func(_ Caller, n *iterm2.CustomEscapeSequenceNotification) {
			if n.GetSenderIdentity() != m.identity {
				return
			}
			matches := m.regex.FindStringSubmatch(n.GetPayload())
			if len(matches) == 0 {
				return
			}
			select {
			case m.C <- matches:
			case <-m.ctx.Done():
			}
		},
		m.sessionID,
	)
	if err != nil {
		return fmt.Errorf("start custom control monitor: %w", err)
	}
	m.token = token
	return nil
}

// Stop unsubscribes from notifications and closes the channel.
func (m *CustomControlSequenceMonitor) Stop(caller Caller) error {
	if m.cancel != nil {
		m.cancel()
	}
	m.conn.Unsubscribe(m.token)
	close(m.C)
	return nil
}
