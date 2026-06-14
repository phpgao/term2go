package term2go

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

type mockCaller struct {
	req       *iterm2.ClientOriginatedMessage
	resp      *iterm2.ServerOriginatedMessage
	responses []*iterm2.ServerOriginatedMessage // sequential responses, used before resp
	callCount int
	err       error
}

func (m *mockCaller) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
	m.req = req
	if m.err != nil {
		return nil, m.err
	}
	if m.callCount < len(m.responses) {
		r := m.responses[m.callCount]
		m.callCount++
		return r, nil
	}
	return m.resp, nil
}

func (m *mockCaller) Send(req *iterm2.ClientOriginatedMessage) error {
	m.req = req
	return m.err
}

func successResp(sub proto.Message) *iterm2.ServerOriginatedMessage {
	var s iterm2.ServerOriginatedMessage
	switch v := sub.(type) {
	case *iterm2.ListSessionsResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_ListSessionsResponse{ListSessionsResponse: v}
	case *iterm2.GetBufferResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_GetBufferResponse{GetBufferResponse: v}
	case *iterm2.CreateTabResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_CreateTabResponse{CreateTabResponse: v}
	case *iterm2.SplitPaneResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_SplitPaneResponse{SplitPaneResponse: v}
	case *iterm2.VariableResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_VariableResponse{VariableResponse: v}
	case *iterm2.SelectionResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_SelectionResponse{SelectionResponse: v}
	case *iterm2.NotificationResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_NotificationResponse{NotificationResponse: v}
	case *iterm2.GetPropertyResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_GetPropertyResponse{GetPropertyResponse: v}
	case *iterm2.GetPromptResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_GetPromptResponse{GetPromptResponse: v}
	case *iterm2.ListPromptsResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_ListPromptsResponse{ListPromptsResponse: v}
	case *iterm2.GetProfilePropertyResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_GetProfilePropertyResponse{GetProfilePropertyResponse: v}
	case *iterm2.ListProfilesResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_ListProfilesResponse{ListProfilesResponse: v}
	case *iterm2.FocusResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_FocusResponse{FocusResponse: v}
	case *iterm2.PreferencesResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_PreferencesResponse{PreferencesResponse: v}
	case *iterm2.TmuxResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_TmuxResponse{TmuxResponse: v}
	case *iterm2.SavedArrangementResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_SavedArrangementResponse{SavedArrangementResponse: v}
	case *iterm2.InvokeFunctionResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_InvokeFunctionResponse{InvokeFunctionResponse: v}
	}
	return &s
}

func errorResp(msg string) *iterm2.ServerOriginatedMessage {
	return &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Error{Error: msg},
	}
}

// TestListSessions_Success tests ListSessions function success return.
func TestListSessions_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.ListSessionsResponse{}),
	}
	resp, err := ListSessions(ctx, mc)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestListSessions_CallError tests ListSessions function when the caller returns an error.
func TestListSessions_CallError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("connection error")}
	_, err := ListSessions(ctx, mc)
	require.Error(t, err)
}

// TestListSessions_RPCError tests ListSessions function when iTerm2 returns an RPC error.
func TestListSessions_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc error")}
	_, err := ListSessions(ctx, mc)
	require.Error(t, err)
	var rpcErr *RPCError
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, "rpc error", rpcErr.Message)
}

// TestSendText_Success tests SendText function success return.
func TestSendText_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SendText(ctx, mc, "S1", "hello"))
	req := mc.req.GetSendTextRequest()
	require.NotNil(t, req)
	assert.Equal(t, "S1", req.GetSession())
	assert.Equal(t, "hello", req.GetText())
	assert.False(t, req.GetSuppressBroadcast())
}

// TestSendText_WithOptions tests SendText function with Option.
func TestSendText_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SendText(ctx, mc, "S1", "hello", WithSendTextSuppressBroadcast(true)))
	req := mc.req.GetSendTextRequest()
	assert.True(t, req.GetSuppressBroadcast())
}

// TestSendText_CallError tests SendText function when the caller returns an error.
func TestSendText_CallError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("send error")}
	require.Error(t, SendText(ctx, mc, "S1", "x"))
}

// TestSendText_RPCError tests SendText function when iTerm2 returns an RPC error.
func TestSendText_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc error")}
	require.Error(t, SendText(ctx, mc, "S1", "x"))
}

