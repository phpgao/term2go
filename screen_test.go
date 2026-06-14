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

// TestScreenContents_NilProto tests ScreenContents behavior with nil proto.
func TestScreenContents_NilProto(t *testing.T) {
	sc := NewScreenContents(nil)
	assert.Equal(t, 0, sc.LineCount())
	assert.Nil(t, sc.Cursor())
	assert.Nil(t, sc.Lines())
}

// TestScreenContents_Empty tests NewScreenContents function with empty response.
func TestScreenContents_Empty(t *testing.T) {
	sc := NewScreenContents(&iterm2.GetBufferResponse{})
	assert.Equal(t, 0, sc.LineCount())
	assert.Nil(t, sc.Cursor())
}

// TestScreenContents_Basic tests NewScreenContents function with basic content.
func TestScreenContents_Basic(t *testing.T) {
	raw := &iterm2.GetBufferResponse{
		Cursor: &iterm2.Coord{X: proto.Int32(5), Y: proto.Int64(10)},
		Contents: []*iterm2.LineContents{
			{
				Text: proto.String("hello"),
				CodePointsPerCell: []*iterm2.CodePointsPerCell{
					{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(3)},
					{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(2)},
				},
				Continuation: iterm2.LineContents_CONTINUATION_HARD_EOL.Enum(),
			},
			{
				Text: proto.String("world"),
				CodePointsPerCell: []*iterm2.CodePointsPerCell{
					{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(5)},
				},
				Continuation: iterm2.LineContents_CONTINUATION_SOFT_EOL.Enum(),
			},
		},
	}
	sc := NewScreenContents(raw)

	assert.Equal(t, 2, sc.LineCount())

	c := sc.Cursor()
	require.NotNil(t, c)
	assert.Equal(t, int32(5), c.X)
	assert.Equal(t, int32(10), c.Y)

	lines := sc.Lines()
	require.Len(t, lines, 2)

	assert.Equal(t, "hello", lines[0].Text())
	assert.Equal(t, "world", lines[1].Text())
}

// TestScreenContents_Raw tests ScreenContents.Raw method for retrieving raw response.
func TestScreenContents_Raw(t *testing.T) {
	raw := &iterm2.GetBufferResponse{}
	sc := NewScreenContents(raw)
	assert.Same(t, raw, sc.Raw())
}

// TestLineContent_Text tests LineContent.Text method.
func TestLineContent_Text(t *testing.T) {
	raw := &iterm2.LineContents{
		Text: proto.String("abc"),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(3)},
		},
	}
	lc := newLineContent(raw)
	require.NotNil(t, lc)
	assert.Equal(t, "abc", lc.Text())
	assert.Equal(t, 3, lc.Len())
}

// TestLineContent_RuneAt tests LineContent.RuneAt method for retrieving rune at position.
func TestLineContent_RuneAt(t *testing.T) {
	raw := &iterm2.LineContents{
		Text: proto.String("a世c"),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)}, // 'a' = 1 byte
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)}, // '世' = 3 bytes
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)}, // 'c' = 1 byte
		},
	}
	lc := newLineContent(raw)
	require.Equal(t, 3, lc.Len())

	r, s := lc.RuneAt(0)
	assert.Equal(t, 'a', r)
	assert.Equal(t, 1, s)

	r, s = lc.RuneAt(1)
	assert.Equal(t, '世', r)
	assert.Equal(t, 3, s)

	r, s = lc.RuneAt(2)
	assert.Equal(t, 'c', r)
	assert.Equal(t, 1, s)
}

// TestLineContent_RuneAt_OutOfBounds tests LineContent.RuneAt method with out of bounds position.
func TestLineContent_RuneAt_OutOfBounds(t *testing.T) {
	raw := &iterm2.LineContents{
		Text: proto.String("x"),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)},
		},
	}
	lc := newLineContent(raw)

	r, s := lc.RuneAt(-1)
	assert.Zero(t, r)
	assert.Zero(t, s)
	r, s = lc.RuneAt(5)
	assert.Zero(t, r)
	assert.Zero(t, s)
}

// TestLineContent_HardEOL tests LineContent.HardEOL method for detecting hard end of line.
func TestLineContent_HardEOL(t *testing.T) {
	// Hard EOL
	raw := &iterm2.LineContents{
		Text:         proto.String("x"),
		Continuation: iterm2.LineContents_CONTINUATION_HARD_EOL.Enum(),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)},
		},
	}
	lc := newLineContent(raw)
	assert.True(t, lc.HardEOL())

	// Soft EOL
	raw.Continuation = iterm2.LineContents_CONTINUATION_SOFT_EOL.Enum()
	lc = newLineContent(raw)
	assert.False(t, lc.HardEOL())
}

