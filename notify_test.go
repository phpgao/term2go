package term2go

import (
	"errors"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

func makeNotifyResp() *iterm2.ServerOriginatedMessage {
	return &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
			NotificationResponse: &iterm2.NotificationResponse{},
		},
	}
}

// newTestConnWithCaller creates a Connection with mock caller for testing notification RPCs.
func newTestConnWithCaller(mc *mockCaller) *Connection {
	c := NewConnection("", "", "test")
	c.notifyCaller = mc
	return c
}

// TestSubscribeNewSession tests SubscribeNewSession function for subscribing to new session events.
func TestSubscribeNewSession(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeNewSession(ctx, mc, c, func(caller Caller, n *iterm2.NewSessionNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeTerminateSession tests SubscribeTerminateSession function for subscribing to session terminate events.
func TestSubscribeTerminateSession(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeTerminateSession(ctx, mc, c, func(caller Caller, n *iterm2.TerminateSessionNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeKeystroke tests SubscribeKeystroke function for subscribing to keystroke events.
func TestSubscribeKeystroke(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeKeystroke(ctx, mc, c, func(caller Caller, n *iterm2.KeystrokeNotification) {}, "s1")
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeScreenUpdate tests SubscribeScreenUpdate function for subscribing to screen update events.
func TestSubscribeScreenUpdate(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeScreenUpdate(ctx, mc, c, func(caller Caller, n *iterm2.ScreenUpdateNotification) {}, "s1")
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribePrompt tests SubscribePrompt function for subscribing to prompt events.
func TestSubscribePrompt(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribePrompt(ctx, mc, c, func(caller Caller, n *iterm2.PromptNotification) {}, "s1")
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeCustomEscapeSequence tests SubscribeCustomEscapeSequence function for subscribing to custom escape sequence events.
func TestSubscribeCustomEscapeSequence(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeCustomEscapeSequence(ctx, mc, c, func(caller Caller, n *iterm2.CustomEscapeSequenceNotification) {}, "s1")
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeVariableChange tests SubscribeVariableChange function for subscribing to variable change events.
func TestSubscribeVariableChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeVariableChange(ctx, mc, c, func(caller Caller, n *iterm2.VariableChangedNotification) {}, "s1", "jobName")
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeLayoutChange tests SubscribeLayoutChange function for subscribing to layout change events.
func TestSubscribeLayoutChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeLayoutChange(ctx, mc, c, func(caller Caller, n *iterm2.LayoutChangedNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeFocusChange tests SubscribeFocusChange function for subscribing to focus change events.
func TestSubscribeFocusChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeFocusChange(ctx, mc, c, func(caller Caller, n *iterm2.FocusChangedNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeServerOriginatedRPC tests SubscribeServerOriginatedRPC function for subscribing to server originated RPC events.
func TestSubscribeServerOriginatedRPC(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeServerOriginatedRPC(ctx, mc, c, func(caller Caller, n *iterm2.ServerOriginatedRPCNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeBroadcastChange tests SubscribeBroadcastChange function for subscribing to broadcast change events.
func TestSubscribeBroadcastChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeBroadcastChange(ctx, mc, c, func(caller Caller, n *iterm2.BroadcastDomainsChangedNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeProfileChange tests SubscribeProfileChange function for subscribing to profile change events.
func TestSubscribeProfileChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)
	tk, err := SubscribeProfileChange(ctx, mc, c, func(caller Caller, n *iterm2.ProfileChangedNotification) {})
	require.NoError(t, err)
	c.Unsubscribe(tk)
}

// TestSubscribeNewSession_Error tests SubscribeNewSession error path.
func TestSubscribeNewSession_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeNewSession(ctx, mc, c, func(caller Caller, n *iterm2.NewSessionNotification) {})
	require.Error(t, err)
}

// TestSubscribeKeystroke_Error tests SubscribeKeystroke function returns error when the caller fails.
func TestSubscribeKeystroke_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeKeystroke(ctx, mc, c, func(caller Caller, n *iterm2.KeystrokeNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeScreenUpdate_Error tests SubscribeScreenUpdate function returns error when the caller fails.
func TestSubscribeScreenUpdate_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeScreenUpdate(ctx, mc, c, func(caller Caller, n *iterm2.ScreenUpdateNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeVariableChange_Error tests SubscribeVariableChange function returns error when the caller fails.
func TestSubscribeVariableChange_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeVariableChange(ctx, mc, c, func(caller Caller, n *iterm2.VariableChangedNotification) {}, "s1", "jobName")
	require.Error(t, err)
}

// TestSubscribeServerOriginatedRPC_Error tests SubscribeServerOriginatedRPC function returns error when the caller fails.
func TestSubscribeServerOriginatedRPC_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeServerOriginatedRPC(ctx, mc, c, func(caller Caller, n *iterm2.ServerOriginatedRPCNotification) {})
	require.Error(t, err)
}

// TestSubscribeTerminateSession_Error tests SubscribeTerminateSession error path.
func TestSubscribeTerminateSession_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeTerminateSession(ctx, mc, c, func(caller Caller, n *iterm2.TerminateSessionNotification) {})
	require.Error(t, err)
}

// TestSubscribePrompt_Error tests SubscribePrompt function returns error when the caller fails.
func TestSubscribePrompt_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribePrompt(ctx, mc, c, func(caller Caller, n *iterm2.PromptNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeCustomEscapeSequence_Error tests SubscribeCustomEscapeSequence function returns error when the caller fails.
func TestSubscribeCustomEscapeSequence_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeCustomEscapeSequence(ctx, mc, c, func(caller Caller, n *iterm2.CustomEscapeSequenceNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeLayoutChange_Error tests SubscribeLayoutChange function returns error when the caller fails.
func TestSubscribeLayoutChange_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeLayoutChange(ctx, mc, c, func(caller Caller, n *iterm2.LayoutChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeFocusChange_Error tests SubscribeFocusChange function returns error when the caller fails.
func TestSubscribeFocusChange_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeFocusChange(ctx, mc, c, func(caller Caller, n *iterm2.FocusChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeBroadcastChange_Error tests SubscribeBroadcastChange function returns error when the caller fails.
func TestSubscribeBroadcastChange_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeBroadcastChange(ctx, mc, c, func(caller Caller, n *iterm2.BroadcastDomainsChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeProfileChange_Error tests SubscribeProfileChange function returns error when the caller fails.
func TestSubscribeProfileChange_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("sub error")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeProfileChange(ctx, mc, c, func(caller Caller, n *iterm2.ProfileChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeTerminateSession_RPCError tests SubscribeTerminateSession RPC error path.
func TestSubscribeTerminateSession_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeTerminateSession(ctx, mc, c, func(caller Caller, n *iterm2.TerminateSessionNotification) {})
	require.Error(t, err)
}

// TestSubscribeKeystroke_RPCError tests SubscribeKeystroke function returns error when iTerm2 returns an RPC error.
func TestSubscribeKeystroke_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeKeystroke(ctx, mc, c, func(caller Caller, n *iterm2.KeystrokeNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeScreenUpdate_RPCError tests SubscribeScreenUpdate function returns error when iTerm2 returns an RPC error.
func TestSubscribeScreenUpdate_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeScreenUpdate(ctx, mc, c, func(caller Caller, n *iterm2.ScreenUpdateNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribePrompt_RPCError tests SubscribePrompt function returns error when iTerm2 returns an RPC error.
func TestSubscribePrompt_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribePrompt(ctx, mc, c, func(caller Caller, n *iterm2.PromptNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeCustomEscapeSequence_RPCError tests SubscribeCustomEscapeSequence function returns error when iTerm2 returns an RPC error.
func TestSubscribeCustomEscapeSequence_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeCustomEscapeSequence(ctx, mc, c, func(caller Caller, n *iterm2.CustomEscapeSequenceNotification) {}, "s1")
	require.Error(t, err)
}

// TestSubscribeVariableChange_RPCError tests SubscribeVariableChange function returns error when iTerm2 returns an RPC error.
func TestSubscribeVariableChange_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeVariableChange(ctx, mc, c, func(caller Caller, n *iterm2.VariableChangedNotification) {}, "s1", "jobName")
	require.Error(t, err)
}

// TestSubscribeLayoutChange_RPCError tests SubscribeLayoutChange function returns error when iTerm2 returns an RPC error.
func TestSubscribeLayoutChange_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeLayoutChange(ctx, mc, c, func(caller Caller, n *iterm2.LayoutChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeFocusChange_RPCError tests SubscribeFocusChange function returns error when iTerm2 returns an RPC error.
func TestSubscribeFocusChange_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeFocusChange(ctx, mc, c, func(caller Caller, n *iterm2.FocusChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeServerOriginatedRPC_RPCError tests SubscribeServerOriginatedRPC function returns error when iTerm2 returns an RPC error.
func TestSubscribeServerOriginatedRPC_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeServerOriginatedRPC(ctx, mc, c, func(caller Caller, n *iterm2.ServerOriginatedRPCNotification) {})
	require.Error(t, err)
}

// TestSubscribeBroadcastChange_RPCError tests SubscribeBroadcastChange function returns error when iTerm2 returns an RPC error.
func TestSubscribeBroadcastChange_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeBroadcastChange(ctx, mc, c, func(caller Caller, n *iterm2.BroadcastDomainsChangedNotification) {})
	require.Error(t, err)
}

// TestSubscribeProfileChange_RPCError tests SubscribeProfileChange function returns error when iTerm2 returns an RPC error.
func TestSubscribeProfileChange_RPCError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: errorResp("rpc err")}
	c := newTestConnWithCaller(mc)
	_, err := SubscribeProfileChange(ctx, mc, c, func(caller Caller, n *iterm2.ProfileChangedNotification) {})
	require.Error(t, err)
}

// TestDispatch_CallsHandler tests Dispatch correctly invokes handler.
func TestDispatch_CallsHandler(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var receivedNewSession *iterm2.NewSessionNotification
	tk, err := SubscribeNewSession(ctx, mc, c, func(caller Caller, n *iterm2.NewSessionNotification) {
		mu.Lock()
		receivedNewSession = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				NewSessionNotification: &iterm2.NewSessionNotification{
					SessionId: proto.String("s-new"),
				},
			},
		},
	}

	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, receivedNewSession)
	assert.Equal(t, "s-new", receivedNewSession.GetSessionId())
	mu.Unlock()
}

// TestDispatch_KeystrokeNotification tests Dispatch method correctly delivers keystroke notifications to handlers.
func TestDispatch_KeystrokeNotification(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.KeystrokeNotification
	tk, err := SubscribeKeystroke(ctx, mc, c, func(caller Caller, n *iterm2.KeystrokeNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	}, "s1")
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				KeystrokeNotification: &iterm2.KeystrokeNotification{
					Session: proto.String("s1"),
				},
			},
		},
	}

	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_TerminateSession tests Dispatch method correctly delivers terminate session notifications to handlers.
func TestDispatch_TerminateSession(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.TerminateSessionNotification
	tk, err := SubscribeTerminateSession(ctx, mc, c, func(caller Caller, n *iterm2.TerminateSessionNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				TerminateSessionNotification: &iterm2.TerminateSessionNotification{
					SessionId: proto.String("s-old"),
				},
			},
		},
	}

	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_NilNotification tests Dispatch method handles nil notification gracefully without panic.
func TestDispatch_NilNotification(t *testing.T) {
	c := NewConnection("", "", "test")
	msg := &iterm2.ServerOriginatedMessage{}
	c.Dispatch(msg) // should not panic
}

// TestDispatch_NoSubmessage tests Dispatch method handles notification without submessage type gracefully.
func TestDispatch_NoSubmessage(t *testing.T) {
	c := NewConnection("", "", "test")
	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{},
		},
	}
	c.Dispatch(msg) // no matching submessage → should not panic
}

// TestDispatch_ScreenUpdate tests Dispatch method correctly delivers screen update notifications to handlers.
func TestDispatch_ScreenUpdate(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.ScreenUpdateNotification
	tk, err := SubscribeScreenUpdate(ctx, mc, c, func(caller Caller, n *iterm2.ScreenUpdateNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	}, "s1")
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				ScreenUpdateNotification: &iterm2.ScreenUpdateNotification{Session: proto.String("s1")},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_Prompt tests Dispatch method correctly delivers prompt notifications to handlers.
func TestDispatch_Prompt(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.PromptNotification
	tk, err := SubscribePrompt(ctx, mc, c, func(caller Caller, n *iterm2.PromptNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	}, "s1")
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				PromptNotification: &iterm2.PromptNotification{Session: proto.String("s1")},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_CustomEscapeSequence tests Dispatch method correctly delivers custom escape sequence notifications to handlers.
func TestDispatch_CustomEscapeSequence(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.CustomEscapeSequenceNotification
	tk, err := SubscribeCustomEscapeSequence(ctx, mc, c, func(caller Caller, n *iterm2.CustomEscapeSequenceNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	}, "s1")
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				CustomEscapeSequenceNotification: &iterm2.CustomEscapeSequenceNotification{Session: proto.String("s1")},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_LayoutChange tests Dispatch method correctly delivers layout change notifications to handlers.
func TestDispatch_LayoutChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.LayoutChangedNotification
	tk, err := SubscribeLayoutChange(ctx, mc, c, func(caller Caller, n *iterm2.LayoutChangedNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				LayoutChangedNotification: &iterm2.LayoutChangedNotification{},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_FocusChange tests Dispatch method correctly delivers focus change notifications to handlers.
func TestDispatch_FocusChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.FocusChangedNotification
	tk, err := SubscribeFocusChange(ctx, mc, c, func(caller Caller, n *iterm2.FocusChangedNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				FocusChangedNotification: &iterm2.FocusChangedNotification{},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_ServerOriginatedRPC tests Dispatch method correctly delivers server originated RPC notifications to handlers.
func TestDispatch_ServerOriginatedRPC(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.ServerOriginatedRPCNotification
	tk, err := SubscribeServerOriginatedRPC(ctx, mc, c, func(caller Caller, n *iterm2.ServerOriginatedRPCNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				ServerOriginatedRpcNotification: &iterm2.ServerOriginatedRPCNotification{},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_BroadcastChange tests Dispatch method correctly delivers broadcast change notifications to handlers.
func TestDispatch_BroadcastChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.BroadcastDomainsChangedNotification
	tk, err := SubscribeBroadcastChange(ctx, mc, c, func(caller Caller, n *iterm2.BroadcastDomainsChangedNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				BroadcastDomainsChanged: &iterm2.BroadcastDomainsChangedNotification{},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_VariableChange tests Dispatch method correctly delivers variable change notifications to handlers.
func TestDispatch_VariableChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.VariableChangedNotification
	tk, err := SubscribeVariableChange(ctx, mc, c, func(caller Caller, n *iterm2.VariableChangedNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	}, "s1", "var1")
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				VariableChangedNotification: &iterm2.VariableChangedNotification{
					Scope:      iterm2.VariableScope_SESSION.Enum(),
					Identifier: proto.String("s1"),
					Name:       proto.String("var1"),
				},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// TestDispatch_ProfileChange tests Dispatch method correctly delivers profile change notifications to handlers.
func TestDispatch_ProfileChange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: makeNotifyResp()}
	c := newTestConnWithCaller(mc)

	var mu sync.Mutex
	var received *iterm2.ProfileChangedNotification
	tk, err := SubscribeProfileChange(ctx, mc, c, func(caller Caller, n *iterm2.ProfileChangedNotification) {
		mu.Lock()
		received = n
		mu.Unlock()
	})
	require.NoError(t, err)
	defer c.Unsubscribe(tk)

	msg := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				ProfileChangedNotification: &iterm2.ProfileChangedNotification{},
			},
		},
	}
	c.Dispatch(msg)

	mu.Lock()
	require.NotNil(t, received)
	mu.Unlock()
}

// notificationKeys (previously 0%)

// TestNotificationKeys_NewSession tests notificationKeys function for new session notification.
func TestNotificationKeys_NewSession(t *testing.T) {
	n := &iterm2.Notification{
		NewSessionNotification: &iterm2.NewSessionNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"new_session"}, keys)
}

// TestNotificationKeys_TerminateSession tests notificationKeys function for terminate session notification.
func TestNotificationKeys_TerminateSession(t *testing.T) {
	n := &iterm2.Notification{
		TerminateSessionNotification: &iterm2.TerminateSessionNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"terminate_session"}, keys)
}

// TestNotificationKeys_Keystroke tests notificationKeys function for keystroke notification with session scope.
func TestNotificationKeys_Keystroke(t *testing.T) {
	n := &iterm2.Notification{
		KeystrokeNotification: &iterm2.KeystrokeNotification{
			Session: proto.String("s1"),
		},
	}
	keys := notificationKeys(n)
	require.Len(t, keys, 2)
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	assert.True(t, found["keystroke:s1"])
	assert.True(t, found["keystroke:"])
}

// TestNotificationKeys_ScreenUpdate tests notificationKeys function for screen update notification with session scope.
func TestNotificationKeys_ScreenUpdate(t *testing.T) {
	n := &iterm2.Notification{
		ScreenUpdateNotification: &iterm2.ScreenUpdateNotification{
			Session: proto.String("s1"),
		},
	}
	keys := notificationKeys(n)
	require.Len(t, keys, 2)
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	assert.True(t, found["screen_update:s1"])
	assert.True(t, found["screen_update:"])
}

// TestNotificationKeys_Prompt tests notificationKeys function for prompt notification with session scope.
func TestNotificationKeys_Prompt(t *testing.T) {
	n := &iterm2.Notification{
		PromptNotification: &iterm2.PromptNotification{
			Session: proto.String("s1"),
		},
	}
	keys := notificationKeys(n)
	require.Len(t, keys, 2)
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	assert.True(t, found["prompt:s1"])
	assert.True(t, found["prompt:"])
}

// TestNotificationKeys_CustomEscapeSequence tests notificationKeys function for custom escape sequence notification with session scope.
func TestNotificationKeys_CustomEscapeSequence(t *testing.T) {
	n := &iterm2.Notification{
		CustomEscapeSequenceNotification: &iterm2.CustomEscapeSequenceNotification{
			Session: proto.String("s1"),
		},
	}
	keys := notificationKeys(n)
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	assert.True(t, found["custom_escape:s1"])
	assert.True(t, found["custom_escape:"])
}

// TestNotificationKeys_LayoutChange tests notificationKeys function for layout change notification.
func TestNotificationKeys_LayoutChange(t *testing.T) {
	n := &iterm2.Notification{
		LayoutChangedNotification: &iterm2.LayoutChangedNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"layout_change"}, keys)
}

// TestNotificationKeys_FocusChange tests notificationKeys function for focus change notification.
func TestNotificationKeys_FocusChange(t *testing.T) {
	n := &iterm2.Notification{
		FocusChangedNotification: &iterm2.FocusChangedNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"focus_change"}, keys)
}

// TestNotificationKeys_ServerOriginatedRPC tests notificationKeys function for server originated RPC notification.
func TestNotificationKeys_ServerOriginatedRPC(t *testing.T) {
	n := &iterm2.Notification{
		ServerOriginatedRpcNotification: &iterm2.ServerOriginatedRPCNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"server_originated_rpc:"}, keys)
}

// TestNotificationKeys_BroadcastChange tests notificationKeys function for broadcast change notification.
func TestNotificationKeys_BroadcastChange(t *testing.T) {
	n := &iterm2.Notification{
		BroadcastDomainsChanged: &iterm2.BroadcastDomainsChangedNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"broadcast_change"}, keys)
}

// TestNotificationKeys_VariableChange tests notificationKeys function for variable change notification with scope.
func TestNotificationKeys_VariableChange(t *testing.T) {
	n := &iterm2.Notification{
		VariableChangedNotification: &iterm2.VariableChangedNotification{
			Scope:      iterm2.VariableScope_SESSION.Enum(),
			Identifier: proto.String("s1"),
			Name:       proto.String("jobName"),
		},
	}
	keys := notificationKeys(n)
	require.Len(t, keys, 2)
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	assert.True(t, found["variable_change:1:s1:jobName:"])
	assert.True(t, found["variable_change:1::jobName:"])
}

// TestNotificationKeys_ProfileChange tests notificationKeys function for profile change notification.
func TestNotificationKeys_ProfileChange(t *testing.T) {
	n := &iterm2.Notification{
		ProfileChangedNotification: &iterm2.ProfileChangedNotification{},
	}
	keys := notificationKeys(n)
	assert.Equal(t, []string{"profile_change:"}, keys)
}
