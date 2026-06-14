package term2go

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// KeystrokeEvent
// ============================================================================

// KeystrokeAction describes the type of keyboard event.
type KeystrokeAction int

const (
	KeystrokeKeyDown      KeystrokeAction = 0
	KeystrokeKeyUp        KeystrokeAction = 1
	KeystrokeFlagsChanged KeystrokeAction = 2
)

// KeystrokeEvent wraps a KeystrokeNotification with convenience accessors.
type KeystrokeEvent struct {
	raw *iterm2.KeystrokeNotification
}

func newKeystrokeEvent(raw *iterm2.KeystrokeNotification) *KeystrokeEvent {
	return &KeystrokeEvent{raw: raw}
}

// Characters returns the characters produced by the keystroke.
func (k *KeystrokeEvent) Characters() string { return k.raw.GetCharacters() }

// CharactersIgnoringModifiers returns the characters ignoring modifier keys.
func (k *KeystrokeEvent) CharactersIgnoringModifiers() string {
	return k.raw.GetCharactersIgnoringModifiers()
}

// Modifiers returns the modifier flags.
func (k *KeystrokeEvent) Modifiers() []iterm2.Modifiers { return k.raw.GetModifiers() }

// KeyCode returns the virtual key code.
func (k *KeystrokeEvent) KeyCode() int32 { return k.raw.GetKeyCode() }

// Session returns the session ID where the keystroke occurred.
func (k *KeystrokeEvent) Session() string { return k.raw.GetSession() }

// Action returns the keystroke action (key down / key up / flags changed).
func (k *KeystrokeEvent) Action() KeystrokeAction {
	return KeystrokeAction(k.raw.GetAction())
}

// Raw returns the underlying proto notification.
func (k *KeystrokeEvent) Raw() *iterm2.KeystrokeNotification { return k.raw }

// ============================================================================
// KeystrokeMonitor
// ============================================================================

// KeystrokeMonitor streams keystroke events from a session.
// Pass sessionID == "" to monitor all sessions.
//
// By default, only key-down events are received.  Pass advanced=true to also
// receive key-up and flags-changed events.
type KeystrokeMonitor struct {
	conn            *Connection
	ch              chan *KeystrokeEvent
	done            chan struct{}
	token           NotificationToken
	dispatchHandler NotificationHandler
	once            sync.Once
}

// NewKeystrokeMonitor subscribes to keystroke notifications.
// If advanced is true, key-up and flags-changed events are included.
//
// Usage:
//
//	km, err := NewKeystrokeMonitor(conn, "s1", true)
//	defer km.Close()
//	for ev := range km.Chan() {
//	    fmt.Printf("key: %s mods: %v\n", ev.Characters(), ev.Modifiers())
//	}
func NewKeystrokeMonitor(conn *Connection, sessionID string, advanced bool) (*KeystrokeMonitor, error) {
	km := &KeystrokeMonitor{
		conn: conn,
		ch:   make(chan *KeystrokeEvent, 32),
		done: make(chan struct{}),
	}

	key := "keystroke:" + sessionID
	if err := notifyRPC(context.Background(), conn, true, iterm2.NotificationType_NOTIFY_ON_KEYSTROKE, sessionID,
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_KeystrokeMonitorRequest{
				KeystrokeMonitorRequest: &iterm2.KeystrokeMonitorRequest{
					Advanced: proto.Bool(advanced),
				},
			}
		},
	); err != nil {
		return nil, fmt.Errorf("subscribe keystroke: %w", err)
	}

	cb := func(msg *iterm2.ServerOriginatedMessage) bool {
		n := msg.GetNotification().GetKeystrokeNotification()
		if n == nil {
			return false
		}
		select {
		case km.ch <- newKeystrokeEvent(n):
		case <-km.done:
		}
		return false
	}
	km.token = conn.storeHandler(key, cb)
	km.token.nt = iterm2.NotificationType_NOTIFY_ON_KEYSTROKE
	km.token.sid = sessionID

	km.dispatchHandler = func(msg *iterm2.ServerOriginatedMessage) bool {
		conn.Dispatch(msg)
		return false
	}
	conn.RegisterHandler(km.dispatchHandler)
	conn.OnDisconnect(func() { km.Close() })

	return km, nil
}

// Chan returns a receive-only channel of KeystrokeEvents.
func (km *KeystrokeMonitor) Chan() <-chan *KeystrokeEvent { return km.ch }

// Close stops the monitor.  Safe to call multiple times.
func (km *KeystrokeMonitor) Close() {
	km.once.Do(func() {
		close(km.done)
		km.conn.Unsubscribe(km.token)
		if km.dispatchHandler != nil {
			km.conn.UnregisterHandler(km.dispatchHandler)
		}
		close(km.ch)
	})
}

// ============================================================================
// KeystrokeFilter
// ============================================================================

// KeystrokeFilter tells iTerm2 to intercept keystrokes matching the given
// patterns.  Intercepted keystrokes are not delivered to the terminal but are
// still sent as KeystrokeNotifications — use a KeystrokeMonitor to receive
// them.
//
// The filter is active from creation until Close() is called.
type KeystrokeFilter struct {
	conn  *Connection
	token NotificationToken
	once  sync.Once
}

// NewKeystrokeFilter subscribes the KEYSTROKE_FILTER with the given patterns.
// sessionID may be "" to filter keystrokes in all sessions.
func NewKeystrokeFilter(conn *Connection, sessionID string, patterns []*iterm2.KeystrokePattern) (*KeystrokeFilter, error) {
	kf := &KeystrokeFilter{conn: conn}

	if err := notifyRPC(context.Background(), conn, true, iterm2.NotificationType_KEYSTROKE_FILTER, sessionID,
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_KeystrokeFilterRequest{
				KeystrokeFilterRequest: &iterm2.KeystrokeFilterRequest{
					PatternsToIgnore: patterns,
				},
			}
		},
	); err != nil {
		return nil, fmt.Errorf("subscribe keystroke filter: %w", err)
	}

	// KEYSTROKE_FILTER does not generate notifications — the handler
	// is a no-op.  The filter applies on the iTerm2 side.
	key := "keystroke:" + sessionID
	cb := func(msg *iterm2.ServerOriginatedMessage) bool { return false }
	kf.token = conn.storeHandler(key, cb)
	kf.token.nt = iterm2.NotificationType_KEYSTROKE_FILTER
	kf.token.sid = sessionID

	return kf, nil
}

// Close removes the filter.  Safe to call multiple times.
func (kf *KeystrokeFilter) Close() {
	kf.once.Do(func() {
		kf.conn.Unsubscribe(kf.token)
	})
}
