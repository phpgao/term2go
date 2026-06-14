package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// skipIfNoITerm2 skips the test when the ITERM2_E2E env is not "1".
func skipIfNoITerm2(t *testing.T) {
	t.Helper()
	if os.Getenv("ITERM2_E2E") != "1" {
		t.Skip("set ITERM2_E2E=1 to run e2e tests against real iTerm2")
	}
}

// connectAndGetApp connects to iTerm2 and returns the full App hierarchy.
// The connection is auto-closed when the callback returns.
func connectAndGetApp(t *testing.T, scriptName string) *term2go.App {
	t.Helper()
	var app *term2go.App
	ctx := context.Background()
	err := term2go.Run(ctx, scriptName, func(caller term2go.Caller) error {
		var appErr error
		app, appErr = term2go.GetApp(ctx, caller)
		return appErr
	})
	require.NoError(t, err, "connectAndGetApp")
	require.NotNil(t, app, "expected non-nil App")
	return app
}

// firstWindow returns the first window from the App, or fails the test.
func firstWindow(t *testing.T, app *term2go.App) *term2go.Window {
	t.Helper()
	require.NotEmpty(t, app.Windows, "no windows available")
	return app.Windows[0]
}

// firstSession returns the first session from the first window/tab, or fails the test.
func firstSession(t *testing.T, app *term2go.App) *term2go.Session {
	t.Helper()
	w := firstWindow(t, app)
	require.NotEmpty(t, w.Tabs, "no tabs in first window")
	sessions := w.Tabs[0].Root.Sessions()
	require.NotEmpty(t, sessions, "no sessions in first tab")
	return sessions[0]
}

// firstTab returns the first tab from the first window, or fails the test.
func firstTab(t *testing.T, app *term2go.App) *term2go.Tab {
	t.Helper()
	w := firstWindow(t, app)
	require.NotEmpty(t, w.Tabs, "no tabs in first window")
	return w.Tabs[0]
}

// runWithCaller is a shorthand for one-off RPC calls that don't need a persistent App.
func runWithCaller(t *testing.T, scriptName string, fn func(context.Context, term2go.Caller) error) {
	t.Helper()
	ctx := context.Background()
	err := term2go.Run(ctx, scriptName, func(caller term2go.Caller) error {
		return fn(ctx, caller)
	})
	require.NoError(t, err, "runWithCaller(%q)", scriptName)
}

// connectForNotifications connects and returns a Connection for long-lived notification tests.
// The caller is responsible for calling conn.Close().
func connectForNotifications(t *testing.T, scriptName string) *term2go.Connection {
	t.Helper()
	conn, err := term2go.Connect(context.Background(), scriptName)
	require.NoError(t, err, "Connect")
	return conn
}
