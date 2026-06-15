package term2go

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Caller abstracts the ability to make RPC calls to iTerm2.
type Caller interface {
	Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error)
	Send(req *iterm2.ClientOriginatedMessage) error
}

// NotificationHandler is a callback for incoming server notifications.
type NotificationHandler func(msg *iterm2.ServerOriginatedMessage) bool

// Notifier abstracts notification subscription management.
type Notifier interface {
	RegisterHandler(h NotificationHandler)
	UnregisterHandler(h NotificationHandler)
}

// wsConn abstracts WebSocket read/write operations for testability.
type wsConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetReadLimit(limit int64)
}

// connDialFunc is a factory for WebSocket connections.
// It accepts a context for cancellation during the dial handshake.
type connDialFunc func(ctx context.Context, urlStr string, requestHeader http.Header) (wsConn, *http.Response, error)

const (
	maxMessageSize          = 10 << 20 // 10 MiB
	defaultReadTimeout      = 60 * time.Second
	defaultCallTimeout      = 30 * time.Second
	defaultHandshakeTimeout = 45 * time.Second
)

// Connection manages the WebSocket connection to iTerm2.
type Connection struct {
	ws               wsConn
	mu               sync.Mutex // protects pending
	writeMu          sync.Mutex // serializes WebSocket writes (gorilla/websocket requirement)
	closeOnce        sync.Once  // makes Close idempotent
	reqID            int64
	pending          map[int64]chan *iterm2.ServerOriginatedMessage
	handlers         []NotificationHandler
	handlersMu       sync.RWMutex
	cookie           string
	key              string
	scriptName       string
	connType         string
	callTimeout      time.Duration
	readTimeout      time.Duration
	handshakeTimeout time.Duration
	connected        atomic.Bool
	dialFunc         connDialFunc // nil → use gorilla dialer; set in tests
	disconnectCbs    []func()
	disconnectMu     sync.Mutex
	// per-connection notification registry
	notifyMu  sync.RWMutex
	notifyMap map[string][]notifyEntry
	notifySeq int64
	// notifyCaller overrides the Caller used for notification RPCs (subscribe/unsubscribe).
	// When nil, c itself is used. This allows tests to inject a mock Caller without
	// needing a real WebSocket connection.
	notifyCaller Caller
	// protoVer stores the iTerm2 protocol version from the WebSocket
	// handshake response header (X-iTerm2-Protocol-Version).
	// Defaults to (0,0) which means no features are gated behind version checks.
	protoVer ProtocolVersion
}