// TestGetBuffer_WithoutLineRange tests GetBuffer without line range.
func TestGetBuffer_WithoutLineRange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.GetBufferResponse{}),
	}
	resp, err := GetBuffer(ctx, mc, "S1", nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetGetBufferRequest()
	assert.Equal(t, "S1", req.GetSession())
	assert.Nil(t, req.GetLineRange())
}

// TestGetBuffer_WithLineRange tests GetBuffer function with line range parameter.
func TestGetBuffer_WithLineRange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	lr := &iterm2.LineRange{ScreenContentsOnly: proto.Bool(true)}
	mc := &mockCaller{
		resp: successResp(&iterm2.GetBufferResponse{}),
	}
	resp, err := GetBuffer(ctx, mc, "S2", lr)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetGetBufferRequest()
	assert.Equal(t, lr, req.GetLineRange())
	assert.True(t, req.GetLineRange().GetScreenContentsOnly())
}

// TestGetBuffer_Error tests GetBuffer function when the caller returns an error.
func TestGetBuffer_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("conn error")}
	_, err := GetBuffer(ctx, mc, "S1", nil)
	require.Error(t, err)
}

// TestCreateTab_Success tests CreateTab function success return.
func TestCreateTab_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.CreateTabResponse{}),
	}
	resp, err := CreateTab(ctx, mc, "W1", "Default")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetCreateTabRequest()
	assert.Equal(t, "W1", req.GetWindowId())
	assert.Equal(t, "Default", req.GetProfileName())
}

// TestCreateTab_Error tests CreateTab function when the caller returns an error.
func TestCreateTab_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("create error")}
	_, err := CreateTab(ctx, mc, "W1", "Default")
	require.Error(t, err)
}

// TestSplitPane_Vertical tests SplitPane vertical split case.
func TestSplitPane_Vertical(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.SplitPaneResponse{}),
	}
	resp, err := SplitPane(ctx, mc, "S1", true, false, "Default")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetSplitPaneRequest()
	assert.Equal(t, iterm2.SplitPaneRequest_VERTICAL, req.GetSplitDirection())
	assert.False(t, req.GetBefore())
}

// TestSplitPane_Horizontal tests SplitPane function with horizontal split direction.
func TestSplitPane_Horizontal(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.SplitPaneResponse{}),
	}
	resp, err := SplitPane(ctx, mc, "S1", false, true, "Custom")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetSplitPaneRequest()
	assert.Equal(t, iterm2.SplitPaneRequest_HORIZONTAL, req.GetSplitDirection())
	assert.True(t, req.GetBefore())
	assert.Equal(t, "Custom", req.GetProfileName())
}

// TestSplitPane_Error tests SplitPane function when the caller returns an error.
func TestSplitPane_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("split error")}
	_, err := SplitPane(ctx, mc, "S1", true, false, "")
	require.Error(t, err)
}

// TestGetVariable_Success tests GetVariable function success return.
func TestGetVariable_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{
			Values: []string{"val1", "val2"},
		}),
	}
	values, err := GetVariable(ctx, mc, "S1", []string{"v1", "v2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"val1", "val2"}, values)
	req := mc.req.GetVariableRequest()
	assert.Equal(t, "S1", req.GetSessionId())
}

// TestGetVariable_Error tests GetVariable function when the caller returns an error.
func TestGetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("var error")}
	_, err := GetVariable(ctx, mc, "S1", []string{"v1"})
	require.Error(t, err)
}

// TestSetVariable_Success tests SetVariable function success return.
func TestSetVariable_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SetVariable(ctx, mc, "S1", "key", "value"))
	req := mc.req.GetVariableRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	sets := req.GetSet()
	require.Len(t, sets, 1)
	assert.Equal(t, "key", sets[0].GetName())
	assert.Equal(t, `"value"`, sets[0].GetValue())
}

// TestSetVariable_Error tests SetVariable function when the caller returns an error.
func TestSetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("set error")}
	require.Error(t, SetVariable(ctx, mc, "S1", "k", "v"))
}

// TestInject_Success tests Inject function success return.
func TestInject_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	data := []byte("hello")
	require.NoError(t, Inject(ctx, mc, []string{"S1", "S2"}, data))
	req := mc.req.GetInjectRequest()
	assert.Len(t, req.GetSessionId(), 2)
	assert.Equal(t, "hello", string(req.GetData()))
}