// TestLineContent_Nil tests newLineContent function with nil input.
func TestLineContent_Nil(t *testing.T) {
	lc := newLineContent(nil)
	assert.Nil(t, lc)
}

// TestLineContent_StyleAt tests LineContent.StyleAt method for retrieving style at position.
func TestLineContent_StyleAt(t *testing.T) {
	boldTrue := proto.Bool(true)
	raw := &iterm2.LineContents{
		Text: proto.String("abc"),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(3)},
		},
		Style: []*iterm2.CellStyle{
			{Bold: boldTrue, Repeats: proto.Uint32(1)},
			{Repeats: proto.Uint32(2)},
		},
	}
	lc := newLineContent(raw)

	// Cell 0 should be bold
	cs := lc.StyleAt(0)
	require.NotNil(t, cs)
	assert.True(t, cs.Bold())

	// Cell 1 should not be bold (repeated style with no bold)
	cs = lc.StyleAt(1)
	require.NotNil(t, cs)
	assert.False(t, cs.Bold())

	// Out of bounds
	assert.Nil(t, lc.StyleAt(99))
}

// TestLineContent_StyleAt_NoStyles tests LineContent.StyleAt method returns nil when no styles defined.
func TestLineContent_StyleAt_NoStyles(t *testing.T) {
	raw := &iterm2.LineContents{
		Text: proto.String("x"),
		CodePointsPerCell: []*iterm2.CodePointsPerCell{
			{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(1)},
		},
	}
	lc := newLineContent(raw)
	assert.Nil(t, lc.StyleAt(0))
}

// TestCellStyle_Bold tests CellStyle.Bold method.
func TestCellStyle_Bold(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{Bold: proto.Bool(true)}}
	assert.True(t, cs.Bold())
}

// TestCellStyle_TextAttributes tests CellStyle basic text attribute methods.
func TestCellStyle_TextAttributes(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		Bold:          proto.Bool(true),
		Italic:        proto.Bool(true),
		Underline:     proto.Bool(true),
		Strikethrough: proto.Bool(true),
	}}
	assert.True(t, cs.Bold())
	assert.True(t, cs.Italic())
	assert.True(t, cs.Underline())
	assert.True(t, cs.Strikethrough())
	assert.False(t, cs.HasFG())
	assert.False(t, cs.HasBG())
}

// TestCellStyle_FGStandard tests CellStyle.FGStandard method for retrieving foreground standard color.
func TestCellStyle_FGStandard(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		FgColor: &iterm2.CellStyle_FgStandard{FgStandard: 7},
	}}
	v, ok := cs.FGStandard()
	assert.True(t, ok)
	assert.Equal(t, uint32(7), v)
	_, ok = cs.FGRGB()
	assert.False(t, ok)
}

// TestCellStyle_FGRGB tests CellStyle.FGRGB method for retrieving foreground RGB color.
func TestCellStyle_FGRGB(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		FgColor: &iterm2.CellStyle_FgRgb{FgRgb: &iterm2.RGBColor{Red: proto.Uint32(255), Green: proto.Uint32(128), Blue: proto.Uint32(64)}},
	}}
	rgb, ok := cs.FGRGB()
	assert.True(t, ok)
	assert.Equal(t, uint32(255), rgb.GetRed())
	assert.Equal(t, uint32(128), rgb.GetGreen())
	assert.Equal(t, uint32(64), rgb.GetBlue())
	assert.True(t, cs.HasFG())
}

// TestCellStyle_BGStandard tests CellStyle.BGStandard method for retrieving background standard color.
func TestCellStyle_BGStandard(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		BgColor: &iterm2.CellStyle_BgStandard{BgStandard: 15},
	}}
	v, ok := cs.BGStandard()
	assert.True(t, ok)
	assert.Equal(t, uint32(15), v)
	assert.True(t, cs.HasBG())
}

// TestCellStyle_Image tests CellStyle.Image method for retrieving image placeholder type.
func TestCellStyle_Image(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		Image: iterm2.ImagePlaceholderType_ITERM2.Enum(),
	}}
	assert.Equal(t, iterm2.ImagePlaceholderType_ITERM2, cs.Image())
}

// TestCellStyle_URL tests CellStyle.URL method for retrieving URL and identifier.
func TestCellStyle_URL(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		Url: &iterm2.URL{Url: proto.String("http://example.com"), Identifier: proto.String("id1")},
	}}
	url, id, ok := cs.URL()
	assert.True(t, ok)
	assert.Equal(t, "http://example.com", url)
	assert.Equal(t, "id1", id)
}

