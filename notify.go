package term2go

import (
	"context"
	"fmt"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// NotificationToken identifies a subscription so it can be unsubscribed.
type NotificationToken struct {
	seq int64
	key string // lookup key in the per-connection handler map
	nt  iterm2.NotificationType
	sid string // session ID, set for per-session subscriptions
}

type notifyEntry struct {
	seq     int64
	handler NotificationHandler
}

// Dispatch routes an incoming ServerOriginatedMessage that carries a
// Notification to every matching subscribed handler on this connection.
//
// The caller is responsible for feeding notifications to this function, for
// example by registering a wrapper handler on a Connection:
//
//	conn.RegisterHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
//	    conn.Dispatch(msg)
//	    return true
//	})
func (c *Connection) Dispatch(msg *iterm2.ServerOriginatedMessage) {
	n := msg.GetNotification()
	if n == nil {
		return
	}

	keys := notificationKeys(n)
	if len(keys) == 0 {
		return
	}

	// Copy entries outside the lock to avoid deadlocks if a handler
	// (directly or indirectly) calls back into Connection methods
	// that require notifyMu or other locks.
	c.notifyMu.RLock()
	type keyEntries struct {
		key     string
		entries []notifyEntry
	}
	var toDispatch []keyEntries
	for _, key := range keys {
		if entries, ok := c.notifyMap[key]; ok {
			toDispatch = append(toDispatch, keyEntries{key: key, entries: entries})
		}
	}
	c.notifyMu.RUnlock()

	for _, kd := range toDispatch {
		for _, e := range kd.entries {
			e.handler(msg)
		}
	}
}

// notificationKeys returns every lookup key that should be examined for the
// given incoming notification.  Session-scoped subscriptions produce two keys:
// "<prefix>:<sessionID>" (exact match) and "<prefix>:" (all-sessions wildcard).
func notificationKeys(n *iterm2.Notification) []string {
	var keys []string

	switch {
	case n.GetNewSessionNotification() != nil:
		keys = append(keys, "new_session")

	case n.GetTerminateSessionNotification() != nil:
		keys = append(keys, "terminate_session")

	case n.GetKeystrokeNotification() != nil:
		sid := n.GetKeystrokeNotification().GetSession()
		if sid != "" {
			keys = append(keys, "keystroke:"+sid)
		}
		keys = append(keys, "keystroke:")

	case n.GetScreenUpdateNotification() != nil:
		sid := n.GetScreenUpdateNotification().GetSession()
		if sid != "" {
			keys = append(keys, "screen_update:"+sid)
		}
		keys = append(keys, "screen_update:")

	case n.GetPromptNotification() != nil:
		sid := n.GetPromptNotification().GetSession()
		if sid != "" {
			keys = append(keys, "prompt:"+sid)
		}
		keys = append(keys, "prompt:")

	case n.GetCustomEscapeSequenceNotification() != nil:
		sid := n.GetCustomEscapeSequenceNotification().GetSession()
		if sid != "" {
			keys = append(keys, "custom_escape:"+sid)
		}
		keys = append(keys, "custom_escape:")

	case n.GetLayoutChangedNotification() != nil:
		keys = append(keys, "layout_change")

	case n.GetFocusChangedNotification() != nil:
		keys = append(keys, "focus_change")

	case n.GetServerOriginatedRpcNotification() != nil:
		keys = append(keys, "server_originated_rpc:")

	case n.GetBroadcastDomainsChanged() != nil:
		keys = append(keys, "broadcast_change")

	case n.GetVariableChangedNotification() != nil:
		v := n.GetVariableChangedNotification()
		// Exact match: variable_change:<scope>:<identifier>:<name>:
		keys = append(keys, fmt.Sprintf("variable_change:%d:%s:%s:",
			v.GetScope(), v.GetIdentifier(), v.GetName()))
		// All-identifiers wildcard
		keys = append(keys, fmt.Sprintf("variable_change:%d::%s:",
			v.GetScope(), v.GetName()))

	case n.GetProfileChangedNotification() != nil:
		// Profile notifications always matched by wildcard key.
		keys = append(keys, "profile_change:")
	}

	return keys
}

// Unsubscribe removes a previously registered notification handler and, if it
// was the last handler for its key, sends an unsubscribe RPC to iTerm2.
func (c *Connection) Unsubscribe(token NotificationToken) {
	c.notifyMu.Lock()
	entries := c.notifyMap[token.key]
	idx := -1
	for i, e := range entries {
		if e.seq == token.seq {
			idx = i
			break
		}
	}
	if idx >= 0 {
		c.notifyMap[token.key] = append(entries[:idx], entries[idx+1:]...)
	}
	isLast := len(c.notifyMap[token.key]) == 0
	if isLast {
		delete(c.notifyMap, token.key)
	}
	c.notifyMu.Unlock()

	if isLast && c.IsConnected() {
		// Best-effort unsubscribe RPC — only when still connected.
		// After disconnection, dispatchLoop has exited so RPCs can't complete.
		_ = doUnsubscribe(context.Background(), c.caller(), token.nt, token.sid)
	}
}

func (c *Connection) storeHandler(key string, h NotificationHandler) NotificationToken {
	tk := NotificationToken{seq: atomic.AddInt64(&c.notifySeq, 1), key: key}
	c.notifyMu.Lock()
	c.notifyMap[key] = append(c.notifyMap[key], notifyEntry{seq: tk.seq, handler: h})
	c.notifyMu.Unlock()
	return tk
}

// notifyRPC sends a NotificationRequest.  setArgs is an optional callback
// that the caller can use to set Arguments on the proto message (e.g. to add
// a KeystrokeMonitorRequest or VariableMonitorRequest).  Using a closure
// avoids needing to reference the proto package's unexported oneof interface
// type.
func notifyRPC(ctx context.Context, caller Caller, subscribe bool, nt iterm2.NotificationType,
	sessionID string, setArgs func(nr *iterm2.NotificationRequest),
) error {
	req := newRequest()
	nr := &iterm2.NotificationRequest{
		Subscribe:        proto.Bool(subscribe),
		NotificationType: nt.Enum(),
	}
	if sessionID != "" {
		nr.Session = proto.String(sessionID)
	}
	if setArgs != nil {
		setArgs(nr)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_NotificationRequest{
		NotificationRequest: nr,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

func doSubscribe(ctx context.Context, caller Caller, nt iterm2.NotificationType, sessionID string, setArgs func(nr *iterm2.NotificationRequest)) error {
	return notifyRPC(ctx, caller, true, nt, sessionID, setArgs)
}

func doUnsubscribe(ctx context.Context, caller Caller, nt iterm2.NotificationType, sessionID string) error {
	return notifyRPC(ctx, caller, false, nt, sessionID, nil)
}

// SubscribeNewSession registers a callback that fires when a new iTerm2
// session is created.
func SubscribeNewSession(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.NewSessionNotification),
) (NotificationToken, error) {
	key := "new_session"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_NEW_SESSION, "", nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetNewSessionNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_NEW_SESSION
	return tk, nil
}

// SubscribeTerminateSession registers a callback that fires when an iTerm2
// session terminates.
func SubscribeTerminateSession(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.TerminateSessionNotification),
) (NotificationToken, error) {
	key := "terminate_session"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_TERMINATE_SESSION, "", nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetTerminateSessionNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_TERMINATE_SESSION
	return tk, nil
}

// SubscribeKeystroke registers a callback that fires when a key is pressed in
// sessionID.  Pass sessionID == "" to monitor all sessions.
func SubscribeKeystroke(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.KeystrokeNotification), sessionID string,
) (NotificationToken, error) {
	key := "keystroke:" + sessionID

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_KEYSTROKE, sessionID,
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_KeystrokeMonitorRequest{
				KeystrokeMonitorRequest: &iterm2.KeystrokeMonitorRequest{},
			}
		},
	); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetKeystrokeNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_KEYSTROKE
	tk.sid = sessionID
	return tk, nil
}