// TestInject_Error tests Inject function when the caller returns an error.
func TestInject_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("inject error")}
	require.Error(t, Inject(ctx, mc, []string{"S1"}, []byte{}))
}

// TestActivate_Success tests Activate function success return.
func TestActivate_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, Activate(ctx, mc, "S1", true, true))
	req := mc.req.GetActivateRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	assert.True(t, req.GetOrderWindowFront())
	assert.True(t, req.GetSelectTab())
}

// TestActivate_Error tests Activate function when the caller returns an error.
func TestActivate_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("activate error")}
	require.Error(t, Activate(ctx, mc, "S1", false, false))
}

// TestNotificationRequest_Subscribe tests NotificationRequest subscribe functionality.
func TestNotificationRequest_Subscribe(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.NotificationResponse{}),
	}
	resp, err := NotificationRequest(ctx, mc, true, iterm2.NotificationType_NOTIFY_ON_NEW_SESSION, "")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetNotificationRequest()
	assert.True(t, req.GetSubscribe())
	assert.Equal(t, iterm2.NotificationType_NOTIFY_ON_NEW_SESSION, req.GetNotificationType())
	assert.Empty(t, req.GetSession())
}

// TestNotificationRequest_WithSession tests NotificationRequest function with session scope.
func TestNotificationRequest_WithSession(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.NotificationResponse{}),
	}
	_, err := NotificationRequest(ctx, mc, true, iterm2.NotificationType_NOTIFY_ON_KEYSTROKE, "S1")
	require.NoError(t, err)
	req := mc.req.GetNotificationRequest()
	assert.Equal(t, "S1", req.GetSession())
}

// TestNotificationRequest_Error tests NotificationRequest function when the caller returns an error.
func TestNotificationRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("notif error")}
	_, err := NotificationRequest(ctx, mc, false, iterm2.NotificationType_NOTIFY_ON_NEW_SESSION, "")
	require.Error(t, err)
}

// TestClose_Success tests Close function success return.
func TestClose_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, Close(ctx, mc, "S1"))
	req := mc.req.GetCloseRequest()
	assert.Equal(t, "S1", req.GetSessions().GetSessionIds()[0])
	assert.False(t, req.GetForce())
}

// TestClose_WithOptions tests Close function with Option.
func TestClose_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}

	t.Run("with force", func(t *testing.T) {
		require.NoError(t, Close(ctx, mc, "S1", WithCloseForce(true)))
		req := mc.req.GetCloseRequest()
		assert.True(t, req.GetForce())
	})

	t.Run("close tabs", func(t *testing.T) {
		require.NoError(t, Close(ctx, mc, "S1", WithCloseTabs([]string{"T1", "T2"})))
		req := mc.req.GetCloseRequest()
		assert.Equal(t, []string{"T1", "T2"}, req.GetTabs().GetTabIds())
	})

	t.Run("close windows", func(t *testing.T) {
		require.NoError(t, Close(ctx, mc, "S1", WithCloseWindows([]string{"W1", "W2"})))
		req := mc.req.GetCloseRequest()
		assert.Equal(t, []string{"W1", "W2"}, req.GetWindows().GetWindowIds())
	})
}

// TestClose_Error tests Close function when the caller returns an error.
func TestClose_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("close error")}
	require.Error(t, Close(ctx, mc, "S1"))
}

// TestRestartSession_Success tests RestartSession function success return.
func TestRestartSession_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, RestartSession(ctx, mc, "S1"))
	req := mc.req.GetRestartSessionRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	assert.False(t, req.GetOnlyIfExited())
}

// TestRestartSession_WithOptions tests RestartSession function with Option.
func TestRestartSession_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, RestartSession(ctx, mc, "S1", WithRestartOnlyIfExited(true)))
	req := mc.req.GetRestartSessionRequest()
	assert.True(t, req.GetOnlyIfExited())
}

// TestRestartSession_Error tests RestartSession function when the caller returns an error.
func TestRestartSession_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("restart error")}
	require.Error(t, RestartSession(ctx, mc, "S1"))
}