// NewConnection creates a new Connection with optional configuration.
func NewConnection(cookie, key, scriptName string, opts ...Option) *Connection {
	c := &Connection{
		cookie:           cookie,
		key:              key,
		scriptName:       scriptName,
		pending:          make(map[int64]chan *iterm2.ServerOriginatedMessage),
		notifyMap:        make(map[string][]notifyEntry),
		callTimeout:      defaultCallTimeout,
		readTimeout:      defaultReadTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// IsConnected reports whether the WebSocket connection is still active.
func (c *Connection) IsConnected() bool {
	return c.connected.Load()
}

// OnDisconnect registers a callback that fires when the WebSocket connection
// is lost. Multiple callbacks can be registered; they are invoked in order.
func (c *Connection) OnDisconnect(fn func()) {
	c.disconnectMu.Lock()
	defer c.disconnectMu.Unlock()
	c.disconnectCbs = append(c.disconnectCbs, fn)
}

// caller returns the Caller to use for notification RPCs.
// If notifyCaller is set (e.g. in tests), it is used; otherwise c itself.
func (c *Connection) caller() Caller {
	if c.notifyCaller != nil {
		return c.notifyCaller
	}
	return c
}

// Connect establishes the WebSocket connection and starts the dispatch loop.
func (c *Connection) Connect(ctx context.Context) error {
	// Try Unix socket
	socketPath := unixSocketPath()
	if socketPath != "" {
		if _, statErr := os.Stat(socketPath); statErr == nil {
			ws, err := c.dialUnix(ctx, socketPath)
			if err == nil {
				c.ws = ws
				c.ws.SetReadLimit(maxMessageSize)
				c.connType = "unix"
				c.connected.Store(true)
				go c.dispatchLoop(ctx)
				return nil
			}
		}
	}

	// TCP fallback
	ws, err := c.dialTCP(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect to iTerm2 (is it running?): %w", err)
	}
	c.ws = ws
	c.ws.SetReadLimit(maxMessageSize)
	c.connType = "tcp"
	c.connected.Store(true)
	go c.dispatchLoop(ctx)
	return nil
}

// getDialer returns the injected dial function, or builds a default
// gorilla-based dialer when none is set.
func (c *Connection) getDialer(netDial func(string, string) (net.Conn, error)) connDialFunc {
	if c.dialFunc != nil {
		return c.dialFunc
	}
	d := &websocket.Dialer{
		NetDial:          netDial,
		HandshakeTimeout: c.handshakeTimeout,
		Subprotocols:     []string{"api.iterm2.com"},
	}
	return func(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
		return d.DialContext(ctx, url, hdr)
	}
}

func (c *Connection) getDialerWithDialContext(netDialContext func(ctx context.Context, network, addr string) (net.Conn, error)) connDialFunc {
	if c.dialFunc != nil {
		return c.dialFunc
	}
	d := &websocket.Dialer{
		NetDialContext:   netDialContext,
		HandshakeTimeout: c.handshakeTimeout,
		Subprotocols:     []string{"api.iterm2.com"},
	}
	return func(ctx context.Context, url string, hdr http.Header) (wsConn, *http.Response, error) {
		return d.DialContext(ctx, url, hdr)
	}
}

// buildHeaders constructs the shared WebSocket upgrade headers.
func (c *Connection) buildHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Origin", "ws://localhost/")
	headers.Set("x-iterm2-library-version", "go 0.0")
	headers.Set("x-iterm2-disable-auth-ui", "true")
	headers.Set("x-iterm2-cookie", c.cookie)
	headers.Set("x-iterm2-key", c.key)
	headers.Set("x-iterm2-advisory-name", c.scriptName)
	return headers
}

// dialUnix connects via Unix domain socket.
func (c *Connection) dialUnix(ctx context.Context, socketPath string) (wsConn, error) {
	dial := c.getDialerWithDialContext(func(ctx context.Context, _, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", socketPath)
	})
	ws, resp, err := dial(ctx, "ws://localhost", c.buildHeaders())
	if err != nil {
		return nil, fmt.Errorf("unix ws dial: %w", err)
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if resp != nil {
		c.parseProtocolVersion(resp.Header)
	}
	return ws, nil
}

// dialTCP connects via TCP.
func (c *Connection) dialTCP(ctx context.Context) (wsConn, error) {
	dial := c.getDialer(nil)
	ws, resp, err := dial(ctx, "ws://localhost:1912", c.buildHeaders())
	if err != nil {
		return nil, fmt.Errorf("tcp ws dial: %w", err)
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if resp != nil {
		c.parseProtocolVersion(resp.Header)
	}
	return ws, nil
}

// ConnectWithWS sets a pre-established WebSocket for testing.
func (c *Connection) ConnectWithWS(ctx context.Context, conn wsConn) {
	c.ws = conn
	c.connType = "mock"
	c.connected.Store(true)
	go c.dispatchLoop(ctx)
}

// ConnType returns the connection type.
func (c *Connection) ConnType() string { return c.connType }

// ProtocolVersion returns the iTerm2 protocol version from the handshake.
// Defaults to (0,0) which means no features are gated behind version checks.
func (c *Connection) ProtocolVersion() ProtocolVersion { return c.protoVer }

// SetProtocolVersion sets the protocol version (for testing or manual override).
func (c *Connection) SetProtocolVersion(v ProtocolVersion) { c.protoVer = v }

// parseProtocolVersion reads the iTerm2 protocol version from the WebSocket
// handshake response header (X-iTerm2-Protocol-Version: major.minor).
// If the header is missing or malformed, protoVer remains {0,0}.
func (c *Connection) parseProtocolVersion(hdr http.Header) {
	v := hdr.Get("X-iTerm2-Protocol-Version")
	if v == "" {
		return
	}
	parts := strings.SplitN(v, ".", 2)
	if len(parts) != 2 {
		return
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return
	}
	c.protoVer = ProtocolVersion{major, minor}
}

// Call implements Caller.
func (c *Connection) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
	// Fast path: if ctx is already done, don't even write to the socket.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("call aborted: %w", err)
	}
	id := atomic.AddInt64(&c.reqID, 1)
	req.Id = proto.Int64(id)

	ch := make(chan *iterm2.ServerOriginatedMessage, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	data, err := proto.Marshal(req)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("marshal: %w", err)
	}
	c.writeMu.Lock()
	err = c.ws.WriteMessage(websocket.BinaryMessage, data)
	c.writeMu.Unlock()
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("send: %w", err)
	}

	// 3-way select: response, callTimeout, or ctx cancellation.
	var timer <-chan time.Time
	if c.callTimeout > 0 {
		timer = time.After(c.callTimeout)
	}
	// context.Background().Done() returns nil, which blocks forever in select.
	// So when ctx is Background(), this case never fires — zero overhead.

	select {
	case resp := <-ch:
		if resp.GetError() != "" {
			return nil, fmt.Errorf("rpc error: %s", resp.GetError())
		}
		return resp, nil
	case <-timer:
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		// Drain channel to prevent goroutine leak in dispatchLoop
		// Use a goroutine to avoid blocking if response arrives late
		go func() {
			select {
			case <-ch:
				// response arrived, channel is now empty
			case <-time.After(c.callTimeout):
				// give up waiting to avoid indefinite blocking
			}
		}()
		return nil, fmt.Errorf("call timeout after %v", c.callTimeout)
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		// Drain channel to prevent goroutine leak in dispatchLoop
		go func() {
			select {
			case <-ch:
				// response arrived, channel is now empty
			case <-time.After(c.callTimeout):
				// give up waiting to avoid indefinite blocking
			}
		}()
		return nil, fmt.Errorf("call aborted: %w", ctx.Err())
	}
}