// TestCellStyle_UnderlineRGB tests CellStyle.UnderlineRGB method for retrieving underline RGB color.
func TestCellStyle_UnderlineRGB(t *testing.T) {
	cs := &CellStyle{raw: &iterm2.CellStyle{
		UnderlineColor: &iterm2.RGBColor{Red: proto.Uint32(100), Green: proto.Uint32(200), Blue: proto.Uint32(50)},
	}}
	rgb, ok := cs.UnderlineRGB()
	assert.True(t, ok)
	assert.Equal(t, uint32(100), rgb.GetRed())
	assert.Equal(t, uint32(50), rgb.GetBlue())
}

// TestExpandStyles_Empty tests expandStyles handling empty style list.
func TestExpandStyles_Empty(t *testing.T) {
	result := expandStyles(nil)
	assert.Empty(t, result)
}

// TestExpandStyles_Simple tests expandStyles function with simple style list.
func TestExpandStyles_Simple(t *testing.T) {
	styles := []*iterm2.CellStyle{
		{Repeats: proto.Uint32(1), Bold: proto.Bool(true)},
		{Repeats: proto.Uint32(2), Italic: proto.Bool(true)},
	}
	result := expandStyles(styles)
	require.Len(t, result, 3)
	assert.True(t, result[0].Bold())
	assert.True(t, result[1].Italic())
	assert.True(t, result[2].Italic())
}

// TestExpandStyles_DefaultRepeats tests expandStyles function with default repeat value.
func TestExpandStyles_DefaultRepeats(t *testing.T) {
	// repeats=0 should default to 1
	styles := []*iterm2.CellStyle{
		{Bold: proto.Bool(true)},
	}
	result := expandStyles(styles)
	assert.Len(t, result, 1)
}

// TestScreenStreamer_NewAndClose tests ScreenStreamer creation and cleanup.
func TestScreenStreamer_NewAndClose(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	// Respond to subscribe RPC + unsubscribe RPC (2 messages).
	errCh := make(chan error, 1)
	go func() {
		for i := 0; i < 2; i++ {
			data, ok := <-mock.writeCh
			if !ok {
				errCh <- nil
				return
			}
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
		errCh <- nil
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)
	s.Close()

	// Second Close must be safe.
	s.Close()

	// Wait for the responder goroutine to finish.
	<-errCh
}

// TestScreenStreamer_NotifiesAndCloses tests ScreenStreamer receives notifications and closes properly.
func TestScreenStreamer_NotifiesAndCloses(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Test that Close properly unsubscribes without leaking.
	// Full integration (subscribe→notify→GetBuffer) is tested via the
	// manual Dispatch test below.
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	errCh := make(chan error, 1)
	go func() {
		for i := 0; i < 2; i++ {
			data, ok := <-mock.writeCh
			if !ok {
				errCh <- nil
				return
			}
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
		errCh <- nil
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)
	s.Close()
	s.Close() // idempotent
	<-errCh
}

// TestScreenStreamer_DispatchCallback tests ScreenStreamer callback fires when Dispatch receives a notification.
func TestScreenStreamer_DispatchCallback(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify the callback fires when Dispatch receives a matching notification.
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
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

		// GetBuffer RPC (triggered by run goroutine)
		data = <-mock.writeCh
		_ = proto.Unmarshal(data, &req)
		resp = &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_GetBufferResponse{
				GetBufferResponse: &iterm2.GetBufferResponse{
					Contents: []*iterm2.LineContents{
						{
							Text: proto.String("test"),
							CodePointsPerCell: []*iterm2.CodePointsPerCell{
								{NumCodePoints: proto.Int32(1), Repeats: proto.Int32(4)},
							},
						},
					},
				},
			},
		}
		b, _ = proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}

		// Unsubscribe RPC
		data = <-mock.writeCh
		_ = proto.Unmarshal(data, &req)
		resp = &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_NotificationResponse{
				NotificationResponse: &iterm2.NotificationResponse{
					Status: iterm2.NotificationResponse_OK.Enum(),
				},
			},
		}
		b, _ = proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)
	defer s.Close()

	// Manually trigger Dispatch with a screen-update notification.
	// This simulates what dispatchLoop does when a notification arrives.
	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				ScreenUpdateNotification: &iterm2.ScreenUpdateNotification{
					Session: proto.String("s1"),
				},
			},
		},
	}
	conn.Dispatch(notif)

	sc := <-s.Chan()
	lines := sc.Lines()
	require.Len(t, lines, 1)
	assert.Equal(t, "test", lines[0].Text())
}