// TestCloseForce_Success tests CloseForce helper function.
func TestCloseForce_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, CloseForce(ctx, mc, "S1"))
	req := mc.req.GetCloseRequest()
	assert.Equal(t, "S1", req.GetSessions().GetSessionIds()[0])
	assert.True(t, req.GetForce())
}

// TestCloseForce_Error tests CloseForce error handling.
func TestCloseForce_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("close force error")}
	require.Error(t, CloseForce(ctx, mc, "S1"))
}

// TestSendTextNoBroadcast_Success tests SendTextNoBroadcast helper function.
func TestSendTextNoBroadcast_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SendTextNoBroadcast(ctx, mc, "S1", "hello"))
	req := mc.req.GetSendTextRequest()
	assert.Equal(t, "S1", req.GetSession())
	assert.Equal(t, "hello", req.GetText())
	assert.True(t, req.GetSuppressBroadcast())
}

// TestSendTextNoBroadcast_Error tests SendTextNoBroadcast error handling.
func TestSendTextNoBroadcast_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("send text no broadcast error")}
	require.Error(t, SendTextNoBroadcast(ctx, mc, "S1", "hello"))
}

// TestRestartSessionIfExited_Success tests RestartSessionIfExited helper function.
func TestRestartSessionIfExited_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, RestartSessionIfExited(ctx, mc, "S1"))
	req := mc.req.GetRestartSessionRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	assert.True(t, req.GetOnlyIfExited())
}

// TestRestartSessionIfExited_Error tests RestartSessionIfExited error handling.
func TestRestartSessionIfExited_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("restart session if exited error")}
	require.Error(t, RestartSessionIfExited(ctx, mc, "S1"))
}

// TestSelectionRequest_Success tests SelectionRequest function success return.
func TestSelectionRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.SelectionResponse{}),
	}
	resp, err := SelectionRequest(ctx, mc, "S1")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetSelectionRequest()
	sel := req.GetGetSelectionRequest()
	require.NotNil(t, sel)
	assert.Equal(t, "S1", sel.GetSessionId())
}

// TestSelectionRequest_Error tests SelectionRequest function when the caller returns an error.
func TestSelectionRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sel error")}
	_, err := SelectionRequest(ctx, mc, "S1")
	require.Error(t, err)
}

// TestSetSelection_Success tests SetSelection function success return.
func TestSetSelection_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	sel := &iterm2.Selection{
		SubSelections: []*iterm2.SubSelection{
			{WindowedCoordRange: &iterm2.WindowedCoordRange{
				CoordRange: &iterm2.CoordRange{},
			}},
		},
	}
	require.NoError(t, SetSelection(ctx, mc, "S1", sel))
	req := mc.req.GetSelectionRequest()
	sreq := req.GetSetSelectionRequest()
	require.NotNil(t, sreq)
	assert.Equal(t, "S1", sreq.GetSessionId())
}

// TestSetSelection_Error tests SetSelection function when the caller returns an error.
func TestSetSelection_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("set sel error")}
	require.Error(t, SetSelection(ctx, mc, "S1", &iterm2.Selection{}))
}

// TestGetPrompt_Success tests GetPrompt function for retrieving prompt information.
func TestGetPrompt_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_GetPromptResponse{GetPromptResponse: &iterm2.GetPromptResponse{}}}}
	_, err := GetPrompt(ctx, mc, "s1")
	require.NoError(t, err)
}

// TestListPrompts_Success tests ListPrompts function for retrieving all prompts.
func TestListPrompts_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_ListPromptsResponse{ListPromptsResponse: &iterm2.ListPromptsResponse{}}}}
	_, err := ListPrompts(ctx, mc, "s1")
	require.NoError(t, err)
}

// TestGetProfileProperty_Success tests GetProfileProperty function for retrieving profile properties.
func TestGetProfileProperty_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_GetProfilePropertyResponse{GetProfilePropertyResponse: &iterm2.GetProfilePropertyResponse{}}}}
	_, err := GetProfileProperty(ctx, mc, "s1", []string{"Name"})
	require.NoError(t, err)
}

// TestSetProfileProperty_Success tests SetProfileProperty function for setting profile properties.
func TestSetProfileProperty_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SetProfileProperty(ctx, mc, "s1", "Name", `"Default"`))
}

