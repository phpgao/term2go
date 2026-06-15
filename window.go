package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Window represents an iTerm2 terminal window.
type Window struct {
	caller Caller
	ID     string
	Tabs   []*Tab
	Frame  *WindowFrame
	Number int32
	// SelectedTabID is the ID of the currently selected tab in this window.
	// It may be empty if unknown.
	SelectedTabID string
}

func windowFromProto(caller Caller, pw *iterm2.ListSessionsResponse_Window) *Window {
	tabs := make([]*Tab, 0, len(pw.GetTabs()))
	for _, pt := range pw.GetTabs() {
		if t := tabFromProto(caller, pt); t != nil {
			tabs = append(tabs, t)
		}
	}
	w := &Window{
		caller: caller,
		ID:     pw.GetWindowId(),
		Tabs:   tabs,
		Number: pw.GetNumber(),
	}
	if f := pw.GetFrame(); f != nil {
		w.Frame = &WindowFrame{
			Origin: Point{X: f.GetOrigin().GetX(), Y: f.GetOrigin().GetY()},
			Size:   Size{Width: f.GetSize().GetWidth(), Height: f.GetSize().GetHeight()},
		}
	}
	return w
}

// CreateTab creates a new tab in this window with the given profile name.
// After creation, it refreshes the window hierarchy to discover the new tab's
// real identifier (CreateTabResponse returns only a tab index, not the UUID
// that Tab.Close needs).
func (w *Window) CreateTab(ctx context.Context, profileName string, opts ...CreateTabOption) (*Tab, error) {
	resp, err := CreateTab(ctx, w.caller, w.ID, profileName, opts...)
	if err != nil {
		return nil, fmt.Errorf("create tab: %w", err)
	}
	// Refresh the app to find the new tab by its session ID.
	app, err := GetApp(ctx, w.caller)
	if err != nil {
		return nil, fmt.Errorf("refresh after create tab: %w", err)
	}
	sessionID := resp.GetSessionId()
	for _, rw := range app.Windows {
		if rw.ID != w.ID {
			continue
		}
		for _, t := range rw.Tabs {
			for _, s := range t.Root.Sessions() {
				if s.ID == sessionID {
					t.caller = w.caller
					return t, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("create tab: created tab not found in refreshed hierarchy")
}

// Close closes the window.
func (w *Window) Close(ctx context.Context, opts ...CloseOption) error {
	req := newRequest()
	closeReq := &iterm2.CloseRequest{
		Target: &iterm2.CloseRequest_Windows{
			Windows: &iterm2.CloseRequest_CloseWindows{
				WindowIds: []string{w.ID},
			},
		},
	}
	for _, o := range opts {
		o(closeReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_CloseRequest{
		CloseRequest: closeReq,
	}
	resp, err := w.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("close window: %w", err)
	}
	return checkError(resp)
}

// Activate gives the window keyboard focus and orders it to the front.
// Note: this does NOT activate the app itself. Call App.Activate() for that.
func (w *Window) Activate(ctx context.Context) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ActivateRequest{
		ActivateRequest: &iterm2.ActivateRequest{
			Identifier: &iterm2.ActivateRequest_WindowId{
				WindowId: w.ID,
			},
			OrderWindowFront: proto.Bool(true),
		},
	}
	resp, err := w.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("activate window: %w", err)
	}
	return checkError(resp)
}

// SetTabs reorders tabs in this window. Tabs not in the list are left in their
// existing positions after the explicitly listed tabs.
func (w *Window) SetTabs(ctx context.Context, tabs []*Tab) error {
	tabIDs := make([]string, len(tabs))
	for i, t := range tabs {
		tabIDs[i] = t.ID
	}
	return ReorderTabs(ctx, w.caller, []*iterm2.ReorderTabsRequest_Assignment{
		{
			WindowId: proto.String(w.ID),
			TabIds:   tabIDs,
		},
	})
}

// GetFullscreen returns whether the window is fullscreen.
func (w *Window) GetFullscreen(ctx context.Context) (bool, error) {
	resp, err := GetProperty(ctx, w.caller, w.ID, "fullscreen")
	if err != nil {
		return false, fmt.Errorf("get fullscreen: %w", err)
	}
	var fullscreen bool
	if err := json.Unmarshal([]byte(resp.GetJsonValue()), &fullscreen); err != nil {
		return false, fmt.Errorf("parse fullscreen: %w", err)
	}
	return fullscreen, nil
}

// SetFullscreen changes the window's fullscreen state.
func (w *Window) SetFullscreen(ctx context.Context, fullscreen bool) error {
	v, _ := json.Marshal(fullscreen)
	return SetProperty(ctx, w.caller, w.ID, "fullscreen", string(v))
}

// Frame (position & size)

// GetFrame fetches the window's current frame from iTerm2.
func (w *Window) GetFrame(ctx context.Context) (*WindowFrame, error) {
	resp, err := GetProperty(ctx, w.caller, w.ID, "frame")
	if err != nil {
		return nil, fmt.Errorf("get frame: %w", err)
	}
	var frame struct {
		Origin struct {
			X int32 `json:"x"`
			Y int32 `json:"y"`
		} `json:"origin"`
		Size struct {
			Width  int32 `json:"width"`
			Height int32 `json:"height"`
		} `json:"size"`
	}
	if err := json.Unmarshal([]byte(resp.GetJsonValue()), &frame); err != nil {
		return nil, fmt.Errorf("parse frame: %w", err)
	}
	return &WindowFrame{
		Origin: Point{X: frame.Origin.X, Y: frame.Origin.Y},
		Size:   Size{Width: frame.Size.Width, Height: frame.Size.Height},
	}, nil
}

// SetFrame sets the window's frame.
func (w *Window) SetFrame(ctx context.Context, f *WindowFrame) error {
	v, _ := json.Marshal(map[string]interface{}{
		"origin": map[string]int32{"x": f.Origin.X, "y": f.Origin.Y},
		"size":   map[string]int32{"width": f.Size.Width, "height": f.Size.Height},
	})
	return SetProperty(ctx, w.caller, w.ID, "frame", string(v))
}

// SetTitle changes the window's title.
func (w *Window) SetTitle(ctx context.Context, title string) error {
	_, err := InvokeFunction(ctx, w.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(w.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.set_title(title: "%s")`, title)),
	})
	return err
}

// SaveArrangement saves the current window as a named arrangement.
func (w *Window) SaveArrangement(ctx context.Context, name string) error {
	_, err := SavedArrangementRequest(ctx, w.caller, &iterm2.SavedArrangementRequest{
		Name:     proto.String(name),
		Action:   iterm2.SavedArrangementRequest_SAVE.Enum(),
		WindowId: proto.String(w.ID),
	})
	return err
}

// RestoreArrangement restores a named arrangement as tabs in this window.
func (w *Window) RestoreArrangement(ctx context.Context, name string) error {
	_, err := SavedArrangementRequest(ctx, w.caller, &iterm2.SavedArrangementRequest{
		Name:     proto.String(name),
		Action:   iterm2.SavedArrangementRequest_RESTORE.Enum(),
		WindowId: proto.String(w.ID),
	})
	return err
}

// Variable (window scope)

// GetVariable fetches a window-level variable.
func (w *Window) GetVariable(ctx context.Context, name string) (string, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_WindowId{
				WindowId: w.ID,
			},
			Get: []string{name},
		},
	}
	resp, err := w.caller.Call(ctx, msg)
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

// SetVariable sets a window-level variable.
func (w *Window) SetVariable(ctx context.Context, name, value string) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_WindowId{
				WindowId: w.ID,
			},
			Set: []*iterm2.VariableRequest_Set{
				{Name: proto.String(name), Value: proto.String(ensureJSONValue(value))},
			},
		},
	}
	resp, err := w.caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// CurrentTab returns the currently selected tab in this window, or nil.
func (w *Window) CurrentTab() *Tab {
	if w.SelectedTabID == "" {
		return nil
	}
	for _, t := range w.Tabs {
		if t.ID == w.SelectedTabID {
			return t
		}
	}
	return nil
}

// Create opens a new window with the given profile and optional command.
// profile may be "" for the default profile.
// command may be "" to use the profile's default command.
func Create(ctx context.Context, caller Caller, profile, command string) (*Window, error) {
	resp, err := CreateTab(ctx, caller, "", profile)
	if err != nil {
		return nil, fmt.Errorf("create window: %w", err)
	}
	app, err := GetApp(ctx, caller)
	if err != nil {
		return nil, fmt.Errorf("find new window: %w", err)
	}
	w := app.GetWindowByID(resp.GetWindowId())
	if w == nil {
		return nil, fmt.Errorf("create window: new window %q not found", resp.GetWindowId())
	}
	return w, nil
}
