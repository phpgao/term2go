package term2go

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestTab_Select tests Tab.Select method.
func TestTab_Select(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.Select(ctx)
	assert.NoError(t, err)
}

// TestTab_Close tests Tab.Close method.
func TestTab_Close(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.Close(ctx, WithCloseForce(true))
	assert.NoError(t, err)
	assert.True(t, mc.req.GetCloseRequest().GetForce())
}

// TestTab_Close_WithOptions tests Tab.Close with Option.
func TestTab_Close_WithOptions(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{caller: mc, ID: "t1"}
	err := tab.Close(ctx)
	assert.NoError(t, err)
	assert.False(t, mc.req.GetCloseRequest().GetForce())
}

// TestTab_UpdateLayout tests Tab.UpdateLayout method.
func TestTab_UpdateLayout(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	// Create mock expecting SetTabLayout request
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}

	// Create a simple splitter with one session
	caller := &mockCaller{}
	session := &Session{caller: caller, ID: "sess-1"}
	root := &Splitter{
		Vertical: false,
		Children: []SplitChild{
			{Session: session},
		},
	}

	tab := &Tab{
		caller: mc,
		ID:     "t1",
		Root:   root,
	}

	err := tab.UpdateLayout(ctx)
	assert.NoError(t, err)

	// Verify request was sent
	assert.NotNil(t, mc.req)
	assert.Equal(t, "t1", mc.req.GetSetTabLayoutRequest().GetTabId())
}

// TestTab_UpdateLayout_NilRoot tests Tab.UpdateLayout method with nil root splitter.
func TestTab_UpdateLayout_NilRoot(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	tab := &Tab{
		caller: mc,
		ID:     "t1",
		Root:   nil,
	}

	err := tab.UpdateLayout(ctx)
	assert.NoError(t, err)

	// With nil root, SetTabLayout should be called with nil root
	assert.NotNil(t, mc.req)
	assert.Nil(t, mc.req.GetSetTabLayoutRequest().GetRoot())
}

// TestTab_UpdateLayout_WithError tests Tab.UpdateLayout method when the caller returns an error.
func TestTab_UpdateLayout_WithError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		err: errors.New("connection refused"),
	}
	tab := &Tab{caller: mc, ID: "t1"}

	err := tab.UpdateLayout(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestSession_GetID tests Session.GetID method.
func TestSession_GetID(t *testing.T) {
	s := &Session{ID: "test-id"}
	assert.Equal(t, "test-id", s.GetID())
}

// TestSession_SetName tests Session.SetName method.
func TestSession_SetName(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "sess-1"}

	err := s.SetName(ctx, "My Session")
	assert.NoError(t, err)

	// Verify InvokeFunction was called
	assert.NotNil(t, mc.req)
	assert.Contains(t, mc.req.GetInvokeFunctionRequest().GetInvocation(), "set_name")
}

// TestSession_SetName_WithError tests Session.SetName method when the caller returns an error.
func TestSession_SetName_WithError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: errorResp("session not found"),
	}
	s := &Session{caller: mc, ID: "sess-1"}

	err := s.SetName(ctx, "My Session")
	assert.Error(t, err)
}

// TestSession_SetBadge tests Session.SetBadge method.
func TestSession_SetBadge(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "sess-1"}

	err := s.SetBadge(ctx, "● NEW")
	assert.NoError(t, err)

	// Verify SetProfileProperty was called
	assert.NotNil(t, mc.req)
	assert.Equal(t, "Badge Text", mc.req.GetSetProfilePropertyRequest().GetAssignments()[0].GetKey())
}

// TestSession_SetBadge_WithError tests Session.SetBadge error path.
func TestSession_SetBadge_WithError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: errorResp("permission denied"),
	}
	s := &Session{caller: mc, ID: "sess-1"}

	err := s.SetBadge(ctx, "● NEW")
	assert.Error(t, err)
}

// TestSession_GetLineInfo tests Session.GetLineInfo method.
func TestSession_GetLineInfo(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetPropertyResponse{
			JsonValue: ptrString(`{"grid":24,"history":1000,"overflow":0,"first_visible":0}`),
		}),
	}
	s := &Session{caller: mc, ID: "sess-1"}

	info, err := s.GetLineInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 24, info.MutableAreaHeight)
	assert.Equal(t, 1000, info.ScrollbackBufferHeight)
	assert.Equal(t, 0, info.Overflow)
	assert.Equal(t, 0, info.FirstVisibleLineNumber)
}

// TestSession_GetLineInfo_WithError tests Session.GetLineInfo error path.
func TestSession_GetLineInfo_WithError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: errorResp("property not found"),
	}
	s := &Session{caller: mc, ID: "sess-1"}

	_, err := s.GetLineInfo(ctx)
	assert.Error(t, err)
}

// TestSession_GetLineInfo_EmptyResponse tests Session.GetLineInfo handling empty response.
func TestSession_GetLineInfo_EmptyResponse(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetPropertyResponse{
			JsonValue: nil,
		}),
	}
	s := &Session{caller: mc, ID: "sess-1"}

	_, err := s.GetLineInfo(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

// TestSession_GetScreenStreamer tests Session.GetScreenStreamer method.
func TestSession_GetScreenStreamer(t *testing.T) {
	// GetScreenStreamer requires a real *Connection, so we test the error path
	s := &Session{caller: &mockCaller{}, ID: "sess-1"}

	_, err := s.GetScreenStreamer()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session caller is not a *Connection")
}

// helper to create a string pointer
func ptrString(s string) *string {
	return &s
}