// TestListProfiles_Success tests ListProfiles function for retrieving all profiles.
func TestListProfiles_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_ListProfilesResponse{ListProfilesResponse: &iterm2.ListProfilesResponse{}}}}
	_, err := ListProfiles(ctx, mc, nil, nil)
	require.NoError(t, err)
}

// TestFocusRequest_Success tests FocusRequest function for retrieving current focus information.
func TestFocusRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_FocusResponse{FocusResponse: &iterm2.FocusResponse{}}}}
	_, err := FocusRequest(ctx, mc)
	require.NoError(t, err)
}

// TestPreferencesRequest_Success tests PreferencesRequest function for retrieving preferences.
func TestPreferencesRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_PreferencesResponse{PreferencesResponse: &iterm2.PreferencesResponse{}}}}
	_, err := PreferencesRequest(ctx, mc, &iterm2.PreferencesRequest{})
	require.NoError(t, err)
}

// TestTmuxRequest_Success tests TmuxRequest function for interacting with tmux.
func TestTmuxRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{TmuxResponse: &iterm2.TmuxResponse{}}}}
	_, err := TmuxRequest(ctx, mc, &iterm2.TmuxRequest{})
	require.NoError(t, err)
}

// TestSavedArrangementRequest_Success tests SavedArrangementRequest function for managing saved arrangements.
func TestSavedArrangementRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_SavedArrangementResponse{SavedArrangementResponse: &iterm2.SavedArrangementResponse{}}}}
	_, err := SavedArrangementRequest(ctx, mc, &iterm2.SavedArrangementRequest{})
	require.NoError(t, err)
}

// TestInvokeFunction_Success tests InvokeFunction function for invoking a function.
func TestInvokeFunction_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{Submessage: &iterm2.ServerOriginatedMessage_InvokeFunctionResponse{InvokeFunctionResponse: &iterm2.InvokeFunctionResponse{}}}}
	_, err := InvokeFunction(ctx, mc, &iterm2.InvokeFunctionRequest{})
	require.NoError(t, err)
}

// TestInvokeFunction_Error tests InvokeFunction function when the caller returns an error.
func TestInvokeFunction_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call failed")}
	_, err := InvokeFunction(ctx, mc, &iterm2.InvokeFunctionRequest{})
	require.Error(t, err)
}

// TestInvokeFunction_RPCError tests InvokeFunction function when iTerm2 returns an RPC error.
func TestInvokeFunction_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := InvokeFunction(ctx, mc, &iterm2.InvokeFunctionRequest{})
	require.Error(t, err)
}

// TestRPCError_Error tests RPCError.Error method returns formatted error message.
func TestRPCError_Error(t *testing.T) {
	e := &RPCError{Message: "test"}
	assert.Equal(t, "rpc error: test", e.Error())
}

// TestEscapeAppleScript tests escapeAppleScript function for escaping special characters.
func TestEscapeAppleScript(t *testing.T) {
	r := escapeAppleScript(`hello "world"`)
	assert.Equal(t, `hello \"world\"`, r)
	// Backslash must be escaped before double-quote to avoid double-escaping.
	r = escapeAppleScript(`path\to\"file"`)
	assert.Equal(t, `path\\to\\\"file\"`, r)
}

// TestSetProperty_Success tests SetProperty function success return.
func TestSetProperty_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, SetProperty(ctx, mc, "S1", "name", `"value"`))
	req := mc.req.GetSetPropertyRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	assert.Equal(t, "name", req.GetName())
	assert.Equal(t, `"value"`, req.GetJsonValue())
}

// TestSetProperty_Error tests SetProperty function when the caller returns an error.
func TestSetProperty_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("set prop error")}
	require.Error(t, SetProperty(ctx, mc, "S1", "n", "v"))
}

// TestSetProperty_RPCError tests SetProperty function when iTerm2 returns an RPC error.
func TestSetProperty_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, SetProperty(ctx, mc, "S1", "n", "v"))
}

// TestGetProperty_Success tests GetProperty function for retrieving a property value.
func TestGetProperty_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: successResp(&iterm2.GetPropertyResponse{})}
	resp, err := GetProperty(ctx, mc, "S1", "prop")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	req := mc.req.GetGetPropertyRequest()
	assert.Equal(t, "S1", req.GetSessionId())
	assert.Equal(t, "prop", req.GetName())
}

