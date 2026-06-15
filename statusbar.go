package term2go

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// StatusBarIcon is a PNG icon for a status bar component.
// Scale gives the ratio between pixels and points (2 for retina, 1 for regular).
type StatusBarIcon struct {
	Scale float64
	Data  []byte // raw PNG bytes
}

// StatusBarFormat describes how a component's output is formatted.
type StatusBarFormat int

const (
	StatusBarFormatPlainText StatusBarFormat = 0
	StatusBarFormatHTML      StatusBarFormat = 1
)

// CheckboxKnob returns a (key, value) pair for a checkbox knob.
func CheckboxKnob(key string, defaultValue bool) (string, string) {
	if defaultValue {
		return key, "true"
	}
	return key, "false"
}

// StringKnob returns a (key, value) pair for a string knob.
// The value is JSON-encoded.
func StringKnob(key string, defaultValue string) (string, string) {
	v, _ := json.Marshal(defaultValue)
	return key, string(v)
}

// FloatKnob returns a (key, value) pair for a float knob.
// The value is JSON-encoded.
func FloatKnob(key string, defaultValue float64) (string, string) {
	v, _ := json.Marshal(defaultValue)
	return key, string(v)
}

// ColorKnob returns a (key, value) pair for a color knob.
// The value should be a JSON-encoded color (e.g., from Color.JSON()).
func ColorKnob(key string, colorJSON string) (string, string) {
	return key, colorJSON
}

// StatusBarComponent describes a script-provided status bar component.
type StatusBarComponent struct {
	ShortDescription    string
	DetailedDescription string
	Knobs               map[string]string
	Exemplar            string
	UpdateCadence       float64 // seconds, 0 means no timer reload
	Identifier          string
	Icons               []StatusBarIcon
	Format              StatusBarFormat
}

// OpenStatusBarPopover opens a popover with HTML content from a status bar component.
func OpenStatusBarPopover(ctx context.Context, caller Caller, identifier, sessionID, html string, width, height int32) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_StatusBarComponentRequest{
		StatusBarComponentRequest: &iterm2.StatusBarComponentRequest{
			Request: &iterm2.StatusBarComponentRequest_OpenPopover_{
				OpenPopover: &iterm2.StatusBarComponentRequest_OpenPopover{
					SessionId: proto.String(sessionID),
					Html:      proto.String(html),
					Size: &iterm2.Size{
						Width:  proto.Int32(width),
						Height: proto.Int32(height),
					},
				},
			},
			Identifier: proto.String(identifier),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	if err = checkError(resp); err != nil {
		return err
	}
	sbResp := resp.GetStatusBarComponentResponse()
	if sbResp != nil && sbResp.GetStatus() != iterm2.StatusBarComponentResponse_OK {
		return &RPCError{Message: sbResp.GetStatus().String()}
	}
	return nil
}

// RegisterStatusBarComponent registers a status bar component with iTerm2.
func RegisterStatusBarComponent(ctx context.Context, caller Caller, component StatusBarComponent) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_RegisterToolRequest{
		RegisterToolRequest: &iterm2.RegisterToolRequest{
			Name:                      proto.String(component.Identifier),
			Identifier:                proto.String(component.Identifier),
			RevealIfAlreadyRegistered: proto.Bool(true),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	if err = checkError(resp); err != nil {
		return err
	}
	regResp := resp.GetRegisterToolResponse()
	if regResp != nil && regResp.GetStatus() != iterm2.RegisterToolResponse_OK {
		return &RPCError{Message: regResp.GetStatus().String()}
	}
	return nil
}

// OnClickHandler is called when a status bar component is clicked.
// It receives the ID of the session owning the status bar component.
type OnClickHandler func(sessionID string)

// SetUnreadCount sets the unread count badge on this status bar component.
// Pass count=0 to remove the badge. If sessionID is empty, updates all instances.
func (c *StatusBarComponent) SetUnreadCount(ctx context.Context, caller Caller, sessionID string, count int) error {
	invocation := fmt.Sprintf(`iterm2.set_status_bar_component_unread_count(identifier: "%s", count: %d)`, c.Identifier, count)
	if sessionID != "" {
		_, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{
			Context: &iterm2.InvokeFunctionRequest_Method_{
				Method: &iterm2.InvokeFunctionRequest_Method{
					Receiver: proto.String(sessionID),
				},
			},
			Timeout:    proto.Float64(-1),
			Invocation: proto.String(invocation),
		})
		return err
	}
	_, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{
		Invocation: proto.String(invocation),
		Timeout:    proto.Float64(-1),
	})
	return err
}

// magicClickName returns the magic RPC name for the onclick handler of this component.
func (c *StatusBarComponent) magicClickName() string {
	magic := "__" + c.Identifier
	magic = strings.NewReplacer(".", "_", "-", "_").Replace(magic)
	return magic + "__on_click"
}

// RegisterWithOnClick registers the status bar component and an onclick handler.
// The onclick callback is invoked in a new goroutine when the user clicks the component.
// It receives the session ID of the session whose status bar was clicked.
func (c *StatusBarComponent) RegisterWithOnClick(ctx context.Context, caller Caller, conn *Connection, onclick OnClickHandler) error {
	// 1. Register the component itself
	if err := RegisterStatusBarComponent(ctx, caller, *c); err != nil {
		return fmt.Errorf("register component: %w", err)
	}

	// 2. Create registry and register the onclick RPC
	reg := NewRPCRegistry(conn)
	rpcName := c.magicClickName()
	if err := reg.Register(ctx, caller, RPCRegistration{
		Name:      rpcName,
		Arguments: []string{"session_id"},
		Timeout:   5,
	}, func(ctx context.Context, args RPCArgs) (interface{}, error) {
		sid, _ := args["session_id"].(string)
		if sid != "" {
			go onclick(sid)
		}
		return nil, nil
	}); err != nil {
		return fmt.Errorf("register onclick rpc: %w", err)
	}

	return nil
}
