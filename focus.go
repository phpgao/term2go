package term2go

import (
	"context"
	"fmt"
	"sync"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// WindowStatus
// ============================================================================

// WindowStatus describes the window focus change reason.
type WindowStatus int

const (
	WindowBecameKey   WindowStatus = 0
	WindowIsCurrent   WindowStatus = 1
	WindowResignedKey WindowStatus = 2
)

// ============================================================================
// WindowFocusChange
// ============================================================================

// WindowFocusChange reports a window-level focus change.
type WindowFocusChange struct {
	WindowID string
	Status   WindowStatus
}

// ============================================================================
// FocusUpdate
// ============================================================================

// FocusUpdate is produced by FocusMonitor on each focus change.
// Exactly one field will be non-nil/non-zero.
type FocusUpdate struct {
	// ApplicationActive is set when the app becomes/resigns active.
	// true = application became active; false = resigned active.
	ApplicationActive *bool

	// WindowChanged reports a window focus change.
	WindowChanged *WindowFocusChange

	// SelectedTab is the tab ID that became selected (non-nil when set).
	SelectedTab *string

	// ActiveSession is the session ID that became active (non-nil when set).
	ActiveSession *string
}

func focusUpdateFromProto(n *iterm2.FocusChangedNotification) *FocusUpdate {
	u := &FocusUpdate{}
	switch e := n.Event.(type) {
	case *iterm2.FocusChangedNotification_ApplicationActive:
		v := e.ApplicationActive
		u.ApplicationActive = &v
	case *iterm2.FocusChangedNotification_Window_:
		w := e.Window
		if w == nil {
			return nil
		}
		u.WindowChanged = &WindowFocusChange{
			WindowID: w.GetWindowId(),
			Status:   WindowStatus(w.GetWindowStatus()),
		}
	case *iterm2.FocusChangedNotification_SelectedTab:
		v := e.SelectedTab
		u.SelectedTab = &v
	case *iterm2.FocusChangedNotification_Session:
		v := e.Session
		u.ActiveSession = &v
	default:
		return nil
	}
	return u
}

// ============================================================================
// FocusMonitor
// ============================================================================

// FocusMonitor streams focus-change events.  Create one with
// NewFocusMonitor, iterate over Chan(), and call Close() when finished.
type FocusMonitor struct {
	conn            *Connection
	ch              chan *FocusUpdate
	done            chan struct{}
	token           NotificationToken
	dispatchHandler NotificationHandler
	once            sync.Once
}

// NewFocusMonitor subscribes to focus-change notifications.
//
// Usage:
//
//	fm, err := NewFocusMonitor(conn)
//	if err != nil { ... }
//	defer fm.Close()
//	for u := range fm.Chan() {
//	    if u.ApplicationActive != nil {
//	        fmt.Println("app active:", *u.ApplicationActive)
//	    }
//	    if u.WindowChanged != nil {
//	        fmt.Println("window:", u.WindowChanged.WindowID)
//	    }
//	}
func NewFocusMonitor(conn *Connection) (*FocusMonitor, error) {
	fm := &FocusMonitor{
		conn: conn,
		ch:   make(chan *FocusUpdate, 8),
		done: make(chan struct{}),
	}

	key := "focus_change"

	if err := notifyRPC(context.Background(), conn, true, iterm2.NotificationType_NOTIFY_ON_FOCUS_CHANGE, "", nil); err != nil {
		return nil, fmt.Errorf("subscribe focus: %w", err)
	}

	cb := func(msg *iterm2.ServerOriginatedMessage) bool {
		n := msg.GetNotification().GetFocusChangedNotification()
		if n == nil {
			return false
		}
		u := focusUpdateFromProto(n)
		if u == nil {
			return false
		}
		select {
		case fm.ch <- u:
		case <-fm.done:
		}
		return false
	}
	fm.token = conn.storeHandler(key, cb)
	fm.token.nt = iterm2.NotificationType_NOTIFY_ON_FOCUS_CHANGE

	// Bridge: dispatchLoop needs this to route notifications to notifyMap.
	fm.dispatchHandler = func(msg *iterm2.ServerOriginatedMessage) bool {
		conn.Dispatch(msg)
		return false
	}
	conn.RegisterHandler(fm.dispatchHandler)

	// Auto-close on disconnect.
	conn.OnDisconnect(func() { fm.Close() })

	return fm, nil
}

// Chan returns a receive-only channel of FocusUpdates.
func (fm *FocusMonitor) Chan() <-chan *FocusUpdate { return fm.ch }

// Close stops the monitor and unsubscribes.  Safe to call multiple times.
func (fm *FocusMonitor) Close() {
	fm.once.Do(func() {
		close(fm.done)
		fm.conn.Unsubscribe(fm.token)
		if fm.dispatchHandler != nil {
			fm.conn.UnregisterHandler(fm.dispatchHandler)
		}
		close(fm.ch)
	})
}