// TestGetProperty_Error tests GetProperty function when the caller returns an error.
func TestGetProperty_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("get prop error")}
	_, err := GetProperty(ctx, mc, "S1", "p")
	require.Error(t, err)
}

// TestGetProperty_RPCError tests GetProperty function when iTerm2 returns an RPC error.
func TestGetProperty_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := GetProperty(ctx, mc, "S1", "p")
	require.Error(t, err)
}

// TestServerOriginatedRPCResultRequest_Success tests ServerOriginatedRPCResultRequest function for sending RPC results.
func TestServerOriginatedRPCResultRequest_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	require.NoError(t, ServerOriginatedRPCResultRequest(ctx, mc, &iterm2.ServerOriginatedRPCResultRequest{}))
}

// TestServerOriginatedRPCResultRequest_Error tests ServerOriginatedRPCResultRequest function when the caller returns an error.
func TestServerOriginatedRPCResultRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("rpc result error")}
	require.Error(t, ServerOriginatedRPCResultRequest(ctx, mc, &iterm2.ServerOriginatedRPCResultRequest{}))
}

// TestServerOriginatedRPCResultRequest_RPCError tests ServerOriginatedRPCResultRequest function when iTerm2 returns an RPC error.
func TestServerOriginatedRPCResultRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, ServerOriginatedRPCResultRequest(ctx, mc, &iterm2.ServerOriginatedRPCResultRequest{}))
}

// TestGetBuffer_RPCError tests GetBuffer RPC error path.
func TestGetBuffer_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := GetBuffer(ctx, mc, "S1", nil)
	require.Error(t, err)
}

// TestCreateTab_RPCError tests CreateTab function when iTerm2 returns an RPC error.
func TestCreateTab_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := CreateTab(ctx, mc, "W1", "Default")
	require.Error(t, err)
}

// TestSplitPane_RPCError tests SplitPane function when iTerm2 returns an RPC error.
func TestSplitPane_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := SplitPane(ctx, mc, "S1", true, false, "")
	require.Error(t, err)
}

// TestGetVariable_RPCError tests GetVariable function when iTerm2 returns an RPC error.
func TestGetVariable_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := GetVariable(ctx, mc, "S1", []string{"v1"})
	require.Error(t, err)
}

// TestSetVariable_RPCError tests SetVariable function when iTerm2 returns an RPC error.
func TestSetVariable_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, SetVariable(ctx, mc, "S1", "k", "v"))
}

