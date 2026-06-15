package term2go

import (
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// VariableScope constants

func TestVariableScope_Values(t *testing.T) {
	assert.Equal(t, VariableScope(0), VariableScopeSession)
	assert.Equal(t, VariableScope(1), VariableScopeTab)
	assert.Equal(t, VariableScope(2), VariableScopeWindow)
	assert.Equal(t, VariableScope(3), VariableScopeApp)
}

// NewVariableMonitor

func TestNewVariableMonitor(t *testing.T) {
	conn := &Connection{}
	m := NewVariableMonitor(conn, VariableScopeSession, "jobName", "s1")
	require.NotNil(t, m)
	assert.Equal(t, VariableScopeSession, m.scope)
	assert.Equal(t, "jobName", m.name)
	assert.Equal(t, "s1", m.identifier)
	assert.False(t, m.started)
}

func TestNewVariableMonitor_AppScope(t *testing.T) {
	m := NewVariableMonitor(&Connection{}, VariableScopeApp, "effectiveTheme", "")
	require.NotNil(t, m)
	assert.Equal(t, VariableScopeApp, m.scope)
}

// VariableMonitor.Changes channel

func TestVariableMonitor_Changes(t *testing.T) {
	m := NewVariableMonitor(&Connection{}, VariableScopeSession, "jobName", "s1")
	ch := m.Changes()
	assert.NotNil(t, ch)
}

// VariableMonitor.Stop on unstarted monitor

func TestVariableMonitor_Stop_Unstarted(t *testing.T) {
	m := NewVariableMonitor(&Connection{}, VariableScopeSession, "jobName", "s1")
	m.Stop() // should not panic
}

// App.GetVariable

func TestApp_GetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{Values: []string{`"dark"`}}),
	}
	app := &App{caller: mc}
	val, err := app.GetVariable(ctx, "effectiveTheme")
	require.NoError(t, err)
	assert.Equal(t, "dark", val)
	// Verify correct scope
	vr := mc.req.GetVariableRequest()
	assert.True(t, vr.GetApp())
}

func TestApp_GetVariable_Empty(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{}),
	}
	app := &App{caller: mc}
	val, err := app.GetVariable(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, val)
}

func TestApp_GetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	app := &App{caller: mc}
	_, err := app.GetVariable(ctx, "key")
	require.Error(t, err)
}

// App.SetVariable

func TestApp_SetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	app := &App{caller: mc}
	err := app.SetVariable(ctx, "user.myVar", "hello")
	require.NoError(t, err)
	vr := mc.req.GetVariableRequest()
	assert.True(t, vr.GetApp())
	require.Len(t, vr.GetSet(), 1)
	assert.Equal(t, "user.myVar", vr.GetSet()[0].GetName())
	assert.Equal(t, `"hello"`, vr.GetSet()[0].GetValue())
}

func TestApp_SetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	app := &App{caller: mc}
	err := app.SetVariable(ctx, "user.myVar", "val")
	require.Error(t, err)
}

// Window.GetVariable / Window.SetVariable

func TestWindow_GetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{Values: []string{`"MyTitle"`}}),
	}
	w := &Window{caller: mc, ID: "w1"}
	val, err := w.GetVariable(ctx, "title")
	require.NoError(t, err)
	assert.Equal(t, "MyTitle", val)
	assert.Equal(t, "w1", mc.req.GetVariableRequest().GetWindowId())
}

func TestWindow_GetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	w := &Window{caller: mc, ID: "w1"}
	_, err := w.GetVariable(ctx, "key")
	require.Error(t, err)
}

func TestWindow_SetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	w := &Window{caller: mc, ID: "w1"}
	err := w.SetVariable(ctx, "user.myVar", "42")
	require.NoError(t, err)
	vr := mc.req.GetVariableRequest()
	assert.Equal(t, "w1", vr.GetWindowId())
	require.Len(t, vr.GetSet(), 1)
	assert.Equal(t, "user.myVar", vr.GetSet()[0].GetName())
	assert.Equal(t, `"42"`, vr.GetSet()[0].GetValue())
}

