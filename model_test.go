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

func makeSessionSummary(id string) *iterm2.SessionSummary {
	return &iterm2.SessionSummary{UniqueIdentifier: proto.String(id)}
}

func makeListResp() *iterm2.ListSessionsResponse {
	return &iterm2.ListSessionsResponse{
		Windows: []*iterm2.ListSessionsResponse_Window{
			{
				WindowId: proto.String("w1"),
				Tabs: []*iterm2.ListSessionsResponse_Tab{
					{
						TabId: proto.String("t1"),
						Root: &iterm2.SplitTreeNode{
							Vertical: proto.Bool(false),
							Links: []*iterm2.SplitTreeNode_SplitTreeLink{
								{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s1")}},
								{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s2")}},
							},
						},
					},
				},
			},
		},
	}
}

// listCaller is a test helper for GetApp and appFromListSessionsResponse.
type listCaller struct {
	resp *iterm2.ListSessionsResponse
	err  error
}

func (l *listCaller) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
	if l.err != nil {
		return nil, l.err
	}
	return &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_ListSessionsResponse{
			ListSessionsResponse: l.resp,
		},
	}, nil
}

func (l *listCaller) Send(req *iterm2.ClientOriginatedMessage) error {
	return l.err
}

// TestGetApp_Success tests GetApp function returns App structure from iTerm2 response.
func TestGetApp_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &listCaller{resp: makeListResp()}
	app, err := GetApp(ctx, mc)
	require.NoError(t, err)
	require.Len(t, app.Windows, 1)
	assert.Equal(t, "w1", app.Windows[0].ID)
	require.Len(t, app.Windows[0].Tabs, 1)
}

// TestGetApp_Error tests GetApp function when the caller returns an error.
func TestGetApp_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &listCaller{err: errors.New("conn error")}
	_, err := GetApp(ctx, mc)
	require.Error(t, err)
}

// TestAppFromListSessionsResponse tests AppFromListSessionsResponse function for parsing iTerm2 response.
func TestAppFromListSessionsResponse(t *testing.T) {
	mc := &listCaller{}
	resp := &iterm2.ListSessionsResponse{
		Windows: []*iterm2.ListSessionsResponse_Window{
			{
				WindowId: proto.String("win1"),
				Number:   proto.Int32(1),
				Frame: &iterm2.Frame{
					Origin: &iterm2.Point{X: proto.Int32(0), Y: proto.Int32(0)},
					Size:   &iterm2.Size{Width: proto.Int32(800), Height: proto.Int32(600)},
				},
				Tabs: []*iterm2.ListSessionsResponse_Tab{
					{
						TabId: proto.String("tab1"),
						Root: &iterm2.SplitTreeNode{
							Vertical: proto.Bool(false),
							Links: []*iterm2.SplitTreeNode_SplitTreeLink{
								{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s1")}},
							},
						},
					},
					{
						TabId: proto.String("tab2"),
						Root: &iterm2.SplitTreeNode{
							Vertical: proto.Bool(true),
							Links: []*iterm2.SplitTreeNode_SplitTreeLink{
								{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s2")}},
							},
						},
					},
				},
			},
			{
				WindowId: proto.String("win2"),
				Tabs:     []*iterm2.ListSessionsResponse_Tab{},
			},
		},
	}
	app := appFromListSessionsResponse(mc, resp)

	require.Len(t, app.Windows, 2)
	assert.Equal(t, "win2", app.Windows[1].ID)
	require.Len(t, app.Windows[0].Tabs, 2)
	f := app.Windows[0].Frame
	require.NotNil(t, f)
	assert.Equal(t, int32(800), f.Size.Width)
	assert.Equal(t, int32(600), f.Size.Height)
}

// TestApp_Refresh tests App.Refresh method for updating the app state from iTerm2.
func TestApp_Refresh(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &listCaller{resp: makeListResp()}
	app, err := GetApp(ctx, mc)
	require.NoError(t, err)

	mc.resp = &iterm2.ListSessionsResponse{
		Windows: []*iterm2.ListSessionsResponse_Window{
			{WindowId: proto.String("w1"), Tabs: []*iterm2.ListSessionsResponse_Tab{}},
			{WindowId: proto.String("w2"), Tabs: []*iterm2.ListSessionsResponse_Tab{}},
		},
	}

	require.NoError(t, app.Refresh(ctx))
	assert.Len(t, app.Windows, 2)
}

// TestSession_SendText tests Session.SendText method.
func TestSession_SendText(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.SendText(ctx, "hello"))
	req := mc.req.GetSendTextRequest()
	require.NotNil(t, req)
	assert.Equal(t, "s1", req.GetSession())
	assert.Equal(t, "hello", req.GetText())
	assert.False(t, req.GetSuppressBroadcast())
}

// TestSession_SendText_WithOptions tests Session.SendText with Option.
func TestSession_SendText_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.SendText(ctx, "hello", WithSendTextSuppressBroadcast(true)))
	assert.True(t, mc.req.GetSendTextRequest().GetSuppressBroadcast())
}

