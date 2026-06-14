package term2go

import (
	"testing"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestWindowStatus_Values tests WindowStatus constant values.
func TestWindowStatus_Values(t *testing.T) {
	assert.Equal(t, 0, int(WindowBecameKey))
	assert.Equal(t, 1, int(WindowIsCurrent))
	assert.Equal(t, 2, int(WindowResignedKey))
}

// TestWindowFocusChange_Basic tests WindowFocusChange struct basic fields.
func TestWindowFocusChange_Basic(t *testing.T) {
	w := &WindowFocusChange{WindowID: "w1", Status: WindowBecameKey}
	assert.Equal(t, "w1", w.WindowID)
	assert.Equal(t, WindowBecameKey, w.Status)
}

// TestFocusUpdateFromProto_ApplicationActive tests proto conversion for application active state.
func TestFocusUpdateFromProto_ApplicationActive(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_ApplicationActive{
			ApplicationActive: true,
		},
	}
	u := focusUpdateFromProto(n)
	require.NotNil(t, u)
	require.NotNil(t, u.ApplicationActive)
	assert.True(t, *u.ApplicationActive)
}

// TestFocusUpdateFromProto_ApplicationInactive tests focusUpdateFromProto for application inactive state.
func TestFocusUpdateFromProto_ApplicationInactive(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_ApplicationActive{
			ApplicationActive: false,
		},
	}
	u := focusUpdateFromProto(n)
	require.NotNil(t, u)
	require.NotNil(t, u.ApplicationActive)
	assert.False(t, *u.ApplicationActive)
}

// TestFocusUpdateFromProto_WindowChanged tests focusUpdateFromProto for window focus change.
func TestFocusUpdateFromProto_WindowChanged(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_Window_{
			Window: &iterm2.FocusChangedNotification_Window{
				WindowId:     proto.String("win-1"),
				WindowStatus: iterm2.FocusChangedNotification_Window_TERMINAL_WINDOW_IS_CURRENT.Enum(),
			},
		},
	}
	u := focusUpdateFromProto(n)
	require.NotNil(t, u)
	require.NotNil(t, u.WindowChanged)
	assert.Equal(t, "win-1", u.WindowChanged.WindowID)
	assert.Equal(t, WindowIsCurrent, u.WindowChanged.Status)
}

// TestFocusUpdateFromProto_WindowNil tests focusUpdateFromProto returns nil when window data is nil.
func TestFocusUpdateFromProto_WindowNil(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_Window_{
			Window: nil,
		},
	}
	u := focusUpdateFromProto(n)
	assert.Nil(t, u)
}

// TestFocusUpdateFromProto_SelectedTab tests focusUpdateFromProto for selected tab change.
func TestFocusUpdateFromProto_SelectedTab(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_SelectedTab{
			SelectedTab: "tab-1",
		},
	}
	u := focusUpdateFromProto(n)
	require.NotNil(t, u)
	require.NotNil(t, u.SelectedTab)
	assert.Equal(t, "tab-1", *u.SelectedTab)
}

// TestFocusUpdateFromProto_ActiveSession tests focusUpdateFromProto for active session change.
func TestFocusUpdateFromProto_ActiveSession(t *testing.T) {
	n := &iterm2.FocusChangedNotification{
		Event: &iterm2.FocusChangedNotification_Session{
			Session: "sess-1",
		},
	}
	u := focusUpdateFromProto(n)
	require.NotNil(t, u)
	require.NotNil(t, u.ActiveSession)
	assert.Equal(t, "sess-1", *u.ActiveSession)
}

// TestFocusUpdateFromProto_Empty tests focusUpdateFromProto returns nil for empty notification.
func TestFocusUpdateFromProto_Empty(t *testing.T) {
	u := focusUpdateFromProto(&iterm2.FocusChangedNotification{})
	assert.Nil(t, u)
}

// TestFocusMonitor_NewAndClose tests FocusMonitor creation and closing.
func TestFocusMonitor_NewAndClose(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	done := make(chan struct{})
	go func() {
		defer close(done)
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

	fm, err := NewFocusMonitor(conn)
	require.NoError(t, err)
	fm.Close()
	fm.Close()
	<-done
}

// TestFocusMonitor_ReceivesAppEvent tests FocusMonitor receives application focus change events.
func TestFocusMonitor_ReceivesAppEvent(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
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

	fm, err := NewFocusMonitor(conn)
	require.NoError(t, err)
	defer fm.Close()

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				FocusChangedNotification: &iterm2.FocusChangedNotification{
					Event: &iterm2.FocusChangedNotification_ApplicationActive{
						ApplicationActive: true,
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	u := <-fm.Chan()
	require.NotNil(t, u.ApplicationActive)
	assert.True(t, *u.ApplicationActive)
}

// TestFocusMonitor_ReceivesWindowEvent tests FocusMonitor receives window focus change events.
func TestFocusMonitor_ReceivesWindowEvent(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
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

	fm, err := NewFocusMonitor(conn)
	require.NoError(t, err)
	defer fm.Close()

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				FocusChangedNotification: &iterm2.FocusChangedNotification{
					Event: &iterm2.FocusChangedNotification_Window_{
						Window: &iterm2.FocusChangedNotification_Window{
							WindowId:     proto.String("win-abc"),
							WindowStatus: iterm2.FocusChangedNotification_Window_TERMINAL_WINDOW_BECAME_KEY.Enum(),
						},
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	u := <-fm.Chan()
	require.NotNil(t, u.WindowChanged)
	assert.Equal(t, "win-abc", u.WindowChanged.WindowID)
}

// TestFocusMonitor_ReceivesTabEvent tests FocusMonitor receives tab selection change events.
func TestFocusMonitor_ReceivesTabEvent(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
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

	fm, err := NewFocusMonitor(conn)
	require.NoError(t, err)
	defer fm.Close()

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				FocusChangedNotification: &iterm2.FocusChangedNotification{
					Event: &iterm2.FocusChangedNotification_SelectedTab{
						SelectedTab: "tab-xyz",
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	u := <-fm.Chan()
	require.NotNil(t, u.SelectedTab)
	assert.Equal(t, "tab-xyz", *u.SelectedTab)
}

// TestFocusMonitor_ReceivesSessionEvent tests FocusMonitor receives active session change events.
func TestFocusMonitor_ReceivesSessionEvent(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
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

	fm, err := NewFocusMonitor(conn)
	require.NoError(t, err)
	defer fm.Close()

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				FocusChangedNotification: &iterm2.FocusChangedNotification{
					Event: &iterm2.FocusChangedNotification_Session{
						Session: "sess-xyz",
					},
				},
			},
		},
	}
	conn.Dispatch(notif)

	u := <-fm.Chan()
	require.NotNil(t, u.ActiveSession)
	assert.Equal(t, "sess-xyz", *u.ActiveSession)
}

// TestFocusMonitor_SubscribeError tests FocusMonitor handles subscription error from server.
func TestFocusMonitor_SubscribeError(t *testing.T) {
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
			Submessage: &iterm2.ServerOriginatedMessage_Error{Error: "fail"},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	_, err := NewFocusMonitor(conn)
	require.Error(t, err)
}
