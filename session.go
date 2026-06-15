package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Session represents an iTerm2 session (a single terminal pane).
type Session struct {
	caller Caller
	ID     string
}

// GetID returns the session's unique identifier.
func (s *Session) GetID() string { return s.ID }

// SendText sends text to the session as if typed by the user.
func (s *Session) SendText(ctx context.Context, text string, opts ...SendTextOption) error {
	return SendText(ctx, s.caller, s.ID, text, opts...)
}

// GetBuffer retrieves the contents of the session's buffer.
func (s *Session) GetBuffer(ctx context.Context, lineRange *iterm2.LineRange) (*iterm2.GetBufferResponse, error) {
	return GetBuffer(ctx, s.caller, s.ID, lineRange)
}

// SplitPane splits this session's pane, creating a new session.
func (s *Session) SplitPane(ctx context.Context, vertical bool, before bool, profile string) (*Session, error) {
	resp, err := SplitPane(ctx, s.caller, s.ID, vertical, before, profile)
	if err != nil {
		return nil, fmt.Errorf("split pane: %w", err)
	}
	if ids := resp.GetSessionId(); len(ids) > 0 {
		return &Session{caller: s.caller, ID: ids[0]}, nil
	}
	return nil, fmt.Errorf("split pane: no session ID returned")
}

// SetVariable sets a variable on this session.
func (s *Session) SetVariable(ctx context.Context, name, value string) error {
	return SetVariable(ctx, s.caller, s.ID, name, value)
}

// GetVariable gets the value of a variable from this session.
// iTerm2 encodes all variable values as JSON; this method decodes them back.
func (s *Session) GetVariable(ctx context.Context, name string) (string, error) {
	values, err := GetVariable(ctx, s.caller, s.ID, []string{name})
	if err != nil {
		return "", err
	}
	if len(values) > 0 {
		return jsonDecodeForVariable(values[0]), nil
	}
	return "", nil
}

// Inject injects raw bytes directly into the session's terminal.
func (s *Session) Inject(ctx context.Context, data []byte) error {
	return Inject(ctx, s.caller, []string{s.ID}, data)
}

// Close closes the session.
func (s *Session) Close(ctx context.Context, opts ...CloseOption) error {
	return Close(ctx, s.caller, s.ID, opts...)
}

// SetBuried sets or unsets the buried (minimized) state of a session.
func (s *Session) SetBuried(ctx context.Context, buried bool) error {
	return SetBuried(ctx, s.caller, s.ID, buried)
}

// SetGridSize sets the visible grid size (columns, rows) of a session.
func (s *Session) SetGridSize(ctx context.Context, width, height int32) error {
	return SetGridSize(ctx, s.caller, s.ID, width, height)
}

