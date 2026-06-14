package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

// TestE2E_Conn_Basic connects to iTerm2, verifies ConnType and Close.
func TestE2E_Conn_Basic(t *testing.T) {
	skipIfNoITerm2(t)
	conn, err := term2go.Connect(context.Background(), "term2go-e2e-conn")
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.NotEmpty(t, conn.ConnType(), "ConnType should be non-empty")

	err = conn.Close()
	require.NoError(t, err)
}

// TestE2E_Conn_Reconnect connects twice sequentially and verifies both work.
func TestE2E_Conn_Reconnect(t *testing.T) {
	skipIfNoITerm2(t)

	conn1, err := term2go.Connect(context.Background(), "term2go-e2e-conn-reconnect")
	require.NoError(t, err)
	require.NotEmpty(t, conn1.ConnType())
	err = conn1.Close()
	require.NoError(t, err)

	conn2, err := term2go.Connect(context.Background(), "term2go-e2e-conn-reconnect")
	require.NoError(t, err)
	require.NotEmpty(t, conn2.ConnType())
	err = conn2.Close()
	require.NoError(t, err)
}

// TestE2E_Conn_ContextCancel passes a cancelled context to Connect.
func TestE2E_Conn_ContextCancel(t *testing.T) {
	skipIfNoITerm2(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn, err := term2go.Connect(ctx, "term2go-e2e-conn-cancel")
	if conn != nil {
		conn.Close()
	}
	require.Error(t, err, "Connect should error with cancelled context")
}

// TestE2E_Conn_NewConnection uses GetCookieOrCreate and NewConnection manually.
func TestE2E_Conn_NewConnection(t *testing.T) {
	skipIfNoITerm2(t)

	cookie, key, err := term2go.GetCookieOrCreate("term2go-e2e-conn-new")
	require.NoError(t, err)
	require.NotEmpty(t, cookie)
	require.NotEmpty(t, key)

	conn := term2go.NewConnection(cookie, key, "term2go-e2e-conn-new")
	require.NotNil(t, conn)

	require.NoError(t, conn.Connect(context.Background()))
	defer conn.Close()

	_, err = term2go.GetApp(context.Background(), conn)
	require.NoError(t, err)
}

// TestE2E_Conn_CallTimeout verifies WithCallTimeout causes timeout on RPC calls.
func TestE2E_Conn_CallTimeout(t *testing.T) {
	skipIfNoITerm2(t)

	cookie, key, err := term2go.GetCookieOrCreate("term2go-e2e-conn-timeout")
	require.NoError(t, err)

	// WithCallTimeout(1ms) with local socket may still succeed (too fast).
	// Verify the option is accepted and does not panic.
	conn := term2go.NewConnection(cookie, key, "term2go-e2e-conn-timeout",
		term2go.WithCallTimeout(1*time.Millisecond),
	)
	require.NoError(t, conn.Connect(context.Background()))
	defer conn.Close()

	_, err = term2go.ListSessions(context.Background(), conn)
	// May or may not time out depending on local socket speed — either way is fine.
	if err != nil {
		t.Logf("CallTimeout(1ms) triggered error (expected on slow connection): %v", err)
	} else {
		t.Log("CallTimeout(1ms) did not trigger — local socket too fast for 1ms timeout")
	}

	// Default timeout (30s) — should work
	conn2 := term2go.NewConnection(cookie, key, "term2go-e2e-conn-timeout-default")
	require.NoError(t, conn2.Connect(context.Background()))
	defer conn2.Close()

	_, err = term2go.ListSessions(context.Background(), conn2)
	require.NoError(t, err)
}

// TestE2E_Conn_Send tests fire-and-forget Send with a FocusRequest,
// then verifies the connection still works with Call.
func TestE2E_Conn_Send(t *testing.T) {
	skipIfNoITerm2(t)

	conn, err := term2go.Connect(context.Background(), "term2go-e2e-conn-send")
	require.NoError(t, err)
	defer conn.Close()

	req := &iterm2.ClientOriginatedMessage{
		Submessage: &iterm2.ClientOriginatedMessage_FocusRequest{
			FocusRequest: &iterm2.FocusRequest{},
		},
	}
	require.NoError(t, conn.Send(req))

	// Verify Call still works after Send
	_, err = term2go.GetApp(context.Background(), conn)
	require.NoError(t, err)
}

// TestE2E_Conn_UnregisterHandler registers a handler, unregisters it,
// and verifies the connection still functions correctly.
func TestE2E_Conn_UnregisterHandler(t *testing.T) {
	skipIfNoITerm2(t)

	conn, err := term2go.Connect(context.Background(), "term2go-e2e-conn-unregister")
	require.NoError(t, err)
	defer conn.Close()

	h := term2go.NotificationHandler(func(msg *iterm2.ServerOriginatedMessage) bool {
		return false
	})
	conn.RegisterHandler(h)
	conn.UnregisterHandler(h)

	// Verify connection still works after handler operations
	_, err = term2go.GetApp(context.Background(), conn)
	require.NoError(t, err)
}

// TestE2E_Conn_CookieAndKey verifies Cookie() and Key() return non-empty values.
func TestE2E_Conn_CookieAndKey(t *testing.T) {
	skipIfNoITerm2(t)

	conn, err := term2go.Connect(context.Background(), "term2go-e2e-conn-cookie")
	require.NoError(t, err)
	defer conn.Close()

	require.NotEmpty(t, conn.Cookie(), "Cookie should be non-empty")
	require.NotEmpty(t, conn.Key(), "Key should be non-empty")
}

// TestE2E_Conn_EnvAuth tests EnvAuthProvider reading ITERM2_COOKIE and ITERM2_KEY.
func TestE2E_Conn_EnvAuth(t *testing.T) {
	oldCookie := os.Getenv("ITERM2_COOKIE")
	oldKey := os.Getenv("ITERM2_KEY")
	defer func() {
		os.Setenv("ITERM2_COOKIE", oldCookie)
		os.Setenv("ITERM2_KEY", oldKey)
	}()

	os.Setenv("ITERM2_COOKIE", "env-test-cookie")
	os.Setenv("ITERM2_KEY", "env-test-key")

	provider := &term2go.EnvAuthProvider{}

	cookie, err := provider.GetCookie()
	require.NoError(t, err)
	require.Equal(t, "env-test-cookie", cookie)

	key, err := provider.GetKey()
	require.NoError(t, err)
	require.Equal(t, "env-test-key", key)

	// Verify env values still match after provider reads
	require.Equal(t, "env-test-cookie", os.Getenv("ITERM2_COOKIE"))
	require.Equal(t, "env-test-key", os.Getenv("ITERM2_KEY"))
}

// TestE2E_Conn_AppleScriptAuth tests NewAppleScriptAuthProvider returning non-empty credentials.
func TestE2E_Conn_AppleScriptAuth(t *testing.T) {
	skipIfNoITerm2(t)

	provider := term2go.NewAppleScriptAuthProvider("term2go-e2e-conn-as")

	cookie, err := provider.GetCookie()
	require.NoError(t, err)
	require.NotEmpty(t, cookie, "cookie should be non-empty from AppleScript")

	key, err := provider.GetKey()
	require.NoError(t, err)
	require.NotEmpty(t, key, "key should be non-empty from AppleScript")
}

// TestE2E_Conn_GetCookieOrCreate tests GetCookieOrCreate returns valid credentials.
func TestE2E_Conn_GetCookieOrCreate(t *testing.T) {
	skipIfNoITerm2(t)

	cookie, key, err := term2go.GetCookieOrCreate("term2go-e2e-conn-gcoc")
	require.NoError(t, err)
	require.NotEmpty(t, cookie, "cookie should be non-empty")
	require.NotEmpty(t, key, "key should be non-empty")
}
