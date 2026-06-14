package term2go

import (
	"errors"
	"testing"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestPrompt_Nil tests Prompt behavior with nil proto.
func TestPrompt_Nil(t *testing.T) {
	p := NewPrompt(nil)
	assert.Equal(t, int32(0), p.PromptRange().Start.X)
	assert.Empty(t, p.WorkingDirectory())
	assert.Nil(t, p.Raw())
}

// TestPrompt_Basic tests NewPrompt function for creating a Prompt from GetPromptResponse.
func TestPrompt_Basic(t *testing.T) {
	raw := &iterm2.GetPromptResponse{
		PromptRange: &iterm2.CoordRange{
			Start: &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(0)},
			End:   &iterm2.Coord{X: proto.Int32(5), Y: proto.Int64(0)},
		},
		WorkingDirectory: proto.String("/home"),
		Command:          proto.String("ls -la"),
		PromptState:      iterm2.GetPromptResponse_FINISHED.Enum(),
		ExitStatus:       proto.Uint32(0),
		UniquePromptId:   proto.String("pid-1"),
	}
	p := NewPrompt(raw)

	assert.Equal(t, "/home", p.WorkingDirectory())
	assert.Equal(t, "ls -la", p.Command())
	assert.Equal(t, PromptFinished, p.State())
	assert.Equal(t, uint32(0), p.ExitStatus())
	assert.Equal(t, "pid-1", p.UniqueID())
}

// TestPrompt_Ranges tests NewPrompt function for extracting range information from GetPromptResponse.
func TestPrompt_Ranges(t *testing.T) {
	raw := &iterm2.GetPromptResponse{
		PromptRange: &iterm2.CoordRange{
			Start: &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(10)},
			End:   &iterm2.Coord{X: proto.Int32(5), Y: proto.Int64(10)},
		},
		CommandRange: &iterm2.CoordRange{
			Start: &iterm2.Coord{X: proto.Int32(4), Y: proto.Int64(10)},
			End:   &iterm2.Coord{X: proto.Int32(10), Y: proto.Int64(10)},
		},
		OutputRange: &iterm2.CoordRange{
			Start: &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(11)},
			End:   &iterm2.Coord{X: proto.Int32(0), Y: proto.Int64(20)},
		},
	}
	p := NewPrompt(raw)

	pr := p.PromptRange()
	assert.Equal(t, int32(0), pr.Start.X)
	assert.Equal(t, int32(10), pr.Start.Y)

	cr := p.CommandRange()
	assert.Equal(t, int32(10), cr.End.X)

	or := p.OutputRange()
	assert.Equal(t, int32(20), or.End.Y)
}

// TestPrompt_ExcludedSubranges tests NewPrompt function for extracting excluded subrange information.
func TestPrompt_ExcludedSubranges(t *testing.T) {
	raw := &iterm2.GetPromptResponse{
		ExcludedSubranges: []*iterm2.CoordRange{
			{
				Start: &iterm2.Coord{X: proto.Int32(1), Y: proto.Int64(0)},
				End:   &iterm2.Coord{X: proto.Int32(3), Y: proto.Int64(0)},
			},
		},
	}
	p := NewPrompt(raw)
	ranges := p.ExcludedSubranges()
	require.Len(t, ranges, 1)
	assert.Equal(t, int32(1), ranges[0].Start.X)
}

// TestPromptState_Values tests PromptState constants have correct values.
func TestPromptState_Values(t *testing.T) {
	assert.Equal(t, 0, int(PromptEditing))
	assert.Equal(t, 1, int(PromptRunning))
	assert.Equal(t, 2, int(PromptFinished))
}

// TestGetLastPrompt_Success tests GetLastPrompt function success return.
func TestGetLastPrompt_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_GetPromptResponse{
			GetPromptResponse: &iterm2.GetPromptResponse{
				Command:        proto.String("echo hi"),
				UniquePromptId: proto.String("pid"),
				PromptState:    iterm2.GetPromptResponse_FINISHED.Enum(),
			},
		},
	}}
	p, err := GetLastPrompt(ctx, mc, "s1")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "echo hi", p.Command())
}

// TestGetLastPrompt_Unavailable tests GetLastPrompt function returns nil when prompt is unavailable.
func TestGetLastPrompt_Unavailable(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_GetPromptResponse{
			GetPromptResponse: &iterm2.GetPromptResponse{
				Status: iterm2.GetPromptResponse_PROMPT_UNAVAILABLE.Enum(),
			},
		},
	}}
	p, err := GetLastPrompt(ctx, mc, "s1")
	require.NoError(t, err)
	assert.Nil(t, p)
}

// TestGetLastPrompt_Error tests GetLastPrompt function when the caller returns an error.
func TestGetLastPrompt_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{err: errors.New("conn error")}
	_, err := GetLastPrompt(ctx, mc, "s1")
	require.Error(t, err)
}

// TestGetPromptByID_Success tests GetPromptByID function for retrieving a specific prompt by ID.
func TestGetPromptByID_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_GetPromptResponse{
			GetPromptResponse: &iterm2.GetPromptResponse{
				UniquePromptId: proto.String("target-id"),
				Command:        proto.String("specific"),
			},
		},
	}}
	p, err := GetPromptByID(ctx, mc, "s1", "target-id")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "target-id", p.UniqueID())
	// Verify RPC was sent with correct ID.
	req := mc.req.GetGetPromptRequest()
	assert.Equal(t, "target-id", req.GetUniquePromptId())
}

// TestListPromptIDs_Success tests ListPromptIDs function for retrieving all prompt IDs.
func TestListPromptIDs_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_ListPromptsResponse{
			ListPromptsResponse: &iterm2.ListPromptsResponse{
				UniquePromptId: []string{"a", "b", "c"},
			},
		},
	}}
	ids, err := ListPromptIDs(ctx, mc, "s1", "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, ids)
}