// SetName sets the session name via the iterm2.set_name RPC.
func (s *Session) SetName(ctx context.Context, name string) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.set_name(name: "%s")`, name)),
	})
	return err
}

// SetBadge sets the session badge text via SetProfileProperty.
func (s *Session) SetBadge(ctx context.Context, text string) error {
	return SetProfileProperty(ctx, s.caller, s.ID, "Badge Text", ensureJSONValue(text))
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------

// LineInfo describes a session's geometry, corresponding to Python's SessionLineInfo.
type LineInfo struct {
	MutableAreaHeight      int // Visible grid rows
	ScrollbackBufferHeight int // History lines
	Overflow               int // Lines lost to overflow
	FirstVisibleLineNumber int // First line on screen, changes on scroll
}

// GetLineInfo fetches the number of lines visible, in history, and overflowed.
// Corresponds to Python's async_get_line_info.
func (s *Session) GetLineInfo(ctx context.Context) (*LineInfo, error) {
	resp, err := GetProperty(ctx, s.caller, s.ID, "number_of_lines")
	if err != nil {
		return nil, fmt.Errorf("get line info: %w", err)
	}
	raw := resp.GetJsonValue()
	if raw == "" {
		return nil, fmt.Errorf("get line info: empty response")
	}
	var m map[string]int
	if err = json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("get line info: %w", err)
	}
	return &LineInfo{
		MutableAreaHeight:      m["grid"],
		ScrollbackBufferHeight: m["history"],
		Overflow:               m["overflow"],
		FirstVisibleLineNumber: m["first_visible"],
	}, nil
}

// GetScreenStreamer creates a ScreenStreamer that watches this session's
// screen updates and streams the contents via a channel.
func (s *Session) GetScreenStreamer() (*ScreenStreamer, error) {
	conn, ok := s.caller.(*Connection)
	if !ok {
		return nil, fmt.Errorf("session caller is not a *Connection")
	}
	return NewScreenStreamer(conn, s.ID)
}

// GetScreenContents returns the contents of the mutable area of the screen
// (the visible portion, excluding scrollback).
func (s *Session) GetScreenContents(ctx context.Context) (*iterm2.GetBufferResponse, error) {
	return GetBuffer(ctx, s.caller, s.ID, &iterm2.LineRange{ScreenContentsOnly: proto.Bool(true)})
}

// GetContents returns the session contents within a given range of lines.
// firstLine should be at least as large as LineInfo.Overflow.
func (s *Session) GetContents(ctx context.Context, firstLine, lineCount int32) (*iterm2.GetBufferResponse, error) {
	return GetBuffer(ctx, s.caller, s.ID, &iterm2.LineRange{
		WindowedCoordRange: &iterm2.WindowedCoordRange{
			CoordRange: &iterm2.CoordRange{
				Start: &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(int64(firstLine))},
				End:   &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(int64(firstLine + lineCount))},
			},
		},
	}, WithIncludeStyles())
}

// SetSelection sets the text selection on this session.
func (s *Session) SetSelection(ctx context.Context, selection *iterm2.Selection) error {
	return SetSelection(ctx, s.caller, s.ID, selection)
}

// GetSelection returns the current text selection.
func (s *Session) GetSelection(ctx context.Context) (*iterm2.SelectionResponse_GetSelectionResponse, error) {
	return GetSelection(ctx, s.caller, s.ID)
}

// MoveToNewTab moves this session from a split pane into a new tab.
// The session must be one of at least two sessions in its current tab.
// window is optional; if nil, the session's current window is used.
// tabIndex is optional; if nil, the tab is placed after the current tab.
func (s *Session) MoveToNewTab(ctx context.Context, window *Window, tabIndex *int) error {
	args := fmt.Sprintf(`iterm2.move_session_to_new_tab(session: "%s"`, s.ID)
	if window != nil {
		args += fmt.Sprintf(`, window_id: "%s"`, window.ID)
	}
	if tabIndex != nil {
		args += fmt.Sprintf(`, tab_index: %d`, *tabIndex)
	}
	args += ")"
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String(args),
		Timeout:    proto.Float64(-1),
	})
	return err
}

// MoveToNewWindow moves this session from a split pane into a new window.
// The session must be one of at least two sessions in its current tab,
// or be in a window with multiple tabs.
func (s *Session) MoveToNewWindow(ctx context.Context) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String(fmt.Sprintf(`iterm2.move_session_to_new_window(session: "%s")`, s.ID)),
		Timeout:    proto.Float64(-1),
	})
	return err
}

// RunCoprocess runs a coprocess in this session, if none is already running.
func (s *Session) RunCoprocess(ctx context.Context, commandLine string) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.run_coprocess(commandLine: "%s")`, commandLine)),
	})
	return err
}

// StopCoprocess stops the currently running coprocess.
func (s *Session) StopCoprocess(ctx context.Context) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String("iterm2.stop_coprocess()"),
	})
	return err
}