// TestScreenStreamer_NotifyDropsOnFullBuffer tests ScreenStreamer drops notifications when buffer is full.
func TestScreenStreamer_NotifyDropsOnFullBuffer(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify signal-drop behavior: when the notify channel is full, the
	// callback drops the signal instead of blocking (select default path).
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 3; i++ { // subscribe + GetBuffer + unsubscribe
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)
			var resp iterm2.ServerOriginatedMessage
			resp.Id = req.Id
			if i == 1 {
				resp.Submessage = &iterm2.ServerOriginatedMessage_GetBufferResponse{
					GetBufferResponse: &iterm2.GetBufferResponse{},
				}
			} else {
				resp.Submessage = &iterm2.ServerOriginatedMessage_NotificationResponse{
					NotificationResponse: &iterm2.NotificationResponse{
						Status: iterm2.NotificationResponse_OK.Enum(),
					},
				}
			}
			b, _ := proto.Marshal(&resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)

	// Fill the notify channel (capacity 1). The run goroutine hasn't
	// drained it yet.
	s.notify <- struct{}{}

	// Second send should be dropped (no deadlock).
	select {
	case s.notify <- struct{}{}:
		t.Error("expected drop on full notify channel")
	default:
		// expected
	}

	// Drain the screen contents channel after GetBuffer completes.
	<-s.Chan()
	s.Close()
	<-done
}

// TestScreenStreamer_SubscribeError tests ScreenStreamer handles subscription error from server.
func TestScreenStreamer_SubscribeError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Use a mock that fails the subscribe RPC
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)

		// Return an error response
		resp := &iterm2.ServerOriginatedMessage{
			Id:         req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_Error{Error: "subscribe failed"},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	_, err := NewScreenStreamer(conn, "s1")
	require.Error(t, err)
}

// TestScreenStreamer_Disconnect tests ScreenStreamer behavior on disconnect.
func TestScreenStreamer_Disconnect(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify that when the WebSocket connection is lost, the ScreenStreamer
	// cleans up gracefully: Chan() is closed, IsConnected() returns false,
	// and Close() remains idempotent.
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	// Responder: only subscribe RPC (no unsubscribe — we're simulating disconnect)
	go func() {
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
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)

	// Simulate WebSocket disconnection by feeding a ReadError.
	mock.readCh <- mockReadResult{msgType: 0, data: nil, err: errors.New("connection closed")}

	// The disconnect callback should have fired -> s.Close() -> close(done) ->
	// run goroutine exits -> close(ch).  Chan() should be closed.
	_, ok := <-s.Chan()
	assert.False(t, ok)

	// IsConnected should be false.
	assert.False(t, conn.IsConnected())

	// Close() must be idempotent.
	s.Close()
	s.Close()
}

// TestScreenStreamer_DisconnectWithUserCallback tests ScreenStreamer fires OnDisconnect callbacks on disconnect.
func TestScreenStreamer_DisconnectWithUserCallback(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify that OnDisconnect callbacks fire on disconnect,
	// and the ScreenStreamer still cleans up properly.
	mock := newMockWS()

	userCalled := make(chan struct{})
	conn := NewConnection("c", "k", "test")
	conn.OnDisconnect(func() {
		close(userCalled)
	})
	conn.ConnectWithWS(ctx, mock)

	// Register a second callback via the OnDisconnect method.
	secondCalled := make(chan struct{})
	conn.OnDisconnect(func() {
		close(secondCalled)
	})

	// Responder: only subscribe RPC
	go func() {
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
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)

	// Disconnect
	mock.readCh <- mockReadResult{msgType: 0, data: nil, err: errors.New("closed")}

	// Wait for ScreenStreamer to fully clean up (Chan closes when done).
	_, ok := <-s.Chan()
	assert.False(t, ok)

	// Now all callbacks must have fired (they fire before Close() completes).
	select {
	case <-userCalled:
	default:
		t.Error("WithOnDisconnect callback should have fired")
	}

	select {
	case <-secondCalled:
	default:
		t.Error("OnDisconnect callback should have fired")
	}

	s.Close()
}

// TestScreenStreamer_Disconnect_NoLeak tests ScreenStreamer disconnects cleanly without goroutine leaks.
func TestScreenStreamer_Disconnect_NoLeak(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify that disconnecting a ScreenStreamer that is NOT in the middle
	// of a GetBuffer call cleans up properly and doesn't deadlock.
	// (Disconnecting while GetBuffer is in-progress is a known limitation:
	// the run goroutine blocks on WriteMessage with writeMu held, and no
	// other goroutine can Close until the write times out.)
	mock := newMockWS()
	conn := NewConnection("c", "k", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
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
	}()

	s, err := NewScreenStreamer(conn, "s1")
	require.NoError(t, err)

	// Disconnect cleanly (no in-flight GetBuffer).
	mock.readCh <- mockReadResult{msgType: 0, data: nil, err: errors.New("closed")}

	// Wait for cleanup: Chan() closes when the disconnect callback fires
	// and the run goroutine exits.
	_, ok := <-s.Chan()
	assert.False(t, ok)

	// Close must be idempotent.
	s.Close()
	s.Close()
}
