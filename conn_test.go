package term2go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// mockWS implements wsConn for testing.
type mockWS struct {
	readCh  chan mockReadResult
	writeCh chan []byte
	closed  chan struct{}
}

type mockReadResult struct {
	msgType int
	data    []byte
	err     error
}

func newMockWS() *mockWS {
	return &mockWS{
		readCh:  make(chan mockReadResult),
		writeCh: make(chan []byte),
		closed:  make(chan struct{}),
	}
}

func (m *mockWS) ReadMessage() (int, []byte, error) {
	r := <-m.readCh
	return r.msgType, r.data, r.err
}

func (m *mockWS) WriteMessage(messageType int, data []byte) error {
	m.writeCh <- data
	return nil
}

func (m *mockWS) Close() error {
	select {
	case <-m.closed:
	default:
		close(m.closed)
	}
	return nil
}

func (m *mockWS) SetReadDeadline(t time.Time) error { return nil }
func (m *mockWS) SetReadLimit(limit int64)          {}

// Interface compliance
func TestCallerInterface(t *testing.T) {
	var _ Caller = (*Connection)(nil)
	var _ Notifier = (*Connection)(nil)
}

// TestConnection_Call tests Connection.Call method for sending a request and receiving a response.
func TestConnection_Call(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)
		id := req.GetId()

		resp := &iterm2.ServerOriginatedMessage{
			Id: proto.Int64(id),
			Submessage: &iterm2.ServerOriginatedMessage_ListSessionsResponse{
				ListSessionsResponse: &iterm2.ListSessionsResponse{},
			},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	req := &iterm2.ClientOriginatedMessage{
		Submessage: &iterm2.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &iterm2.ListSessionsRequest{},
		},
	}
	resp, err := conn.Call(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp.GetListSessionsResponse())
}

// TestConnection_Call_Error tests Connection.Call method when the server returns an error message.
func TestConnection_Call_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)

		resp := &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_Error{
				Error: "test error",
			},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	req := &iterm2.ClientOriginatedMessage{}
	_, err := conn.Call(ctx, req)
	require.Error(t, err)
}

// TestConnection_Send tests Connection.Send method for sending a fire-and-forget request.
func TestConnection_Send(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		<-mock.writeCh
	}()

	req := &iterm2.ClientOriginatedMessage{}
	require.NoError(t, conn.Send(req))
}

// TestConnection_Notify tests Connection as a Notifier for receiving notifications.
func TestConnection_Notify(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	handlerCalled := make(chan struct{})
	conn.RegisterHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
		close(handlerCalled)
		return true
	})

	notif := &iterm2.ServerOriginatedMessage{
		Submessage: &iterm2.ServerOriginatedMessage_Notification{
			Notification: &iterm2.Notification{
				NewSessionNotification: &iterm2.NewSessionNotification{},
			},
		},
	}
	b, _ := proto.Marshal(notif)
	mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}

	<-handlerCalled
}

// TestConnection_Close_Nil tests Connection.Close method when the connection is nil.
func TestConnection_Close_Nil(t *testing.T) {
	conn := NewConnection("cookie", "key", "test")
	require.NoError(t, conn.Close())
}

// TestConnection_Close tests Connection.Close method for closing an active connection.
func TestConnection_Close(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	require.NoError(t, conn.Close())

	select {
	case <-mock.closed:
	default:
		t.Error("expected mock to be closed")
	}
}

// TestConnection_NewConnection tests NewConnection function for creating a new connection.
func TestConnection_NewConnection(t *testing.T) {
	conn := NewConnection("my-cookie", "my-key", "my-script")
	assert.Equal(t, "my-cookie", conn.Cookie())
	assert.Equal(t, "my-key", conn.Key())
	assert.NotNil(t, conn.pending)
}

// TestConnection_ConnType tests Connection.ConnType method for returning the connection type.
func TestConnection_ConnType(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	assert.Equal(t, "mock", conn.ConnType())
}

// TestGetCookieOrCreate tests GetCookieOrCreate function for obtaining credentials from environment variables.
func TestGetCookieOrCreate(t *testing.T) {
	t.Setenv("ITERM2_COOKIE", "test-cookie")
	t.Setenv("ITERM2_KEY", "test-key")

	cookie, key, err := GetCookieOrCreate("")
	require.NoError(t, err)
	assert.Equal(t, "test-cookie", cookie)
	assert.Equal(t, "test-key", key)
}