// Send implements Caller.
func (c *Connection) Send(req *iterm2.ClientOriginatedMessage) error {
	id := atomic.AddInt64(&c.reqID, 1)
	req.Id = proto.Int64(id)
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	c.writeMu.Lock()
	err = c.ws.WriteMessage(websocket.BinaryMessage, data)
	c.writeMu.Unlock()
	return err
}

// RegisterHandler implements Notifier.
func (c *Connection) RegisterHandler(h NotificationHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers = append(c.handlers, h)
}

// UnregisterHandler implements Notifier.
func (c *Connection) UnregisterHandler(h NotificationHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	target := reflect.ValueOf(h).Pointer()
	for i, handler := range c.handlers {
		if reflect.ValueOf(handler).Pointer() == target {
			c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)
			return
		}
	}
}

// Cookie returns the cookie.
func (c *Connection) Cookie() string { return c.cookie }

// Key returns the key.
func (c *Connection) Key() string { return c.key }

// Close closes the WebSocket connection.  Safe to call multiple times.
func (c *Connection) Close() error {
	var err error
	c.closeOnce.Do(func() {
		if c.ws != nil {
			err = c.ws.Close()
		}
	})
	return err
}

func (c *Connection) dispatchLoop(ctx context.Context) {
	// Close the connection when the context is cancelled so ReadMessage unblocks.
	go func() {
		<-ctx.Done()
		_ = c.Close()
	}()

	defer func() {
		c.connected.Store(false)
		c.disconnectMu.Lock()
		for _, fn := range c.disconnectCbs {
			fn()
		}
		c.disconnectMu.Unlock()
	}()

	for {
		_ = c.ws.SetReadDeadline(time.Now().Add(c.readTimeout))
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		var msg iterm2.ServerOriginatedMessage
		if err = proto.Unmarshal(data, &msg); err != nil {
			// Log raw data for debugging (truncate to avoid flooding logs with large messages)
			const maxLogLen = 200
			if len(data) > maxLogLen {
				log.Printf("term2go: unmarshal error: %v, data=%s...", err, string(data[:maxLogLen]))
			} else {
				log.Printf("term2go: unmarshal error: %v, data=%s", err, string(data))
			}
			continue
		}

		if msg.Id != nil && *msg.Id > 0 {
			c.mu.Lock()
			ch, ok := c.pending[*msg.Id]
			if ok {
				delete(c.pending, *msg.Id)
			}
			c.mu.Unlock()
			if ok {
				// Non-blocking send: if channel is full, log and discard to avoid blocking dispatchLoop
				select {
				case ch <- &msg:
				default:
					log.Printf("term2go: dropping RPC response for id=%d (channel full)", *msg.Id)
					// Clean up the pending entry to prevent resource leak
					c.mu.Lock()
					delete(c.pending, *msg.Id)
					c.mu.Unlock()
				}
			}
		} else {
			c.handlersMu.RLock()
			handlers := make([]NotificationHandler, len(c.handlers))
			copy(handlers, c.handlers)
			c.handlersMu.RUnlock()
			for _, h := range handlers {
				if h(&msg) {
					break
				}
			}
		}
	}
}