func TestWindow_SetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	w := &Window{caller: mc, ID: "w1"}
	err := w.SetVariable(ctx, "user.key", "v")
	require.Error(t, err)
}

// Tab.GetVariable / Tab.SetVariable

func TestTab_GetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.VariableResponse{Values: []string{`"MyTab"`}}),
	}
	tab := &Tab{caller: mc, ID: "t1"}
	val, err := tab.GetVariable(ctx, "title")
	require.NoError(t, err)
	assert.Equal(t, "MyTab", val)
	assert.Equal(t, "t1", mc.req.GetVariableRequest().GetTabId())
}

func TestTab_GetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	tab := &Tab{caller: mc, ID: "t1"}
	_, err := tab.GetVariable(ctx, "key")
	require.Error(t, err)
}

func TestTab_SetVariable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.SetVariable(ctx, "user.myVar", "tabVal")
	require.NoError(t, err)
	vr := mc.req.GetVariableRequest()
	assert.Equal(t, "t1", vr.GetTabId())
	require.Len(t, vr.GetSet(), 1)
	assert.Equal(t, "user.myVar", vr.GetSet()[0].GetName())
	assert.Equal(t, `"tabVal"`, vr.GetSet()[0].GetValue())
}

func TestTab_SetVariable_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.SetVariable(ctx, "user.key", "v")
	require.Error(t, err)
}

// Tab.SelectPaneInDirection

func TestTab_SelectPaneInDirection(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}

	err := tab.SelectPaneInDirection(ctx, DirectionLeft)
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "select_pane_in_direction")
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "left")
}

func TestTab_SelectPaneInDirection_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.SelectPaneInDirection(ctx, DirectionRight)
	require.Error(t, err)
}

// Tab.SetTitle

func TestTab_SetTitle(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}

	err := tab.SetTitle(ctx, "My Tab")
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "set_title")
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "My Tab")
}

func TestTab_SetTitle_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.SetTitle(ctx, "Title")
	require.Error(t, err)
}

// Tab.CurrentSession

func TestTab_CurrentSession(t *testing.T) {
	tab := &Tab{
		ID:              "t1",
		ActiveSessionID: "s2",
		Root: &Splitter{
			Children: []SplitChild{
				{Session: &Session{ID: "s1"}},
				{Session: &Session{ID: "s2"}},
			},
		},
	}
	s := tab.CurrentSession()
	require.NotNil(t, s)
	assert.Equal(t, "s2", s.ID)
}

func TestTab_CurrentSession_NoActive(t *testing.T) {
	tab := &Tab{
		ID: "t1",
		Root: &Splitter{
			Children: []SplitChild{
				{Session: &Session{ID: "s1"}},
			},
		},
	}
	assert.Nil(t, tab.CurrentSession())
}

func TestTab_CurrentSession_NotFound(t *testing.T) {
	tab := &Tab{
		ID:              "t1",
		ActiveSessionID: "s3",
		Root: &Splitter{
			Children: []SplitChild{
				{Session: &Session{ID: "s1"}},
			},
		},
	}
	assert.Nil(t, tab.CurrentSession())
}

// Session.MoveToNewTab / MoveToNewWindow

func TestSession_MoveToNewTab(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.MoveToNewTab(ctx, nil, nil)
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "move_session_to_new_tab")
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "s1")
}

func TestSession_MoveToNewTab_WithWindow(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}
	w := &Window{ID: "w1"}

	err := s.MoveToNewTab(ctx, w, nil)
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "window_id")
}

func TestSession_MoveToNewWindow(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.MoveToNewWindow(ctx)
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "move_session_to_new_window")
}

// Session.RunCoprocess / StopCoprocess

func TestSession_RunCoprocess(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.RunCoprocess(ctx, "/usr/bin/cat")
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "run_coprocess")
}