// TestSession_GetBuffer tests Session.GetBuffer method for reading terminal buffer.
func TestSession_GetBuffer(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.GetBufferResponse{}),
	}
	s := &Session{caller: mc, ID: "s1"}
	resp, err := s.GetBuffer(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "s1", mc.req.GetGetBufferRequest().GetSession())
}

// TestSession_SplitPane tests Session.SplitPane method for splitting the pane.
func TestSession_SplitPane(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.SplitPaneResponse{
			SessionId: []string{"new-session"},
		}),
	}
	s := &Session{caller: mc, ID: "s1"}
	ns, err := s.SplitPane(ctx, true, false, "Default")
	require.NoError(t, err)
	assert.Equal(t, "new-session", ns.ID)
	req := mc.req.GetSplitPaneRequest()
	assert.Equal(t, iterm2.SplitPaneRequest_VERTICAL, req.GetSplitDirection())
}

// TestSession_SplitPane_Error tests Session.SplitPane method when the caller returns an error.
func TestSession_SplitPane_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("split error")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.SplitPane(ctx, false, true, "")
	require.Error(t, err)
}

// TestSession_SetVariable tests Session.SetVariable method for setting a session variable.
func TestSession_SetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.SetVariable(ctx, "key", "val"))
	req := mc.req.GetVariableRequest()
	sets := req.GetSet()
	require.Len(t, sets, 1)
	assert.Equal(t, "key", sets[0].GetName())
	assert.Equal(t, `"val"`, sets[0].GetValue())
}

// TestSession_GetVariable tests Session.GetVariable method for getting a session variable.
func TestSession_GetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// iTerm2 returns JSON-encoded values: the string "val" is encoded as `"val"`
	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{Values: []string{`"val"`}}),
	}
	s := &Session{caller: mc, ID: "s1"}
	v, err := s.GetVariable(ctx, "name")
	require.NoError(t, err)
	assert.Equal(t, "val", v) // decoded from JSON
}

// TestSession_GetVariable_Empty tests Session.GetVariable method returns empty string when variable is not set.
func TestSession_GetVariable_Empty(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{Values: []string{}}),
	}
	s := &Session{caller: mc, ID: "s1"}
	v, err := s.GetVariable(ctx, "name")
	require.NoError(t, err)
	assert.Empty(t, v)
}

// TestJsonDecodeForVariable tests jsonDecodeForVariable function for decoding JSON values.
func TestJsonDecodeForVariable(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		// JSON strings → decoded inner string
		{`string`, `"hello"`, "hello"},
		{`encoded path`, `"/tmp/test"`, "/tmp/test"},
		// JSON numbers → preserve raw representation
		{`int`, "42", "42"},
		{`float`, "3.14", "3.14"},
		// JSON bool → preserve raw
		{`true`, "true", "true"},
		{`false`, "false", "false"},
		// JSON null → empty string
		{`null`, "null", ""},
		// Invalid JSON → return raw
		{`bare string`, "hello", "hello"},
		{`empty`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, jsonDecodeForVariable(tt.raw))
		})
	}
}

// TestSession_Inject tests Session.Inject method for injecting raw bytes.
func TestSession_Inject(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.Inject(ctx, []byte("data")))
	req := mc.req.GetInjectRequest()
	require.Len(t, req.GetSessionId(), 1)
	assert.Equal(t, "s1", req.GetSessionId()[0])
}

// TestSession_Close tests Session.Close method for closing a session.
func TestSession_Close(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.Close(ctx, WithCloseForce(true)))
	req := mc.req.GetCloseRequest()
	ids := req.GetSessions().GetSessionIds()
	require.Len(t, ids, 1)
	assert.Equal(t, "s1", ids[0])
	assert.True(t, req.GetForce())
}

// TestSession_Close_WithOptions tests Session.Close with Option.
func TestSession_Close_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.Close(ctx))
	assert.False(t, mc.req.GetCloseRequest().GetForce())
}