// TestEnvAuthProvider_GetCookie tests EnvAuthProvider.GetCookie method for getting cookie from environment.
func TestEnvAuthProvider_GetCookie(t *testing.T) {
	t.Setenv("ITERM2_COOKIE", "env-cookie")
	p := &EnvAuthProvider{}
	cookie, err := p.GetCookie()
	require.NoError(t, err)
	assert.Equal(t, "env-cookie", cookie)
}

// TestEnvAuthProvider_GetCookie_Missing tests EnvAuthProvider.GetCookie method when ITERM2_COOKIE is not set.
func TestEnvAuthProvider_GetCookie_Missing(t *testing.T) {
	t.Setenv("ITERM2_COOKIE", "")
	p := &EnvAuthProvider{}
	_, err := p.GetCookie()
	require.Error(t, err)
}

// TestEnvAuthProvider_GetKey tests EnvAuthProvider.GetKey method for getting key from environment.
func TestEnvAuthProvider_GetKey(t *testing.T) {
	t.Setenv("ITERM2_KEY", "env-key")
	p := &EnvAuthProvider{}
	key, err := p.GetKey()
	require.NoError(t, err)
	assert.Equal(t, "env-key", key)
}

// TestUnixSocketPath tests unixSocketPath function for returning the iTerm2 Unix socket path.
func TestUnixSocketPath(t *testing.T) {
	path := unixSocketPath()
	if path == "" {
		t.Skip("cannot determine home dir")
	}
}

// TestConnection_MultipleCalls tests Connection.Call method for handling multiple sequential calls.
func TestConnection_MultipleCalls(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		for i := 0; i < 3; i++ {
			data := <-mock.writeCh
			var req iterm2.ClientOriginatedMessage
			_ = proto.Unmarshal(data, &req)

			resp := &iterm2.ServerOriginatedMessage{
				Id: req.Id,
				Submessage: &iterm2.ServerOriginatedMessage_ListSessionsResponse{
					ListSessionsResponse: &iterm2.ListSessionsResponse{},
				},
			}
			b, _ := proto.Marshal(resp)
			mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
		}
	}()

	for i := 0; i < 3; i++ {
		req := &iterm2.ClientOriginatedMessage{}
		_, err := conn.Call(ctx, req)
		require.NoErrorf(t, err, "call %d: unexpected error", i)
	}
}

// TestConnection_UnregisterHandler tests Connection.UnregisterHandler method for removing a notification handler.
func TestConnection_UnregisterHandler(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	unregisteredCalled := false
	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		unregisteredCalled = true
		return false
	}
	conn.RegisterHandler(h)
	conn.UnregisterHandler(h)

	// Use a second handler as a signal that dispatch happened.
	dispatched := make(chan struct{})
	conn.RegisterHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
		close(dispatched)
		return false
	})

	b, _ := proto.Marshal(&iterm2.ServerOriginatedMessage{})
	mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	<-dispatched

	assert.False(t, unregisteredCalled)
}

// TestConnection_Dispatcher_StopsOnTrue tests that notification dispatch stops when a handler returns true.
func TestConnection_Dispatcher_StopsOnTrue(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	firstCalled := make(chan struct{})
	secondCalled := false

	conn.RegisterHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
		close(firstCalled)
		return true // stop processing
	})
	conn.RegisterHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
		secondCalled = true
		return false
	})

	b, _ := proto.Marshal(&iterm2.ServerOriginatedMessage{})
	mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}

	<-firstCalled
	assert.False(t, secondCalled)
}

// writeErrorMock returns error on WriteMessage; used for Send/Call error paths.
type writeErrorMock struct {
	*mockWS
	writeErr error
}

func (m *writeErrorMock) WriteMessage(messageType int, data []byte) error {
	return m.writeErr
}

// TestConnection_Call_WriteError tests Connection.Call method when WriteMessage fails.
func TestConnection_Call_WriteError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	m := &writeErrorMock{
		mockWS:   newMockWS(),
		writeErr: fmt.Errorf("simulated write failure"),
	}
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, m)

	req := &iterm2.ClientOriginatedMessage{}
	_, err := conn.Call(ctx, req)
	require.Error(t, err)
}

// TestConnection_Send_WriteError tests Connection.Send method when WriteMessage fails.
func TestConnection_Send_WriteError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	m := &writeErrorMock{
		mockWS:   newMockWS(),
		writeErr: fmt.Errorf("simulated write failure"),
	}
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, m)

	req := &iterm2.ClientOriginatedMessage{}
	require.Error(t, conn.Send(req))
}