// SubscribeScreenUpdate registers a callback that fires when the screen
// contents change for sessionID.  Pass sessionID == "" for all sessions.
func SubscribeScreenUpdate(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.ScreenUpdateNotification), sessionID string,
) (NotificationToken, error) {
	key := "screen_update:" + sessionID

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_SCREEN_UPDATE, sessionID, nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetScreenUpdateNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_SCREEN_UPDATE
	tk.sid = sessionID
	return tk, nil
}

// SubscribePrompt registers a callback that fires when a shell prompt is
// detected in sessionID.
func SubscribePrompt(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.PromptNotification), sessionID string,
) (NotificationToken, error) {
	key := "prompt:" + sessionID

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_PROMPT, sessionID,
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_PromptMonitorRequest{
				PromptMonitorRequest: &iterm2.PromptMonitorRequest{
					Modes: []iterm2.PromptMonitorMode{
						iterm2.PromptMonitorMode_PROMPT,
					},
				},
			}
		},
	); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetPromptNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_PROMPT
	tk.sid = sessionID
	return tk, nil
}

// SubscribeCustomEscapeSequence registers a callback that fires when a custom
// escape sequence (OSC 1337 ; Custom=...) is received in sessionID.
func SubscribeCustomEscapeSequence(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.CustomEscapeSequenceNotification), sessionID string,
) (NotificationToken, error) {
	key := "custom_escape:" + sessionID

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE, sessionID, nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetCustomEscapeSequenceNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE
	tk.sid = sessionID
	return tk, nil
}