// Connect is a convenience function.
func Connect(ctx context.Context, scriptName string) (*Connection, error) {
	cookie, key, err := GetCookieOrCreate(scriptName)
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	conn := NewConnection(cookie, key, scriptName)
	if err = conn.Connect(ctx); err != nil {
		_ = conn.Close() // safe: no ws opened if Connect failed
		return nil, err
	}
	return conn, nil
}

// AuthProvider provides authentication credentials.
type AuthProvider interface {
	GetCookie() (string, error)
	GetKey() (string, error)
}

// EnvAuthProvider reads ITERM2_COOKIE / ITERM2_KEY.
type EnvAuthProvider struct{}

func (p *EnvAuthProvider) GetCookie() (string, error) {
	cookie := os.Getenv("ITERM2_COOKIE")
	if cookie == "" {
		return "", fmt.Errorf("ITERM2_COOKIE not set")
	}
	return cookie, nil
}

func (p *EnvAuthProvider) GetKey() (string, error) {
	return os.Getenv("ITERM2_KEY"), nil
}

// AppleScriptAuthProvider obtains credentials via osascript.
type AppleScriptAuthProvider struct {
	scriptName string
	runner     appleScriptRunner // nil → use osascript
}

// appleScriptRunner can be injected by tests to avoid calling osascript.
type appleScriptRunner func(scriptName string) (cookie, key string, err error)

func NewAppleScriptAuthProvider(scriptName string) *AppleScriptAuthProvider {
	return &AppleScriptAuthProvider{scriptName: scriptName}
}

func (p *AppleScriptAuthProvider) GetCookie() (string, error) {
	cookie, _, err := p.runAppleScript()
	if err != nil {
		return "", err
	}
	return cookie, nil
}

func (p *AppleScriptAuthProvider) GetKey() (string, error) {
	_, key, err := p.runAppleScript()
	return key, err
}

func (p *AppleScriptAuthProvider) runAppleScript() (cookie, key string, err error) {
	if p.runner != nil {
		return p.runner(p.scriptName)
	}
	script := fmt.Sprintf(
		`tell application "iTerm2" to request cookie and key for app named "%s"`,
		escapeAppleScript(p.scriptName),
	)
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("osascript: %w (is iTerm2 running?)", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 2 {
		return "", "", fmt.Errorf("unexpected osascript output: %q", string(out))
	}
	return parts[0], parts[1], nil
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// GetCookieOrCreate returns credentials, trying env var first then AppleScript.
func GetCookieOrCreate(scriptName string) (cookie, key string, err error) {
	env := &EnvAuthProvider{}
	cookie, err = env.GetCookie()
	if err == nil && cookie != "" {
		k, _ := env.GetKey()
		return cookie, k, nil
	}
	as := NewAppleScriptAuthProvider(scriptName)
	// Call runAppleScript once — GetCookie/GetKey each call it, wasting an osascript invocation.
	return as.runAppleScript()
}

func unixSocketPath() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "iTerm2", "private", "socket")
}