// TestWindow_CreateTab tests Window.CreateTab method.
func TestWindow_CreateTab(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		responses: []*iterm2.ServerOriginatedMessage{
			successResp(&iterm2.CreateTabResponse{SessionId: proto.String("new-tab")}),
			successResp(&iterm2.ListSessionsResponse{
				Windows: []*iterm2.ListSessionsResponse_Window{
					{
						WindowId: proto.String("w1"),
						Tabs: []*iterm2.ListSessionsResponse_Tab{
							{
								TabId: proto.String("tab-1"),
								Root: &iterm2.SplitTreeNode{
									Links: []*iterm2.SplitTreeNode_SplitTreeLink{
										{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("new-tab")}},
									},
								},
							},
						},
					},
				},
			}),
		},
	}
	w := &Window{caller: mc, ID: "w1"}
	tab, err := w.CreateTab(ctx, "Default")
	require.NoError(t, err)
	assert.Equal(t, "tab-1", tab.ID)
}

// TestWindow_Close tests Window.Close method for closing a window.
func TestWindow_Close(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	w := &Window{caller: mc, ID: "w1"}
	require.NoError(t, w.Close(ctx, WithCloseForce(true)))
	req := mc.req.GetCloseRequest()
	ids := req.GetWindows().GetWindowIds()
	require.Len(t, ids, 1)
	assert.Equal(t, "w1", ids[0])
	assert.True(t, req.GetForce())
}

// TestWindow_Close_WithOptions tests Window.Close with Option.
func TestWindow_Close_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	w := &Window{caller: mc, ID: "w1"}
	require.NoError(t, w.Close(ctx))
	assert.False(t, mc.req.GetCloseRequest().GetForce())
}

// TestSplitterFromProto_Simple tests converting Splitter from simple proto node.
func TestSplitterFromProto_Simple(t *testing.T) {
	node := &iterm2.SplitTreeNode{
		Vertical: proto.Bool(true),
		Links: []*iterm2.SplitTreeNode_SplitTreeLink{
			{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s1")}},
			{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s2")}},
		},
	}
	s := SplitterFromProto(node, &listCaller{})
	require.NotNil(t, s)
	assert.True(t, s.Vertical)
	require.Len(t, s.Children, 2)
}

// TestSplitterFromProto_Nested tests SplitterFromProto function for nested splitter structure.
func TestSplitterFromProto_Nested(t *testing.T) {
	node := &iterm2.SplitTreeNode{
		Vertical: proto.Bool(false),
		Links: []*iterm2.SplitTreeNode_SplitTreeLink{
			{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s1")}},
			{
				Child: &iterm2.SplitTreeNode_SplitTreeLink_Node{
					Node: &iterm2.SplitTreeNode{
						Vertical: proto.Bool(true),
						Links: []*iterm2.SplitTreeNode_SplitTreeLink{
							{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s2")}},
							{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("s3")}},
						},
					},
				},
			},
		},
	}
	s := SplitterFromProto(node, &listCaller{})

	assert.False(t, s.Vertical)
	require.Len(t, s.Children, 2)
	sess := s.Children[0].Session
	require.NotNil(t, sess)
	assert.Equal(t, "s1", sess.ID)
	sub := s.Children[1].Splitter
	require.NotNil(t, sub)
	assert.True(t, sub.Vertical)
	assert.Len(t, sub.Children, 2)
}

// TestSplitterFromProto_Nil tests SplitterFromProto function returns nil for nil input.
func TestSplitterFromProto_Nil(t *testing.T) {
	s := SplitterFromProto(nil, nil)
	assert.Nil(t, s)
}

// TestSplitter_Sessions tests Splitter.Sessions method for collecting all sessions recursively.
func TestSplitter_Sessions(t *testing.T) {
	node := &iterm2.SplitTreeNode{
		Links: []*iterm2.SplitTreeNode_SplitTreeLink{
			{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("a")}},
			{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("b")}},
			{
				Child: &iterm2.SplitTreeNode_SplitTreeLink_Node{
					Node: &iterm2.SplitTreeNode{
						Links: []*iterm2.SplitTreeNode_SplitTreeLink{
							{Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{Session: makeSessionSummary("c")}},
						},
					},
				},
			},
		},
	}
	s := SplitterFromProto(node, &listCaller{})

	sessions := s.Sessions()
	require.Len(t, sessions, 3)
	assert.Equal(t, "a", sessions[0].ID)
	assert.Equal(t, "b", sessions[1].ID)
	assert.Equal(t, "c", sessions[2].ID)
}

// TestSplitter_Sessions_Empty tests Splitter.Sessions method returns empty slice when no children.
func TestSplitter_Sessions_Empty(t *testing.T) {
	s := &Splitter{Children: nil}
	sessions := s.Sessions()
	assert.Empty(t, sessions)
}

// TestApp_Refresh_Error tests App.Refresh method error path.
func TestApp_Refresh_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &listCaller{err: errors.New("conn error")}
	app := &App{caller: mc}
	require.Error(t, app.Refresh(ctx))
}