// GetCoprocess returns the command line of the currently running coprocess, or empty string.
func (s *Session) GetCoprocess(ctx context.Context) (string, error) {
	resp, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String("iterm2.get_coprocess()"),
	})
	if err != nil {
		return "", err
	}
	r := resp.GetSuccess().GetJsonResult()
	return jsonDecodeForVariable(r), nil
}

// RunTmuxCommand invokes a tmux command and returns its output.
// Returns an error if the session is not a tmux integration session.
func (s *Session) RunTmuxCommand(ctx context.Context, command string) (string, error) {
	resp, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.run_tmux_command(command: "%s")`, command)),
	})
	if err != nil {
		return "", err
	}
	r := resp.GetSuccess().GetJsonResult()
	return jsonDecodeForVariable(r), nil
}

// AddAnnotation adds a text annotation to a range of the session.
func (s *Session) AddAnnotation(ctx context.Context, startX, startY, endX, endY int32, text string) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout: proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(
			`iterm2.add_annotation(startX: %d, startY: %d, endX: %d, endY: %d, text: "%s")`,
			startX, startY, endX, endY, text)),
	})
	return err
}

// LoadURL loads a URL in a browser session.
// The first time a domain is loaded, the user will be prompted for approval.
func (s *Session) LoadURL(ctx context.Context, url string) error {
	_, err := InvokeFunction(ctx, s.caller, &iterm2.InvokeFunctionRequest{
		Context: &iterm2.InvokeFunctionRequest_Method_{
			Method: &iterm2.InvokeFunctionRequest_Method{
				Receiver: proto.String(s.ID),
			},
		},
		Timeout:    proto.Float64(-1),
		Invocation: proto.String(fmt.Sprintf(`iterm2.load_url(url: "%s")`, url)),
	})
	return err
}

// Activate / Restart

// Activate makes this session the active session in its tab.
func (s *Session) Activate(ctx context.Context, selectTab, orderWindowFront bool) error {
	return Activate(ctx, s.caller, s.ID, orderWindowFront, selectTab)
}

// Restart restarts the session.
func (s *Session) Restart(ctx context.Context, onlyIfExited bool) error {
	return RestartSession(ctx, s.caller, s.ID, WithRestartOnlyIfExited(onlyIfExited))
}

// SplitPaneWithCustomizations splits the pane with custom profile properties.
func (s *Session) SplitPaneWithCustomizations(ctx context.Context, vertical, before bool, profile string, props []*iterm2.ProfileProperty) (*Session, error) {
	resp, err := SplitPane(ctx, s.caller, s.ID, vertical, before, profile, WithSplitPaneCustomProfileProperties(props))
	if err != nil {
		return nil, fmt.Errorf("split pane with customizations: %w", err)
	}
	if ids := resp.GetSessionId(); len(ids) > 0 {
		return &Session{caller: s.caller, ID: ids[0]}, nil
	}
	return nil, fmt.Errorf("split pane: no session ID returned")
}

// GetSelectionText extracts the actual text content of the current selection.
func (s *Session) GetSelectionText(ctx context.Context) (string, error) {
	sel, err := s.GetSelection(ctx)
	if err != nil {
		return "", fmt.Errorf("get selection text: %w", err)
	}
	subs := sel.GetSelection().GetSubSelections()
	if len(subs) == 0 {
		return "", nil
	}
	// Use the first sub-selection's coordinates to fetch the content
	first := subs[0]
	wcr := first.GetWindowedCoordRange()
	if wcr == nil {
		return "", nil
	}
	cr := wcr.GetCoordRange()
	if cr == nil {
		return "", nil
	}
	startY := int32(cr.GetStart().GetY())
	endY := int32(cr.GetEnd().GetY())
	if endY < startY {
		return "", nil
	}
	resp, err := s.GetContents(ctx, startY, endY-startY+1)
	if err != nil {
		return "", fmt.Errorf("get selection text: %w", err)
	}
	var text string
	for _, line := range resp.GetContents() {
		text += line.GetText()
	}
	return text, nil
}
