package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// App represents the iTerm2 application. It holds all terminal windows and
// provides the entry point for navigating the session hierarchy.
type App struct {
	caller  Caller
	Windows []*Window

	// BroadcastDomains contains the current broadcast domains.
	// Call RefreshBroadcastDomains() to populate this field.
	BroadcastDomains BroadcastDomains

	// BuriedSessions contains sessions that are buried (minimized).
	BuriedSessions []*Session

	// AppActive indicates whether the app is the active application.
	AppActive bool

	// CurrentTerminalWindowID is the ID of the current terminal window.
	CurrentTerminalWindowID string
}

// GetApp retrieves the full iTerm2 session hierarchy by calling ListSessions
// and constructing the object tree from the response.
func GetApp(ctx context.Context, caller Caller) (*App, error) {
	resp, err := ListSessions(ctx, caller)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return appFromListSessionsResponse(caller, resp), nil
}

func appFromListSessionsResponse(caller Caller, resp *iterm2.ListSessionsResponse) *App {
	windows := make([]*Window, 0, len(resp.GetWindows()))
	for _, pw := range resp.GetWindows() {
		if w := windowFromProto(caller, pw); w != nil {
			windows = append(windows, w)
		}
	}
	var buried []*Session
	for _, bs := range resp.GetBuriedSessions() {
		buried = append(buried, &Session{
			caller: caller,
			ID:     bs.GetUniqueIdentifier(),
		})
	}
	return &App{
		caller:         caller,
		Windows:        windows,
		BuriedSessions: buried,
	}
}

// Refresh reloads the full window/tab/session hierarchy from iTerm2.
func (a *App) Refresh(ctx context.Context) error {
	resp, err := ListSessions(ctx, a.caller)
	if err != nil {
		return fmt.Errorf("refresh: %w", err)
	}
	newApp := appFromListSessionsResponse(a.caller, resp)
	a.Windows = newApp.Windows
	a.BuriedSessions = newApp.BuriedSessions
	return nil
}

// AppActivateOption controls how the app is activated.
type AppActivateOption func(*iterm2.ActivateRequest_App)

// WithAppActivateRaiseAllWindows raises all windows when activating the app.
func WithAppActivateRaiseAllWindows() AppActivateOption {
	return func(app *iterm2.ActivateRequest_App) {
		app.RaiseAllWindows = proto.Bool(true)
	}
}

// WithAppActivateIgnoringOtherApps activates even if the user interacts with
// another app after the call.
func WithAppActivateIgnoringOtherApps() AppActivateOption {
	return func(app *iterm2.ActivateRequest_App) {
		app.IgnoringOtherApps = proto.Bool(true)
	}
}

// Activate brings the app to the front and gives it keyboard focus.
func (a *App) Activate(ctx context.Context, opts ...AppActivateOption) error {
	appOpt := &iterm2.ActivateRequest_App{}
	for _, o := range opts {
		o(appOpt)
	}
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ActivateRequest{
		ActivateRequest: &iterm2.ActivateRequest{
			Identifier:       nil, // app-level activation
			OrderWindowFront: proto.Bool(false),
			SelectTab:        proto.Bool(false),
			SelectSession:    proto.Bool(false),
			ActivateApp:      appOpt,
		},
	}
	resp, err := a.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("activate app: %w", err)
	}
	return checkError(resp)
}

// GetSessionByID finds a session by its ID, including buried sessions.
func (a *App) GetSessionByID(sessionID string) *Session {
	for _, w := range a.Windows {
		for _, t := range w.Tabs {
			for _, s := range t.Root.Sessions() {
				if s.ID == sessionID {
					return s
				}
			}
		}
	}
	for _, s := range a.BuriedSessions {
		if s.ID == sessionID {
			return s
		}
	}
	return nil
}

// GetTabByID finds a tab by its ID.
func (a *App) GetTabByID(tabID string) *Tab {
	for _, w := range a.Windows {
		for _, t := range w.Tabs {
			if t.ID == tabID {
				return t
			}
		}
	}
	return nil
}

// GetWindowByID finds a window by its ID.
func (a *App) GetWindowByID(windowID string) *Window {
	for _, w := range a.Windows {
		if w.ID == windowID {
			return w
		}
	}
	return nil
}

