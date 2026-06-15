package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// NavigationDirection specifies a direction for selecting a pane.
type NavigationDirection string

const (
	DirectionLeft  NavigationDirection = "left"
	DirectionRight NavigationDirection = "right"
	DirectionAbove NavigationDirection = "above"
	DirectionBelow NavigationDirection = "below"
)

// Tab represents an iTerm2 tab, which contains a tree of split panes.
type Tab struct {
	caller Caller
	ID     string
	Root   *Splitter
	// ActiveSessionID is the ID of the currently focused session in this tab.
	ActiveSessionID string
	// TmuxWindowID is set when this tab belongs to a tmux integration window.
	TmuxWindowID string
	// TmuxConnectionID is set when this tab belongs to a tmux integration window.
	TmuxConnectionID string
	// MinimizedSessions contains sessions in this tab that are minimized
	// (another session is maximized).
	MinimizedSessions []*Session
}

func tabFromProto(caller Caller, pt *iterm2.ListSessionsResponse_Tab) *Tab {
	var minimized []*Session
	for _, ms := range pt.GetMinimizedSessions() {
		minimized = append(minimized, &Session{caller: caller, ID: ms.GetUniqueIdentifier()})
	}
	return &Tab{
		caller:            caller,
		ID:                pt.GetTabId(),
		Root:              SplitterFromProto(pt.GetRoot(), caller),
		TmuxWindowID:      pt.GetTmuxWindowId(),
		TmuxConnectionID:  pt.GetTmuxConnectionId(),
		MinimizedSessions: minimized,
	}
}

// AllSessions returns both visible and minimized sessions in this tab.
func (t *Tab) AllSessions() []*Session {
	return append(t.Root.Sessions(), t.MinimizedSessions...)
}

// Select makes this tab the active tab.
func (t *Tab) Select(ctx context.Context) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ActivateRequest{
		ActivateRequest: &iterm2.ActivateRequest{
			Identifier: &iterm2.ActivateRequest_TabId{
				TabId: t.ID,
			},
			SelectTab:        proto.Bool(true),
			OrderWindowFront: proto.Bool(true),
		},
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("select tab: %w", err)
	}
	return checkError(resp)
}

// Close closes the tab.
func (t *Tab) Close(ctx context.Context, opts ...CloseOption) error {
	req := newRequest()
	closeReq := &iterm2.CloseRequest{
		Target: &iterm2.CloseRequest_Tabs{
			Tabs: &iterm2.CloseRequest_CloseTabs{
				TabIds: []string{t.ID},
			},
		},
	}
	for _, o := range opts {
		o(closeReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_CloseRequest{
		CloseRequest: closeReq,
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("close tab: %w", err)
	}
	return checkError(resp)
}

// UpdateLayout sends the current split-pane layout to iTerm2 to adjust sizes.
func (t *Tab) UpdateLayout(ctx context.Context) error {
	return SetTabLayout(ctx, t.caller, t.ID, t.Root.ToProto())
}

// SelectPaneInDirection activates the split pane adjacent to the currently
// selected pane in the given direction.
func (t *Tab) SelectPaneInDirection(ctx context.Context, dir NavigationDirection) error {
	_, err := InvokeFunction(ctx, t.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(t.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.select_pane_in_direction(direction: "%s")`, dir)),
	})
	return err
}

// SetTitle changes the tab's title.
func (t *Tab) SetTitle(ctx context.Context, title string) error {
	_, err := InvokeFunction(ctx, t.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(t.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.set_title(title: "%s")`, title)),
	})
	return err
}

// MoveToWindow moves this tab to its own window.
func (t *Tab) MoveToWindow(ctx context.Context) (*Window, error) {
	resp, err := InvokeFunction(ctx, t.caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String("iterm2.move_tab_to_window()"),
		Timeout:    proto.Float64(-1),
	})
	if err != nil {
		return nil, fmt.Errorf("move tab to window: %w", err)
	}
	// The response is the new window ID as a JSON string.
	windowID := resp.GetSuccess().GetJsonResult()
	// Decode the JSON result (it's a string in JSON, e.g. `"window-id"`).
	var id string
	// try plain string first
	id = windowID
	// handle JSON-quoted string
	if len(id) >= 2 && id[0] == '"' && id[len(id)-1] == '"' {
		id = id[1 : len(id)-1]
	}

	app, err := GetApp(ctx, t.caller)
	if err != nil {
		return nil, fmt.Errorf("find new window: %w", err)
	}
	if err := app.Refresh(ctx); err != nil {
		return nil, fmt.Errorf("refresh for new window: %w", err)
	}
	w := app.GetWindowByID(id)
	if w == nil {
		return nil, fmt.Errorf("move tab to window: window %q not found", id)
	}
	return w, nil
}

// CurrentSession returns the active session in this tab, or nil if unknown.
func (t *Tab) CurrentSession() *Session {
	if t.ActiveSessionID == "" {
		return nil
	}
	for _, s := range t.Root.Sessions() {
		if s.ID == t.ActiveSessionID {
			return s
		}
	}
	return nil
}

// Variable (tab scope)

// GetVariable fetches a tab-level variable.
func (t *Tab) GetVariable(ctx context.Context, name string) (string, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_TabId{
				TabId: t.ID,
			},
			Get: []string{name},
		},
	}
	resp, err := t.caller.Call(ctx, msg)
	if err != nil {
		return "", err
	}
	if err = checkError(resp); err != nil {
		return "", err
	}
	values := resp.GetVariableResponse().GetValues()
	if len(values) > 0 {
		return jsonDecodeForVariable(values[0]), nil
	}
	return "", nil
}

// SetVariable sets a tab-level variable.
func (t *Tab) SetVariable(ctx context.Context, name, value string) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_TabId{
				TabId: t.ID,
			},
			Set: []*iterm2.VariableRequest_Set{
				{Name: proto.String(name), Value: proto.String(ensureJSONValue(value))},
			},
		},
	}
	resp, err := t.caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}
