package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// Session
// ============================================================================

// ============================================================================
// Session
// ============================================================================

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
// LineInfo
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