// connDialRecorder is used for testing Connect functionality, records dial calls and delegates to user-provided function.
type connDialRecorder struct {
	fn    func(context.Context, string, http.Header) (wsConn, *http.Response, error)
	urls  []string
	calls int
}

func (r *connDialRecorder) dial(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
	r.urls = append(r.urls, url)
	r.calls++
	if r.fn != nil {
		return r.fn(ctx, url, hdr)
	}
	return nil, nil, errors.New("no fn set")
}

// simpleDialer returns a connDialFunc that always returns the given wsConn.
func simpleDialer(ws wsConn) connDialFunc {
	return func(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
		return ws, nil, nil
	}
}

// mockSocketDir creates temporary directory structure simulating iTerm2 Unix socket path.
func mockSocketDir(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	socketDir := filepath.Join(home, "Library", "Application Support", "iTerm2", "private")
	if err := os.MkdirAll(socketDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	p := filepath.Join(socketDir, "socket")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
	t.Setenv("HOME", home)
}

// TestConnection_Connect_UnixSuccess tests Connection.Connect method when Unix socket connection succeeds.
func TestConnection_Connect_UnixSuccess(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mockSocketDir(t)

	ws := newMockWS()
	r := &connDialRecorder{fn: simpleDialer(ws)}
	conn := NewConnection("cookie", "key", "test")
	conn.dialFunc = r.dial

	require.NoError(t, conn.Connect(ctx))
	assert.Equal(t, "unix", conn.ConnType())
	assert.Equal(t, 1, r.calls)
}

// TestConnection_Connect_UnixFail_TCPFallback tests Connection.Connect method falls back to TCP when Unix socket fails.
func TestConnection_Connect_UnixFail_TCPFallback(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mockSocketDir(t)

	ws := newMockWS()
	first := true
	r := &connDialRecorder{
		fn: func(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
			if first {
				first = false
				return nil, nil, errors.New("unix fail")
			}
			return ws, nil, nil
		},
	}
	conn := NewConnection("cookie", "key", "test")
	conn.dialFunc = r.dial

	require.NoError(t, conn.Connect(ctx))
	assert.Equal(t, "tcp", conn.ConnType())
	assert.Equal(t, 2, r.calls)
}

// TestConnection_Connect_BothFail tests Connection.Connect method returns error when both Unix socket and TCP fail.
func TestConnection_Connect_BothFail(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mockSocketDir(t)

	first := true
	r := &connDialRecorder{
		fn: func(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
			if first {
				first = false
				return nil, nil, errors.New("unix fail")
			}
			return nil, nil, errors.New("tcp fail")
		},
	}
	conn := NewConnection("cookie", "key", "test")
	conn.dialFunc = r.dial

	require.Error(t, conn.Connect(ctx))
	assert.Equal(t, 2, r.calls)
}

// TestConnection_Connect_NoUnixSocket_TCPOnly tests Connection.Connect method uses TCP only when HOME is not set.
func TestConnection_Connect_NoUnixSocket_TCPOnly(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	t.Setenv("HOME", "")
	_ = os.Unsetenv("HOME")

	ws := newMockWS()
	r := &connDialRecorder{fn: simpleDialer(ws)}
	conn := NewConnection("cookie", "key", "test")
	conn.dialFunc = r.dial

	require.NoError(t, conn.Connect(ctx))
	assert.Equal(t, "tcp", conn.ConnType())
	assert.Equal(t, 1, r.calls)
}

// TestAppleScriptAuthProvider_Success tests AppleScriptAuthProvider success return.
func TestAppleScriptAuthProvider_Success(t *testing.T) {
	p := NewAppleScriptAuthProvider("my-app")
	p.runner = func(scriptName string) (cookie, key string, err error) {
		assert.Equal(t, "my-app", scriptName)
		return "as-cookie", "as-key", nil
	}

	cookie, err := p.GetCookie()
	require.NoError(t, err)
	assert.Equal(t, "as-cookie", cookie)

	key, err := p.GetKey()
	require.NoError(t, err)
	assert.Equal(t, "as-key", key)
}

// TestAppleScriptAuthProvider_Error tests AppleScriptAuthProvider when osascript execution fails.
func TestAppleScriptAuthProvider_Error(t *testing.T) {
	p := NewAppleScriptAuthProvider("my-app")
	p.runner = func(_ string) (string, string, error) {
		return "", "", errors.New("osascript error")
	}

	_, err := p.GetCookie()
	require.Error(t, err)

	_, err = p.GetKey()
	require.Error(t, err)
}

// TestAppleScriptAuthProvider_New tests NewAppleScriptAuthProvider function for creating a new auth provider.
func TestAppleScriptAuthProvider_New(t *testing.T) {
	p := NewAppleScriptAuthProvider("my-app")
	assert.Equal(t, "my-app", p.scriptName)
}

// TestGetCookieOrCreate_EnvSuccess tests GetCookieOrCreate successfully getting credentials from environment variables.
func TestGetCookieOrCreate_EnvSuccess(t *testing.T) {
	t.Setenv("ITERM2_COOKIE", "env-c")
	t.Setenv("ITERM2_KEY", "env-k")

	cookie, key, err := GetCookieOrCreate("")
	require.NoError(t, err)
	assert.Equal(t, "env-c", cookie)
	assert.Equal(t, "env-k", key)
}

// TestConnection_DispatchLoop_ReadError tests dispatchLoop behavior when encountering error while reading messages.
func TestConnection_DispatchLoop_ReadError(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	// Feed a read error — dispatchLoop should exit (no panic).
	mock.readCh <- mockReadResult{msgType: 0, data: nil, err: errors.New("closed")}
}

// TestCall_Success tests Call function success return.
func TestCall_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	go func() {
		data := <-mock.writeCh
		var req iterm2.ClientOriginatedMessage
		_ = proto.Unmarshal(data, &req)

		resp := &iterm2.ServerOriginatedMessage{
			Id: req.Id,
			Submessage: &iterm2.ServerOriginatedMessage_ListSessionsResponse{
				ListSessionsResponse: &iterm2.ListSessionsResponse{},
			},
		}
		b, _ := proto.Marshal(resp)
		mock.readCh <- mockReadResult{msgType: websocket.BinaryMessage, data: b}
	}()

	req := &iterm2.ClientOriginatedMessage{
		Submessage: &iterm2.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &iterm2.ListSessionsRequest{},
		},
	}
	resp, err := conn.Call(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp.GetListSessionsResponse())
}