// TestWindow_CreateTab_Error tests Window.CreateTab method when the caller returns an error.
func TestWindow_CreateTab_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("create error")}
	w := &Window{caller: mc, ID: "w1"}
	_, err := w.CreateTab(ctx, "Default")
	require.Error(t, err)
}

// TestWindow_Close_Error tests Window.Close method when the caller returns an error.
func TestWindow_Close_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("close error")}
	w := &Window{caller: mc, ID: "w1"}
	require.Error(t, w.Close(ctx))
}

// TestWindow_Close_RPCError tests Window.Close method when iTerm2 returns an RPC error.
func TestWindow_Close_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	w := &Window{caller: mc, ID: "w1"}
	require.Error(t, w.Close(ctx))
}

// TestTab_Select_Error tests Tab.Select method when the caller returns an error.
func TestTab_Select_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("call error")}
	tab := &Tab{caller: mc, ID: "t1"}
	require.Error(t, tab.Select(ctx))
}

// TestTab_Select_RPCError tests Tab.Select method when iTerm2 returns an RPC error.
func TestTab_Select_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	tab := &Tab{caller: mc, ID: "t1"}
	require.Error(t, tab.Select(ctx))
}

// TestTab_Close_Error tests Tab.Close method when the caller returns an error.
func TestTab_Close_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("close error")}
	tab := &Tab{caller: mc, ID: "t1"}
	require.Error(t, tab.Close(ctx))
}

// TestTab_Close_RPCError tests Tab.Close method when iTerm2 returns an RPC error.
func TestTab_Close_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	tab := &Tab{caller: mc, ID: "t1"}
	require.Error(t, tab.Close(ctx))
}

// TestSession_SplitPane_EmptySessionID tests Session.SplitPane method returns error when no session ID returned.
func TestSession_SplitPane_EmptySessionID(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{
		resp: successResp(&iterm2.SplitPaneResponse{}), // no session IDs
	}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.SplitPane(ctx, true, false, "Default")
	require.Error(t, err)
}

// TestSession_GetVariable_Error tests Session.GetVariable method when the caller returns an error.
func TestSession_GetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("var error")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.GetVariable(ctx, "name")
	require.Error(t, err)
}

// TestSession_SendText_Error tests Session.SendText method when the caller returns an error.
func TestSession_SendText_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("send error")}
	s := &Session{caller: mc, ID: "s1"}
	require.Error(t, s.SendText(ctx, "hello"))
}

// TestSession_GetBuffer_Error tests Session.GetBuffer method when the caller returns an error.
func TestSession_GetBuffer_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("buf error")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.GetBuffer(ctx, nil)
	require.Error(t, err)
}

// TestSession_Inject_Error tests Session.Inject method when the caller returns an error.
func TestSession_Inject_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("inject error")}
	s := &Session{caller: mc, ID: "s1"}
	require.Error(t, s.Inject(ctx, []byte("data")))
}

// TestSession_Close_Error tests Session.Close method when the caller returns an error.
func TestSession_Close_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("close error")}
	s := &Session{caller: mc, ID: "s1"}
	require.Error(t, s.Close(ctx))
}

// TestSession_SetVariable_Error tests Session.SetVariable method when the caller returns an error.
func TestSession_SetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("set error")}
	s := &Session{caller: mc, ID: "s1"}
	require.Error(t, s.SetVariable(ctx, "k", "v"))
}

// TestSession_SetBuried tests Session.SetBuried method for burying a session.
func TestSession_SetBuried(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.SetBuried(ctx, true))
	assert.Equal(t, "s1", mc.req.GetSetPropertyRequest().GetSessionId())
	assert.Equal(t, "buried", mc.req.GetSetPropertyRequest().GetName())
	assert.Equal(t, "true", mc.req.GetSetPropertyRequest().GetJsonValue())
}

// TestSession_SetBuried_False tests Session.SetBuried method for unburying a session.
func TestSession_SetBuried_False(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s2"}
	require.NoError(t, s.SetBuried(ctx, false))
	assert.Equal(t, "false", mc.req.GetSetPropertyRequest().GetJsonValue())
}

// TestSession_SetGridSize tests Session.SetGridSize method for setting terminal grid size.
func TestSession_SetGridSize(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	require.NoError(t, s.SetGridSize(ctx, 80, 24))
	assert.Equal(t, "s1", mc.req.GetSetPropertyRequest().GetSessionId())
	assert.Equal(t, "grid_size", mc.req.GetSetPropertyRequest().GetName())
	assert.Equal(t, `{"width":80,"height":24}`, mc.req.GetSetPropertyRequest().GetJsonValue())
}