func TestSession_StopCoprocess(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.StopCoprocess(ctx)
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "stop_coprocess")
}

// Session.RunTmuxCommand

func TestSession_RunTmuxCommand(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.InvokeFunctionResponse{
			Disposition: &iterm2.InvokeFunctionResponse_Success_{
				Success: &iterm2.InvokeFunctionResponse_Success{
					JsonResult: proto.String(`"tmux-output"`),
				},
			},
		}),
	}
	s := &Session{caller: mc, ID: "s1"}

	result, err := s.RunTmuxCommand(ctx, "list-windows")
	require.NoError(t, err)
	assert.Equal(t, "tmux-output", result)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "run_tmux_command")
}

func TestSession_RunTmuxCommand_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.RunTmuxCommand(ctx, "cmd")
	require.Error(t, err)
}

// Session.AddAnnotation

func TestSession_AddAnnotation(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.AddAnnotation(ctx, 0, 10, 80, 15, "note")
	require.NoError(t, err)
	inv := mc.req.GetInvokeFunctionRequest().GetInvocation()
	assert.Contains(t, inv, "add_annotation")
	assert.Contains(t, inv, "note")
}

// Session.LoadURL

func TestSession_LoadURL(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.LoadURL(ctx, "https://example.com")
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "load_url")
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "example.com")
}

// Session.Activate

func TestSession_Activate(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.Activate(ctx, true, true)
	require.NoError(t, err)
	ar := mc.req.GetActivateRequest()
	assert.Equal(t, "s1", ar.GetSessionId())
	assert.True(t, ar.GetSelectTab())
	assert.True(t, ar.GetOrderWindowFront())
}

// Session.Restart

func TestSession_Restart(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	err := s.Restart(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, "s1", mc.req.GetRestartSessionRequest().GetSessionId())
	assert.True(t, mc.req.GetRestartSessionRequest().GetOnlyIfExited())
}

// Session.SplitPaneWithCustomizations

func TestSession_SplitPaneWithCustomizations(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.SplitPaneResponse{SessionId: []string{"new-s"}}),
	}
	s := &Session{caller: mc, ID: "s1"}

	props := []*iterm2.ProfileProperty{
		{Key: proto.String("Font"), JsonValue: proto.String(`"Menlo"`)},
	}
	ns, err := s.SplitPaneWithCustomizations(ctx, false, true, "Default", props)
	require.NoError(t, err)
	assert.Equal(t, "new-s", ns.ID)
	assert.Len(t, mc.req.GetSplitPaneRequest().GetCustomProfileProperties(), 1)
}

// App.GetSessionByID / GetTabByID / GetWindowByID / GetWindowForTab / GetWindowAndTabForSession

func TestApp_GetSessionByID(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	_ = ctx

	app := &App{
		Windows: []*Window{
			{
				ID: "w1",
				Tabs: []*Tab{
					{
						ID: "t1",
						Root: &Splitter{
							Children: []SplitChild{
								{Session: &Session{ID: "s1"}},
								{Session: &Session{ID: "s2"}},
							},
						},
					},
				},
			},
		},
	}
	s := app.GetSessionByID("s2")
	require.NotNil(t, s)
	assert.Equal(t, "s2", s.ID)
}

func TestApp_GetSessionByID_Buried(t *testing.T) {
	app := &App{
		BuriedSessions: []*Session{{ID: "b1"}},
	}
	s := app.GetSessionByID("b1")
	require.NotNil(t, s)
	assert.Equal(t, "b1", s.ID)
}

func TestApp_GetSessionByID_NotFound(t *testing.T) {
	app := &App{}
	assert.Nil(t, app.GetSessionByID("nonexistent"))
}

func TestApp_GetTabByID(t *testing.T) {
	app := &App{
		Windows: []*Window{
			{ID: "w1", Tabs: []*Tab{{ID: "t1"}}},
		},
	}
	tab := app.GetTabByID("t1")
	require.NotNil(t, tab)
	assert.Equal(t, "t1", tab.ID)
}

