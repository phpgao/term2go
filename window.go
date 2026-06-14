package term2go

import (
	"context"
	"fmt"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// Window
// ============================================================================

// Window represents an iTerm2 terminal window.
type Window struct {
	caller Caller
	ID     string
	Tabs   []*Tab
	Frame  *WindowFrame
	Number int32
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
