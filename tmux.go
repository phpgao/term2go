package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TmuxConnection represents an open tmux integration connection.
type TmuxConnection struct {
	caller Caller

	// ConnectionID uniquely identifies this tmux connection within iTerm2.
	ConnectionID string

	// OwningSessionID is the iTerm2 session that owns this tmux connection.
	OwningSessionID string
}

// SendCommand sends a tmux command on this connection and returns the output.
func (t *TmuxConnection) SendCommand(ctx context.Context, command string) (string, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_TmuxRequest{
		TmuxRequest: &iterm2.TmuxRequest{
			Payload: &iterm2.TmuxRequest_SendCommand_{
				SendCommand: &iterm2.TmuxRequest_SendCommand{
					ConnectionId: proto.String(t.ConnectionID),
					Command:      proto.String(command),
				},
			},
		},
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return "", err
	}
	if err = checkError(resp); err != nil {
		return "", err
	}
	tr := resp.GetTmuxResponse()
	if tr.GetStatus() == iterm2.TmuxResponse_INVALID_CONNECTION_ID {
		return "", fmt.Errorf("tmux: invalid connection id %q", t.ConnectionID)
	}
	return tr.GetSendCommand().GetOutput(), nil
}

// SetWindowVisible shows or hides a tmux window.
func (t *TmuxConnection) SetWindowVisible(ctx context.Context, windowID string, visible bool) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_TmuxRequest{
		TmuxRequest: &iterm2.TmuxRequest{
			Payload: &iterm2.TmuxRequest_SetWindowVisible_{
				SetWindowVisible: &iterm2.TmuxRequest_SetWindowVisible{
					ConnectionId: proto.String(t.ConnectionID),
					WindowId:     proto.String(windowID),
					Visible:      proto.Bool(visible),
				},
			},
		},
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return err
	}
	if err = checkError(resp); err != nil {
		return err
	}
	tr := resp.GetTmuxResponse()
	if s := tr.GetStatus(); s == iterm2.TmuxResponse_INVALID_CONNECTION_ID {
		return fmt.Errorf("tmux: invalid connection id %q", t.ConnectionID)
	} else if s == iterm2.TmuxResponse_INVALID_WINDOW_ID {
		return fmt.Errorf("tmux: invalid window id %q", windowID)
	}
	return nil
}

// CreateWindow creates a new tmux window on this connection.
// affinity is optional — pass "" for none.
// Returns the new iTerm2 tab ID.
func (t *TmuxConnection) CreateWindow(ctx context.Context, affinity string) (string, error) {
	cw := &iterm2.TmuxRequest_CreateWindow{
		ConnectionId: proto.String(t.ConnectionID),
	}
	if affinity != "" {
		cw.Affinity = proto.String(affinity)
	}
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_TmuxRequest{
		TmuxRequest: &iterm2.TmuxRequest{
			Payload: &iterm2.TmuxRequest_CreateWindow_{
				CreateWindow: cw,
			},
		},
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return "", err
	}
	if err = checkError(resp); err != nil {
		return "", err
	}
	tr := resp.GetTmuxResponse()
	if tr.GetStatus() == iterm2.TmuxResponse_INVALID_CONNECTION_ID {
		return "", fmt.Errorf("tmux: invalid connection id %q", t.ConnectionID)
	}
	return tr.GetCreateWindow().GetTabId(), nil
}

// Top-level helpers

// GetTmuxConnections returns all open tmux connections.
func GetTmuxConnections(ctx context.Context, caller Caller) ([]*TmuxConnection, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_TmuxRequest{
		TmuxRequest: &iterm2.TmuxRequest{
			Payload: &iterm2.TmuxRequest_ListConnections_{
				ListConnections: &iterm2.TmuxRequest_ListConnections{},
			},
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	lc := resp.GetTmuxResponse().GetListConnections()
	conns := make([]*TmuxConnection, len(lc.GetConnections()))
	for i, pc := range lc.GetConnections() {
		conns[i] = &TmuxConnection{
			caller:          caller,
			ConnectionID:    pc.GetConnectionId(),
			OwningSessionID: pc.GetOwningSessionId(),
		}
	}
	return conns, nil
}

// GetTmuxConnectionByID finds a single tmux connection by ID.
// Returns nil if not found (no error).
func GetTmuxConnectionByID(ctx context.Context, caller Caller, id string) (*TmuxConnection, error) {
	conns, err := GetTmuxConnections(ctx, caller)
	if err != nil {
		return nil, err
	}
	for _, tc := range conns {
		if tc.ConnectionID == id {
			return tc, nil
		}
	}
	return nil, nil
}