// TestCall_PreCanceled tests Call method returns context.Canceled when context is already canceled.
func TestCall_PreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(context.Background(), mock)

	req := &iterm2.ClientOriginatedMessage{}
	_, err := conn.Call(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCall_MidCallCancel tests Call method handles cancellation while waiting for response.
func TestCall_MidCallCancel(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	// Don't provide any response — Call will block in the select phase.
	// We cancel the context after starting the call.

	callCtx, midCancel := context.WithCancel(ctx)

	// Responder: consume the write but don't send response
	go func() {
		<-mock.writeCh // consume the request
		// Don't send response — call will block
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		req := &iterm2.ClientOriginatedMessage{}
		_, err := conn.Call(callCtx, req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	}()

	// Cancel immediately to demonstrate non-blocking cancellation.
	midCancel()
	<-done
}

// TestCall_WriteBlockedCancel tests Call method handles cancellation when WriteMessage is blocked.
func TestCall_WriteBlockedCancel(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	// Verify that cancelling while WriteMessage blocks does NOT deadlock.
	// The caller returns immediately, and the background goroutine eventually
	// releases writeMu when WriteMessage completes.

	mock := newMockWS()
	conn := NewConnection("cookie", "key", "test")
	conn.ConnectWithWS(ctx, mock)

	// First call acquires writeMu and blocks WriteMessage (nobody reads writeCh).
	// We do this from a goroutine because it will block forever.
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		req := &iterm2.ClientOriginatedMessage{
			Submessage: &iterm2.ClientOriginatedMessage_ListSessionsRequest{
				ListSessionsRequest: &iterm2.ListSessionsRequest{},
			},
		}
		// This will block on WriteMessage because nobody reads writeCh.
		conn.Call(ctx, req) //nolint:errcheck
	}()

	// Give the first call time to acquire writeMu.
	time.Sleep(10 * time.Millisecond)

	// Second call with a pre-cancelled context should return immediately
	// instead of blocking on writeMu.
	callCtx, blockCancel := context.WithCancel(ctx)
	blockCancel()

	req := &iterm2.ClientOriginatedMessage{}
	_, err := conn.Call(callCtx, req)
	require.Error(t, err)
	// The key is: we returned without deadlocking, even though writeMu is held.
}