// SubscribeVariableChange registers a callback that fires when variableName
// changes in sessionID.  The sessionID is used as both the session filter on
// the notification request and the identifier in the variable monitor.
func SubscribeVariableChange(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.VariableChangedNotification),
	sessionID, variableName string,
) (NotificationToken, error) {
	scope := iterm2.VariableScope_SESSION
	key := fmt.Sprintf("variable_change:%d:%s:%s:", scope, sessionID, variableName)

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_VARIABLE_CHANGE, "",
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_VariableMonitorRequest{
				VariableMonitorRequest: &iterm2.VariableMonitorRequest{
					Name:       proto.String(variableName),
					Scope:      scope.Enum(),
					Identifier: proto.String(sessionID),
				},
			}
		},
	); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetVariableChangedNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_VARIABLE_CHANGE
	// Session is not used for unsubscribe; the VariableMonitorRequest is sent.
	tk.sid = sessionID
	return tk, nil
}

// SubscribeLayoutChange registers a callback that fires when the window/tab
// layout changes.
func SubscribeLayoutChange(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.LayoutChangedNotification),
) (NotificationToken, error) {
	key := "layout_change"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_LAYOUT_CHANGE, "", nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetLayoutChangedNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_LAYOUT_CHANGE
	return tk, nil
}

// SubscribeFocusChange registers a callback that fires when the focused window
// or session changes.
func SubscribeFocusChange(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.FocusChangedNotification),
) (NotificationToken, error) {
	key := "focus_change"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_FOCUS_CHANGE, "", nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetFocusChangedNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_FOCUS_CHANGE
	return tk, nil
}

// SubscribeServerOriginatedRPC registers a callback that fires when iTerm2
// invokes a server-originated RPC.  Use name == "" to match all RPC names.
func SubscribeServerOriginatedRPC(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.ServerOriginatedRPCNotification),
) (NotificationToken, error) {
	key := "server_originated_rpc:"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC, "",
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_RpcRegistrationRequest{
				RpcRegistrationRequest: &iterm2.RPCRegistrationRequest{},
			}
		},
	); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetServerOriginatedRpcNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC
	return tk, nil
}

// SubscribeBroadcastChange registers a callback that fires when the broadcast
// domains change.
func SubscribeBroadcastChange(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.BroadcastDomainsChangedNotification),
) (NotificationToken, error) {
	key := "broadcast_change"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_BROADCAST_CHANGE, "", nil); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetBroadcastDomainsChanged(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_BROADCAST_CHANGE
	return tk, nil
}

// SubscribeProfileChange registers a callback that fires when a profile
// changes.  Pass guid == "" to match all profiles.
func SubscribeProfileChange(ctx context.Context, caller Caller, c *Connection,
	callback func(Caller, *iterm2.ProfileChangedNotification),
) (NotificationToken, error) {
	key := "profile_change:"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_PROFILE_CHANGE, "",
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_ProfileChangeRequest{
				ProfileChangeRequest: &iterm2.ProfileChangeRequest{},
			}
		},
	); err != nil {
		return NotificationToken{}, err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetProfileChangedNotification(); n != nil {
			callback(caller, n)
		}
		return false
	}

	tk := c.storeHandler(key, h)
	tk.nt = iterm2.NotificationType_NOTIFY_ON_PROFILE_CHANGE
	return tk, nil
}