// TestEnsureJSONValue tests ensureJSONValue function for JSON encoding values.
func TestEnsureJSONValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Go string → always JSON-encoded as string
		{"42", `"42"`},
		{"-3.14", `"-3.14"`},
		{"true", `"true"`},
		{"false", `"false"`},
		{"null", `"null"`},
		{"8080", `"8080"`},
		// already looks like JSON → still encoded as string (double-encoded)
		{`"hello"`, `"\"hello\""`},
		{`[1,2,3]`, `"[1,2,3]"`},
		{`{"a":1}`, `"{\"a\":1}"`},
		// plain strings
		{"hello", `"hello"`},
		{"hello world", `"hello world"`},
		{"/tmp/test", `"/tmp/test"`},
		{"main", `"main"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ensureJSONValue(tt.input))
		})
	}
}

// TestInject_RPCError tests Inject function when iTerm2 returns an RPC error.
func TestInject_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, Inject(ctx, mc, []string{"S1"}, []byte{}))
}

// TestActivate_RPCError tests Activate function when iTerm2 returns an RPC error.
func TestActivate_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, Activate(ctx, mc, "S1", false, false))
}

// TestClose_RPCError tests Close function when iTerm2 returns an RPC error.
func TestClose_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, Close(ctx, mc, "S1"))
}

// TestNotificationRequest_RPCError tests NotificationRequest function when iTerm2 returns an RPC error.
func TestNotificationRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := NotificationRequest(ctx, mc, false, iterm2.NotificationType_NOTIFY_ON_NEW_SESSION, "")
	require.Error(t, err)
}

// TestRestartSession_RPCError tests RestartSession function when iTerm2 returns an RPC error.
func TestRestartSession_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, RestartSession(ctx, mc, "S1"))
}

// TestSelectionRequest_RPCError tests SelectionRequest function when iTerm2 returns an RPC error.
func TestSelectionRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := SelectionRequest(ctx, mc, "S1")
	require.Error(t, err)
}

// TestSetSelection_RPCError tests SetSelection function when iTerm2 returns an RPC error.
func TestSetSelection_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, SetSelection(ctx, mc, "S1", &iterm2.Selection{}))
}

// TestGetPrompt_Error tests GetPrompt error path.
func TestGetPrompt_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := GetPrompt(ctx, mc, "s1")
	require.Error(t, err)
}

// TestGetPrompt_RPCError tests GetPrompt function when iTerm2 returns an RPC error.
func TestGetPrompt_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := GetPrompt(ctx, mc, "s1")
	require.Error(t, err)
}

// TestListPrompts_Error tests ListPrompts function when the caller returns an error.
func TestListPrompts_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := ListPrompts(ctx, mc, "s1")
	require.Error(t, err)
}

// TestListPrompts_RPCError tests ListPrompts function when iTerm2 returns an RPC error.
func TestListPrompts_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := ListPrompts(ctx, mc, "s1")
	require.Error(t, err)
}

// TestGetProfileProperty_Error tests GetProfileProperty function when the caller returns an error.
func TestGetProfileProperty_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := GetProfileProperty(ctx, mc, "s1", []string{"Name"})
	require.Error(t, err)
}

// TestGetProfileProperty_RPCError tests GetProfileProperty function when iTerm2 returns an RPC error.
func TestGetProfileProperty_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := GetProfileProperty(ctx, mc, "s1", []string{"Name"})
	require.Error(t, err)
}

// TestSetProfileProperty_Error tests SetProfileProperty function when the caller returns an error.
func TestSetProfileProperty_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	require.Error(t, SetProfileProperty(ctx, mc, "s1", "Name", `"v"`))
}

// TestSetProfileProperty_RPCError tests SetProfileProperty function when iTerm2 returns an RPC error.
func TestSetProfileProperty_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	require.Error(t, SetProfileProperty(ctx, mc, "s1", "Name", `"v"`))
}

// TestListProfiles_Error tests ListProfiles function when the caller returns an error.
func TestListProfiles_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := ListProfiles(ctx, mc, nil, nil)
	require.Error(t, err)
}

// TestListProfiles_RPCError tests ListProfiles function when iTerm2 returns an RPC error.
func TestListProfiles_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := ListProfiles(ctx, mc, nil, nil)
	require.Error(t, err)
}

// TestFocusRequest_Error tests FocusRequest function when the caller returns an error.
func TestFocusRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := FocusRequest(ctx, mc)
	require.Error(t, err)
}

// TestFocusRequest_RPCError tests FocusRequest function when iTerm2 returns an RPC error.
func TestFocusRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := FocusRequest(ctx, mc)
	require.Error(t, err)
}

// TestPreferencesRequest_Error tests PreferencesRequest function when the caller returns an error.
func TestPreferencesRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := PreferencesRequest(ctx, mc, &iterm2.PreferencesRequest{})
	require.Error(t, err)
}

// TestPreferencesRequest_RPCError tests PreferencesRequest function when iTerm2 returns an RPC error.
func TestPreferencesRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := PreferencesRequest(ctx, mc, &iterm2.PreferencesRequest{})
	require.Error(t, err)
}

// TestTmuxRequest_Error tests TmuxRequest function when the caller returns an error.
func TestTmuxRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := TmuxRequest(ctx, mc, &iterm2.TmuxRequest{})
	require.Error(t, err)
}

// TestTmuxRequest_RPCError tests TmuxRequest function when iTerm2 returns an RPC error.
func TestTmuxRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := TmuxRequest(ctx, mc, &iterm2.TmuxRequest{})
	require.Error(t, err)
}

// TestSavedArrangementRequest_Error tests SavedArrangementRequest function when the caller returns an error.
func TestSavedArrangementRequest_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	_, err := SavedArrangementRequest(ctx, mc, &iterm2.SavedArrangementRequest{})
	require.Error(t, err)
}

// TestSavedArrangementRequest_RPCError tests SavedArrangementRequest function when iTerm2 returns an RPC error.
func TestSavedArrangementRequest_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	_, err := SavedArrangementRequest(ctx, mc, &iterm2.SavedArrangementRequest{})
	require.Error(t, err)
}
