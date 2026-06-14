package term2go

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestKeystrokeEvent_Basic tests KeystrokeEvent basic properties.
func TestKeystrokeEvent_Basic(t *testing.T) {
	raw := &iterm2.KeystrokeNotification{
		Characters:                  proto.String("a"),
		CharactersIgnoringModifiers: proto.String("a"),
		Modifiers:                   []iterm2.Modifiers{iterm2.Modifiers_SHIFT},
		KeyCode:                     proto.Int32(0),
		Session:                     proto.String("s1"),
		Action:                      iterm2.KeystrokeNotification_KEY_DOWN.Enum(),
	}
	ev := newKeystrokeEvent(raw)

	assert.Equal(t, "a", ev.Characters())
	assert.Equal(t, "a", ev.CharactersIgnoringModifiers())
	assert.Equal(t, []iterm2.Modifiers{iterm2.Modifiers_SHIFT}, ev.Modifiers())
	assert.Equal(t, int32(0), ev.KeyCode())
	assert.Equal(t, "s1", ev.Session())
	assert.Equal(t, KeystrokeKeyDown, ev.Action())
	assert.Equal(t, raw, ev.Raw())
}

// TestKeystrokeAction_Values tests KeystrokeAction constants have correct values.
func TestKeystrokeAction_Values(t *testing.T) {
	assert.Equal(t, KeystrokeAction(0), KeystrokeKeyDown)
	assert.Equal(t, KeystrokeAction(1), KeystrokeKeyUp)
	assert.Equal(t, KeystrokeAction(2), KeystrokeFlagsChanged)
}

// TestKeystrokeMonitor_NewAndClose tests KeystrokeMonitor creation and closing.
func TestKeystrokeMonitor_NewAndClose(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
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
			mock.readCh <- mockReadResult{msgType: 1, data: b}
		}
	}()

	km, err := NewKeystrokeMonitor(conn, "s1", false)
	require.NoError(t, err)
	km.Close()
	km.Close()
	<-done
}

// TestKeystrokeMonitor_ReceivesEvent tests KeystrokeMonitor receives keystroke events.
func TestKeystrokeMonitor_ReceivesEvent(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{Status: iterm2.NotificationResponse_OK.Enum()},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: 1, data: b}
		}
	}()

	km, err := NewKeystrokeMonitor(conn, "s1", false)
	require.NoError(t, err)
	defer km.Close()

	sessionID := "s1"
	keyCode := int32(97)
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				KeystrokeNotification: &iterm2.KeystrokeNotification{
					Characters: proto.String("a"),
					Session:    &sessionID,
					KeyCode:    &keyCode,
					Action:     iterm2.KeystrokeNotification_KEY_DOWN.Enum(),
				},
			},
		},
	}
	conn.Dispatch(notif)

	ev := <-km.Chan()
	assert.Equal(t, "a", ev.Characters())
	assert.Equal(t, int32(97), ev.KeyCode())
}

// TestKeystrokeMonitor_Advanced tests KeystrokeMonitor with IncludeUpEvents option.
func TestKeystrokeMonitor_Advanced(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 2; i++ {
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{Status: iterm2.NotificationResponse_OK.Enum()},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: 1, data: b}
		}
	}()

	km, err := NewKeystrokeMonitor(conn, "s1", true)
	require.NoError(t, err)
	defer km.Close()

	sessionID := "s1"
	code := int32(97)
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				KeystrokeNotification: &iterm2.KeystrokeNotification{
					Characters: proto.String("a"),
					Session:    &sessionID,
					KeyCode:    &code,
					Action:     iterm2.KeystrokeNotification_KEY_UP.Enum(),
				},
			},
		},
	}
	conn.Dispatch(notif)

	ev := <-km.Chan()
	assert.Equal(t, KeystrokeKeyUp, ev.Action())
}

// TestKeystrokeMonitor_SubscribeError tests KeystrokeMonitor handles subscription error from server.
func TestKeystrokeMonitor_SubscribeError(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
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
		mock.readCh <- mockReadResult{msgType: 1, data: b}
	}()

	_, err := NewKeystrokeMonitor(conn, "s1", false)
	require.Error(t, err)
}

// TestKeystrokeFilter_NewAndClose tests KeystrokeFilter creation and closing.
func TestKeystrokeFilter_NewAndClose(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
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
					NotificationResponse: &iterm2.NotificationResponse{Status: iterm2.NotificationResponse_OK.Enum()},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: 1, data: b}
		}
	}()

	patterns := []*iterm2.KeystrokePattern{
		{Characters: []string{"a"}},
	}
	kf, err := NewKeystrokeFilter(conn, "", patterns)
	require.NoError(t, err)
	kf.Close()
	kf.Close()
	<-done
}

// TestKeystrokeFilter_SubscribeError tests KeystrokeFilter handles subscription error from server.
func TestKeystrokeFilter_SubscribeError(t *testing.T) {
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	ctx, cancel := testCtx()
	defer cancel()
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
		mock.readCh <- mockReadResult{msgType: 1, data: b}
	}()

	_, err := NewKeystrokeFilter(conn, "", nil)
	require.Error(t, err)
}