func TestApp_GetTabByID_NotFound(t *testing.T) {
	assert.Nil(t, (&App{}).GetTabByID("nonexistent"))
}

func TestApp_GetWindowByID(t *testing.T) {
	app := &App{
		Windows: []*Window{{ID: "w1"}, {ID: "w2"}},
	}
	w := app.GetWindowByID("w2")
	require.NotNil(t, w)
	assert.Equal(t, "w2", w.ID)
}

func TestApp_GetWindowByID_NotFound(t *testing.T) {
	assert.Nil(t, (&App{}).GetWindowByID("nonexistent"))
}

func TestApp_GetWindowForTab(t *testing.T) {
	app := &App{
		Windows: []*Window{
			{ID: "w1", Tabs: []*Tab{{ID: "t1"}}},
		},
	}
	w := app.GetWindowForTab("t1")
	require.NotNil(t, w)
	assert.Equal(t, "w1", w.ID)
}

func TestApp_GetWindowForTab_NotFound(t *testing.T) {
	assert.Nil(t, (&App{}).GetWindowForTab("nonexistent"))
}

func TestApp_GetWindowAndTabForSession(t *testing.T) {
	app := &App{
		Windows: []*Window{
			{
				ID: "w1",
				Tabs: []*Tab{
					{
						ID: "t1",
						Root: &Splitter{
							Children: []SplitChild{
								{Session: &Session{ID: "s1"}},
							},
						},
					},
				},
			},
		},
	}
	s := &Session{ID: "s1"}
	win, tab := app.GetWindowAndTabForSession(s)
	require.NotNil(t, win)
	assert.Equal(t, "w1", win.ID)
	require.NotNil(t, tab)
	assert.Equal(t, "t1", tab.ID)
}

func TestApp_GetWindowAndTabForSession_NotFound(t *testing.T) {
	app := &App{}
	win, tab := app.GetWindowAndTabForSession(&Session{ID: "x"})
	assert.Nil(t, win)
	assert.Nil(t, tab)
}

// App.CurrentWindow

func TestApp_CurrentWindow(t *testing.T) {
	app := &App{
		Windows:                 []*Window{{ID: "w1"}},
		CurrentTerminalWindowID: "w1",
	}
	w := app.CurrentWindow()
	require.NotNil(t, w)
	assert.Equal(t, "w1", w.ID)
}

func TestApp_CurrentWindow_None(t *testing.T) {
	assert.Nil(t, (&App{}).CurrentWindow())
}

// Window.SetTabs

func TestWindow_SetTabs(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	w := &Window{caller: mc, ID: "w1"}
	tabs := []*Tab{{ID: "t2"}, {ID: "t1"}}

	err := w.SetTabs(ctx, tabs)
	require.NoError(t, err)
	as := mc.req.GetReorderTabsRequest().GetAssignments()
	require.Len(t, as, 1)
	assert.Equal(t, "w1", as[0].GetWindowId())
	assert.Equal(t, []string{"t2", "t1"}, as[0].GetTabIds())
}

// Window.SetTitle

func TestWindow_SetTitle(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	w := &Window{caller: mc, ID: "w1"}

	err := w.SetTitle(ctx, "My Window")
	require.NoError(t, err)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "set_title")
}

// Window.CurrentTab

func TestWindow_CurrentTab(t *testing.T) {
	w := &Window{
		ID:            "w1",
		SelectedTabID: "t2",
		Tabs: []*Tab{
			{ID: "t1"},
			{ID: "t2"},
		},
	}
	tab := w.CurrentTab()
	require.NotNil(t, tab)
	assert.Equal(t, "t2", tab.ID)
}

func TestWindow_CurrentTab_None(t *testing.T) {
	assert.Nil(t, (&Window{}).CurrentTab())
}