// GetWindowForTab finds the window that contains the given tab.
func (a *App) GetWindowForTab(tabID string) *Window {
	for _, w := range a.Windows {
		for _, t := range w.Tabs {
			if t.ID == tabID {
				return w
			}
		}
	}
	return nil
}

// GetWindowAndTabForSession finds the window and tab that contain a session.
func (a *App) GetWindowAndTabForSession(session *Session) (*Window, *Tab) {
	for _, w := range a.Windows {
		for _, t := range w.Tabs {
			for _, s := range t.Root.Sessions() {
				if s.ID == session.ID {
					return w, t
				}
			}
		}
	}
	return nil, nil
}

// CurrentWindow returns the current terminal window.
func (a *App) CurrentWindow() *Window {
	if a.CurrentTerminalWindowID == "" {
		return nil
	}
	return a.GetWindowByID(a.CurrentTerminalWindowID)
}

// RefreshFocus updates the current focus state.
func (a *App) RefreshFocus(ctx context.Context) error {
	resp, err := FocusRequest(ctx, a.caller)
	if err != nil {
		return fmt.Errorf("focus request: %w", err)
	}
	for _, n := range resp.GetNotifications() {
		a.applyFocusNotification(n)
	}
	return nil
}

func (a *App) applyFocusNotification(n *iterm2.FocusChangedNotification) {
	if n.GetApplicationActive() {
		a.AppActive = true
		return
	}
	if w := n.GetWindow(); w != nil {
		if w.GetWindowStatus() != iterm2.FocusChangedNotification_Window_TERMINAL_WINDOW_RESIGNED_KEY {
			a.CurrentTerminalWindowID = w.GetWindowId()
		}
	}
}

// GetTheme returns the current theme attributes (e.g., ["dark"] or ["light"]).
func (a *App) GetTheme(ctx context.Context) ([]string, error) {
	raw, err := GetVariable(ctx, a.caller, "", []string{"effectiveTheme"})
	if err != nil {
		return nil, fmt.Errorf("get theme: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	value := jsonDecodeForVariable(raw[0])
	if value == "" {
		return nil, nil
	}
	// The theme value is a space-separated string like "dark" or "light highContrast"
	parts := []string{}
	current := ""
	for _, ch := range value {
		if ch == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts, nil
}

// MoveSession moves a session to be a split pane of another session.
func (a *App) MoveSession(ctx context.Context, session, destination *Session, splitVertical, before bool) error {
	_, err := InvokeFunction(ctx, a.caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String(fmt.Sprintf(
			`iterm2.move_session(session: "%s", destination: "%s", vertical: %t, before: %t)`,
			session.ID, destination.ID, splitVertical, before)),
		Timeout: proto.Float64(-1),
	})
	return err
}

// ApplyLayout applies a target layout to one or more tabs.
// The spec is a JSON string describing the target state. See the Python
// API documentation for async_apply_layout for the full schema.
//
// The spec is base64-encoded (matching the Python implementation) because
// iTerm2's expression parser does not decode \" inside string literals.
func (a *App) ApplyLayout(ctx context.Context, specJSON string) error {
	specB64 := base64Encode(specJSON)
	_, err := InvokeFunction(ctx, a.caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String(fmt.Sprintf(
			`iterm2.apply_layout(spec_json_b64: "%s")`, specB64)),
		Timeout: proto.Float64(-1),
	})
	return err
}

// base64Encode is a simple base64 encoder that avoids importing encoding/base64
// just for this one function.
func base64Encode(s string) string {
	const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	bytes := []byte(s)
	var result []byte
	for i := 0; i < len(bytes); i += 3 {
		b0 := bytes[i]
		b1, b2 := byte(0), byte(0)
		pad1, pad2 := byte('='), byte('=')
		if i+1 < len(bytes) {
			b1 = bytes[i+1]
			pad1 = 0
		}
		if i+2 < len(bytes) {
			b2 = bytes[i+2]
			pad2 = 0
		}
		result = append(result,
			b64[b0>>2],
			b64[((b0&0x03)<<4)|(b1>>4)],
		)
		if pad1 == 0 {
			result = append(result, b64[((b1&0x0f)<<2)|(b2>>6)])
		} else {
			result = append(result, '=')
		}
		if pad2 == 0 {
			result = append(result, b64[b2&0x3f])
		} else {
			result = append(result, '=')
		}
	}
	return string(result)
}

// wait is used internally by App to poll for state refreshes.