// TestListPromptIDs_WithRange tests ListPromptIDs function with first and last unique ID range.
func TestListPromptIDs_WithRange(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_ListPromptsResponse{
			ListPromptsResponse: &iterm2.ListPromptsResponse{},
		},
	}}
	_, err := ListPromptIDs(ctx, mc, "s1", "first", "last")
	require.NoError(t, err)
	req := mc.req.GetListPromptsRequest()
	assert.Equal(t, "first", req.GetFirstUniqueId())
	assert.Equal(t, "last", req.GetLastUniqueId())
}

// TestPromptEvent_Defaults tests PromptEvent default values.
func TestPromptEvent_Defaults(t *testing.T) {
	ev := PromptEvent{}
	assert.Zero(t, ev.Mode)
	assert.Nil(t, ev.Prompt)
	assert.Empty(t, ev.Command)
}

// TestPromptMonitor_NewAndClose tests PromptMonitor creation and closing.
func TestPromptMonitor_NewAndClose(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Subscribe RPC
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)
		resp := &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
				NotificationResponse: &iterm2.NotificationResponse{
					Status: iterm2.NotificationResponse_OK.Enum(),
				},
			},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}

		// Unsubscribe RPC (from pm.Close())
		data = <-mock.writeCh
		_ = proto.Unmarshal(data, &req)
		resp.Id = req.Id
		b, _ = proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	pm.Close()
	pm.Close()
	<-done
}

// TestPromptMonitor_DefaultModes tests PromptMonitor uses default modes when no options are provided.
func TestPromptMonitor_DefaultModes(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)
		// Verify the subscribe request includes PROMPT mode
		nr := req.GetNotificationRequest()
		require.NotNil(t, nr)
		resp := &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
				NotificationResponse: &iterm2.NotificationResponse{
					Status: iterm2.NotificationResponse_OK.Enum(),
				},
			},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}

		// Unsubscribe
		data = <-mock.writeCh
		_ = proto.Unmarshal(data, &req)
		resp.Id = req.Id
		b, _ = proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	pm.Close()
}

// TestPromptMonitor_ReceivesPromptEvent tests PromptMonitor receives prompt events from iTerm2.
func TestPromptMonitor_ReceivesPromptEvent(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ { // subscribe + unsubscribe
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{
						Status: iterm2.NotificationResponse_OK.Enum(),
					},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	defer pm.Close()

	// Send a prompt notification through Dispatch.
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				PromptNotification: &iterm2.PromptNotification{
					Session: proto.String("s1"),
					Event: &iterm2.PromptNotification_Prompt{
						Prompt: &iterm2.PromptNotificationPrompt{
							Prompt: &iterm2.GetPromptResponse{
								Command: proto.String("ls"),
							},
						},
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	ev := <-pm.Chan()
	assert.Equal(t, iterm2.PromptMonitorMode_PROMPT, ev.Mode)
	require.NotNil(t, ev.Prompt)
	assert.Equal(t, "ls", ev.Prompt.Command())
}

// TestPromptMonitor_ReceivesCommandStart tests PromptMonitor receives command start events.
func TestPromptMonitor_ReceivesCommandStart(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ { // subscribe + unsubscribe
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{
						Status: iterm2.NotificationResponse_OK.Enum(),
					},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	defer pm.Close()

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				PromptNotification: &iterm2.PromptNotification{
					Session: proto.String("s1"),
					Event: &iterm2.PromptNotification_CommandStart{
						CommandStart: &iterm2.PromptNotificationCommandStart{
							Command: proto.String("make build"),
						},
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	ev := <-pm.Chan()
	assert.Equal(t, iterm2.PromptMonitorMode_COMMAND_START, ev.Mode)
	assert.Equal(t, "make build", ev.Command)
}

// TestPromptMonitor_ReceivesCommandEnd tests PromptMonitor receives command end events.
func TestPromptMonitor_ReceivesCommandEnd(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ { // subscribe + unsubscribe
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{
						Status: iterm2.NotificationResponse_OK.Enum(),
					},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	defer pm.Close()

	exitStatus := int32(42)
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				PromptNotification: &iterm2.PromptNotification{
					Session: proto.String("s1"),
					Event: &iterm2.PromptNotification_CommandEnd{
						CommandEnd: &iterm2.PromptNotificationCommandEnd{
							Status: &exitStatus,
						},
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	ev := <-pm.Chan()
	assert.Equal(t, iterm2.PromptMonitorMode_COMMAND_END, ev.Mode)
	assert.Equal(t, int32(42), ev.Status)
}

// TestPromptMonitor_NonPromptNotification tests PromptMonitor ignores non-prompt notifications.
func TestPromptMonitor_NonPromptNotification(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ { // subscribe + unsubscribe
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{
						Status: iterm2.NotificationResponse_OK.Enum(),
					},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	pm, err := NewPromptMonitor(conn, "s1", nil)
	require.NoError(t, err)
	defer pm.Close()

	// Send a non-prompt notification → should not appear in channel.
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				NewSessionNotification: &iterm2.NewSessionNotification{},
			},
		},
	}
	conn.Dispatch(notif)

	// Channel should still be open with no prompt event.
}

// TestPromptMonitor_SubscribeError tests PromptMonitor handles subscription error from server.
func TestPromptMonitor_SubscribeError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)
		resp := &iterm2.ServerOriginatedMessage{
			Id:         req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_Error{Error: "subscribe failed"},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	_, err := NewPromptMonitor(conn, "s1", nil)
	require.Error(t, err)
}
