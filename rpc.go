package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

func newRequest() *iterm2.ClientOriginatedMessage {
	return &iterm2.ClientOriginatedMessage{}
}

// RPCError is returned when iTerm2 responds with an error.
type RPCError struct {
	Message string
}

func (e *RPCError) Error() string {
	return "rpc error: " + e.Message
}

// ensureJSONValue encodes a Go string as a JSON value for iTerm2.
// Since the Go API only accepts string, we always encode as a JSON string
// (not inferring types).  This avoids ambiguity like "true" → boolean 1.
func ensureJSONValue(v string) string {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	return string(b)
}

func checkError(resp *iterm2.ServerOriginatedMessage) error {
	if msg := resp.GetError(); msg != "" {
		return &RPCError{Message: msg}
	}
	return nil
}

// ListSessions returns a list of all sessions.
func ListSessions(ctx context.Context, caller Caller) (*iterm2.ListSessionsResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ListSessionsRequest{
		ListSessionsRequest: &iterm2.ListSessionsRequest{},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetListSessionsResponse(), nil
}

// SendText sends text to a session as if typed.
func SendText(ctx context.Context, caller Caller, sessionID string, text string, opts ...SendTextOption) error {
	req := newRequest()
	sendReq := &iterm2.SendTextRequest{
		Session: proto.String(sessionID),
		Text:    proto.String(text),
	}
	for _, o := range opts {
		o(sendReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_SendTextRequest{
		SendTextRequest: sendReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// GetBuffer returns the contents of a session's buffer.
func GetBuffer(ctx context.Context, caller Caller, sessionID string, lineRange *iterm2.LineRange, opts ...GetBufferOption) (*iterm2.GetBufferResponse, error) {
	req := newRequest()
	bufReq := &iterm2.GetBufferRequest{
		Session:   proto.String(sessionID),
		LineRange: lineRange,
	}
	for _, o := range opts {
		o(bufReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_GetBufferRequest{
		GetBufferRequest: bufReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetGetBufferResponse(), nil
}

// GetPrompt returns prompt metadata for a session.
func GetPrompt(ctx context.Context, caller Caller, sessionID string, opts ...GetPromptOption) (*iterm2.GetPromptResponse, error) {
	req := newRequest()
	promptReq := &iterm2.GetPromptRequest{
		Session: proto.String(sessionID),
	}
	for _, o := range opts {
		o(promptReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_GetPromptRequest{
		GetPromptRequest: promptReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetGetPromptResponse(), nil
}

// ListPrompts lists all prompts for a session.
func ListPrompts(ctx context.Context, caller Caller, sessionID string, opts ...ListPromptsOption) (*iterm2.ListPromptsResponse, error) {
	req := newRequest()
	listReq := &iterm2.ListPromptsRequest{
		Session: proto.String(sessionID),
	}
	for _, o := range opts {
		o(listReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_ListPromptsRequest{
		ListPromptsRequest: listReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetListPromptsResponse(), nil
}

// CreateTab creates a new tab.
func CreateTab(ctx context.Context, caller Caller, windowID string, profileName string, opts ...CreateTabOption) (*iterm2.CreateTabResponse, error) {
	req := newRequest()
	createReq := &iterm2.CreateTabRequest{}
	if windowID != "" {
		createReq.WindowId = proto.String(windowID)
	}
	if profileName != "" {
		createReq.ProfileName = proto.String(profileName)
	}
	for _, o := range opts {
		o(createReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_CreateTabRequest{
		CreateTabRequest: createReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetCreateTabResponse(), nil
}

// SplitPane splits a session's pane.
func SplitPane(ctx context.Context, caller Caller, sessionID string, vertical bool, before bool,
	profileName string, opts ...SplitPaneOption,
) (*iterm2.SplitPaneResponse, error) {
	var direction iterm2.SplitPaneRequest_SplitDirection
	if vertical {
		direction = iterm2.SplitPaneRequest_VERTICAL
	} else {
		direction = iterm2.SplitPaneRequest_HORIZONTAL
	}
	req := newRequest()
	splitReq := &iterm2.SplitPaneRequest{
		Session:        proto.String(sessionID),
		SplitDirection: &direction,
		Before:         proto.Bool(before),
		ProfileName:    proto.String(profileName),
	}
	for _, o := range opts {
		o(splitReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_SplitPaneRequest{
		SplitPaneRequest: splitReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetSplitPaneResponse(), nil
}

// GetProfileProperty gets a profile property.
func GetProfileProperty(ctx context.Context, caller Caller, sessionID string, keys []string) (*iterm2.GetProfilePropertyResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_GetProfilePropertyRequest{
		GetProfilePropertyRequest: &iterm2.GetProfilePropertyRequest{
			Session: proto.String(sessionID),
			Keys:    keys,
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetGetProfilePropertyResponse(), nil
}

// SetProfileProperty sets a profile property.
func SetProfileProperty(ctx context.Context, caller Caller, sessionID string, key string, jsonValue string) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_SetProfilePropertyRequest{
		SetProfilePropertyRequest: &iterm2.SetProfilePropertyRequest{
			Target: &iterm2.SetProfilePropertyRequest_Session{
				Session: sessionID,
			},
			Assignments: []*iterm2.SetProfilePropertyRequest_Assignment{
				{Key: proto.String(key), JsonValue: proto.String(jsonValue)},
			},
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// ListProfiles lists all available profiles.
func ListProfiles(ctx context.Context, caller Caller, properties []string, guids []string) (*iterm2.ListProfilesResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ListProfilesRequest{
		ListProfilesRequest: &iterm2.ListProfilesRequest{
			Properties: properties,
			Guids:      guids,
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetListProfilesResponse(), nil
}

// Inject injects bytes directly into the terminal.
// sessionIDs must not be empty.
func Inject(ctx context.Context, caller Caller, sessionIDs []string, data []byte) error {
	if len(sessionIDs) == 0 {
		return fmt.Errorf("term2go: Inject requires at least one sessionID")
	}
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_InjectRequest{
		InjectRequest: &iterm2.InjectRequest{
			SessionId: sessionIDs,
			Data:      data,
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// Activate activates a session/tab/window/app.
func Activate(ctx context.Context, caller Caller, sessionID string, orderWindowFront bool, selectTab bool, opts ...ActivateOption) error {
	req := newRequest()
	activateReq := &iterm2.ActivateRequest{
		Identifier: &iterm2.ActivateRequest_SessionId{
			SessionId: sessionID,
		},
		OrderWindowFront: proto.Bool(orderWindowFront),
		SelectTab:        proto.Bool(selectTab),
	}
	for _, o := range opts {
		o(activateReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_ActivateRequest{
		ActivateRequest: activateReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// GetVariable gets session variables.
func GetVariable(ctx context.Context, caller Caller, sessionID string, names []string) ([]string, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_SessionId{
				SessionId: sessionID,
			},
			Get: names,
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetVariableResponse().GetValues(), nil
}

// SetVariable sets a session variable.
func SetVariable(ctx context.Context, caller Caller, sessionID string, name string, value string) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_SessionId{
				SessionId: sessionID,
			},
			Set: []*iterm2.VariableRequest_Set{
				{Name: proto.String(name), Value: proto.String(ensureJSONValue(value))},
			},
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// SetProperty sets a property on a window or session.
func SetProperty(ctx context.Context, caller Caller, sessionID string, name string, jsonValue string) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_SetPropertyRequest{
		SetPropertyRequest: &iterm2.SetPropertyRequest{
			Identifier: &iterm2.SetPropertyRequest_SessionId{
				SessionId: sessionID,
			},
			Name:      proto.String(name),
			JsonValue: proto.String(jsonValue),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// SetBuried sets or unsets the buried (minimized) state of a session.
func SetBuried(ctx context.Context, caller Caller, sessionID string, buried bool) error {
	v := "false"
	if buried {
		v = "true"
	}
	return SetProperty(ctx, caller, sessionID, "buried", v)
}

// SetGridSize sets the visible grid size of a session.
func SetGridSize(ctx context.Context, caller Caller, sessionID string, width, height int32) error {
	jsonValue := fmt.Sprintf(`{"width":%d,"height":%d}`, width, height)
	return SetProperty(ctx, caller, sessionID, "grid_size", jsonValue)
}

// Close closes a session, tab, or window.
func Close(ctx context.Context, caller Caller, sessionID string, opts ...CloseOption) error {
	req := newRequest()
	closeReq := &iterm2.CloseRequest{
		Target: &iterm2.CloseRequest_Sessions{
			Sessions: &iterm2.CloseRequest_CloseSessions{
				SessionIds: []string{sessionID},
			},
		},
	}
	for _, o := range opts {
		o(closeReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_CloseRequest{
		CloseRequest: closeReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// RestartSession restarts a session.
func RestartSession(ctx context.Context, caller Caller, sessionID string, opts ...RestartSessionOption) error {
	req := newRequest()
	restartReq := &iterm2.RestartSessionRequest{
		SessionId: proto.String(sessionID),
	}
	for _, o := range opts {
		o(restartReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_RestartSessionRequest{
		RestartSessionRequest: restartReq,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// CloseForce closes a session with force=true.
// This is a convenience function equivalent to Close(ctx, caller, sessionID, WithCloseForce(true)).
func CloseForce(ctx context.Context, caller Caller, sessionID string) error {
	return Close(ctx, caller, sessionID, WithCloseForce(true))
}

// SendTextNoBroadcast sends text to a session with suppress_broadcast=true.
// This is a convenience function equivalent to SendText(ctx, caller, sessionID, text, WithSendTextSuppressBroadcast(true)).
func SendTextNoBroadcast(ctx context.Context, caller Caller, sessionID string, text string) error {
	return SendText(ctx, caller, sessionID, text, WithSendTextSuppressBroadcast(true))
}

// RestartSessionIfExited restarts a session only if it has exited.
// This is a convenience function equivalent to RestartSession(ctx, caller, sessionID, WithRestartOnlyIfExited(true)).
func RestartSessionIfExited(ctx context.Context, caller Caller, sessionID string) error {
	return RestartSession(ctx, caller, sessionID, WithRestartOnlyIfExited(true))
}

// NotificationRequest sends a notification subscription request.
func NotificationRequest(ctx context.Context, caller Caller, subscribe bool, notificationType iterm2.NotificationType,
	sessionID string,
) (*iterm2.NotificationResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_NotificationRequest{
		NotificationRequest: &iterm2.NotificationRequest{
			Subscribe:        proto.Bool(subscribe),
			NotificationType: notificationType.Enum(),
		},
	}
	if sessionID != "" {
		req.GetNotificationRequest().Session = proto.String(sessionID)
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetNotificationResponse(), nil
}

// GetProperty gets a property from a window or session.
func GetProperty(ctx context.Context, caller Caller, sessionID string, name string) (*iterm2.GetPropertyResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_GetPropertyRequest{
		GetPropertyRequest: &iterm2.GetPropertyRequest{
			Identifier: &iterm2.GetPropertyRequest_SessionId{
				SessionId: sessionID,
			},
			Name: proto.String(name),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetGetPropertyResponse(), nil
}

// FocusRequest returns information about the currently focused element.
func FocusRequest(ctx context.Context, caller Caller) (*iterm2.FocusResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_FocusRequest{
		FocusRequest: &iterm2.FocusRequest{},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetFocusResponse(), nil
}

// SelectionRequest returns the current selection.
func SelectionRequest(ctx context.Context, caller Caller, sessionID string) (*iterm2.SelectionResponse, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_SelectionRequest{
		SelectionRequest: &iterm2.SelectionRequest{
			Request: &iterm2.SelectionRequest_GetSelectionRequest_{
				GetSelectionRequest: &iterm2.SelectionRequest_GetSelectionRequest{
					SessionId: proto.String(sessionID),
				},
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
	return resp.GetSelectionResponse(), nil
}

// SetSelection sets the selection on a session.
func SetSelection(ctx context.Context, caller Caller, sessionID string, selection *iterm2.Selection) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_SelectionRequest{
		SelectionRequest: &iterm2.SelectionRequest{
			Request: &iterm2.SelectionRequest_SetSelectionRequest_{
				SetSelectionRequest: &iterm2.SelectionRequest_SetSelectionRequest{
					SessionId: proto.String(sessionID),
					Selection: selection,
				},
			},
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// PreferencesRequest gets or sets preferences.
func PreferencesRequest(ctx context.Context, caller Caller, req *iterm2.PreferencesRequest) (*iterm2.PreferencesResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_PreferencesRequest{
		PreferencesRequest: req,
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetPreferencesResponse(), nil
}

// TmuxRequest sends a tmux command.
func TmuxRequest(ctx context.Context, caller Caller, req *iterm2.TmuxRequest) (*iterm2.TmuxResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_TmuxRequest{
		TmuxRequest: req,
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetTmuxResponse(), nil
}

// SavedArrangementRequest manages saved window arrangements.
func SavedArrangementRequest(ctx context.Context, caller Caller, req *iterm2.SavedArrangementRequest) (*iterm2.SavedArrangementResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_SavedArrangementRequest{
		SavedArrangementRequest: req,
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetSavedArrangementResponse(), nil
}

// InvokeFunction invokes a registered function.
func InvokeFunction(ctx context.Context, caller Caller, req *iterm2.InvokeFunctionRequest) (*iterm2.InvokeFunctionResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_InvokeFunctionRequest{
		InvokeFunctionRequest: req,
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetInvokeFunctionResponse(), nil
}

// ServerOriginatedRPCResultRequest sends the result of a server-originated RPC.
func ServerOriginatedRPCResultRequest(ctx context.Context, caller Caller, req *iterm2.ServerOriginatedRPCResultRequest) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_ServerOriginatedRpcResultRequest{
		ServerOriginatedRpcResultRequest: req,
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// SetTabLayout adjusts the split-pane sizes of a tab. The root tree must
// match the tab's actual split structure exactly (only grid_sizes may change).
func SetTabLayout(ctx context.Context, caller Caller, tabID string, root *iterm2.SplitTreeNode) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_SetTabLayoutRequest{
		SetTabLayoutRequest: &iterm2.SetTabLayoutRequest{
			TabId: proto.String(tabID),
			Root:  root,
		},
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// GetSelection returns the current text selection in a session.
func GetSelection(ctx context.Context, caller Caller, sessionID string) (*iterm2.SelectionResponse_GetSelectionResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_SelectionRequest{
		SelectionRequest: &iterm2.SelectionRequest{
			Request: &iterm2.SelectionRequest_GetSelectionRequest_{
				GetSelectionRequest: &iterm2.SelectionRequest_GetSelectionRequest{
					SessionId: proto.String(sessionID),
				},
			},
		},
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetSelectionResponse().GetGetSelectionResponse(), nil
}

// GetBroadcastDomains returns the current broadcast domains.
func GetBroadcastDomains(ctx context.Context, caller Caller) (*iterm2.GetBroadcastDomainsResponse, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_GetBroadcastDomainsRequest{
		GetBroadcastDomainsRequest: &iterm2.GetBroadcastDomainsRequest{},
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetGetBroadcastDomainsResponse(), nil
}

// SetBroadcastDomains sets the broadcast domains.
func SetBroadcastDomains(ctx context.Context, caller Caller, domains []*iterm2.BroadcastDomain) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_SetBroadcastDomainsRequest{
		SetBroadcastDomainsRequest: &iterm2.SetBroadcastDomainsRequest{
			BroadcastDomains: domains,
		},
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}

// ReorderTabs reorders tabs within windows.
// Each assignment specifies a window and the desired tab order.
func ReorderTabs(ctx context.Context, caller Caller, assignments []*iterm2.ReorderTabsRequest_Assignment) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_ReorderTabsRequest{
		ReorderTabsRequest: &iterm2.ReorderTabsRequest{
			Assignments: assignments,
		},
	}
	resp, err := caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}
