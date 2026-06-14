package term2go

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// StatusBar Icon
// ============================================================================

// StatusBarIcon is a PNG icon for a status bar component.
// Scale gives the ratio between pixels and points (2 for retina, 1 for regular).
type StatusBarIcon struct {
	Scale float64
	Data  []byte // raw PNG bytes
}

// ============================================================================
// StatusBar Format
// ============================================================================

// StatusBarFormat describes how a component's output is formatted.
type StatusBarFormat int

const (
	StatusBarFormatPlainText StatusBarFormat = 0
	StatusBarFormatHTML      StatusBarFormat = 1
)

// ============================================================================
// Knob helpers
// ============================================================================

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

// ============================================================================
// StatusBarComponent
// ============================================================================

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

// ============================================================================
// OpenPopover
// ============================================================================

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

// ============================================================================
// Register
// ============================================================================

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
