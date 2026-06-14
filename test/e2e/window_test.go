package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// findWindowAndSession locates a Window by ID within an App and returns it
// along with its first Session. Fails the test if the window is not found
// or has no sessions.
func findWindowAndSession(t *testing.T, app *term2go.App, windowID string) (*term2go.Window, *term2go.Session) {
	t.Helper()
	for _, w := range app.Windows {
		if w.ID == windowID {
			require.NotEmpty(t, w.Tabs, "window %s has no tabs", windowID)
			sessions := w.Tabs[0].Root.Sessions()
			require.NotEmpty(t, sessions, "window %s tab 0 has no sessions", windowID)
			return w, sessions[0]
		}
	}
	t.Fatalf("window %s not found in App", windowID)
	return nil, nil
}

// TestE2E_Window_CreateTab creates a new tab in the first window with an
// empty profile name, then verifies the tab exists and has sessions.
// Cleanup closes the tab with force.
func TestE2E_Window_CreateTab(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.window_createtab", func(ctx context.Context, caller term2go.Caller) error {
		app, err := term2go.GetApp(ctx, caller)
		require.NoError(t, err, "GetApp")

		w := firstWindow(t, app)

		tab, err := w.CreateTab(ctx, "")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}
		defer func() {
			_ = tab.Close(ctx, term2go.WithCloseForce(true))
		}()

		assert.NotEmpty(t, tab.ID, "created tab should have an ID")
		require.NotNil(t, tab.Root, "created tab should have a Root splitter")
		sessions := tab.Root.Sessions()
		assert.NotEmpty(t, sessions, "created tab should have sessions")

		// Set session name and badge
		if len(sessions) > 0 {
			s := sessions[0]
			require.NoError(t, s.SetName(ctx, "term2go-e2e-createtab"), "SetName")
			require.NoError(t, s.SetBadge(ctx, "e2e"), "SetBadge")
			t.Logf("Session name/badge set: id=%s", s.ID)
		}

		t.Logf("Created tab: id=%s sessions=%d", tab.ID, len(sessions))
		return nil
	})
}

// TestE2E_Window_CreateTabProfile creates a new tab in the first window with
// the "Default" profile name, then verifies the tab exists and has sessions.
func TestE2E_Window_CreateTabProfile(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.window_createtabprofile", func(ctx context.Context, caller term2go.Caller) error {
		app, err := term2go.GetApp(ctx, caller)
		require.NoError(t, err, "GetApp")

		w := firstWindow(t, app)

		tab, err := w.CreateTab(ctx, "Default")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}
		defer func() {
			_ = tab.Close(ctx, term2go.WithCloseForce(true))
		}()

		assert.NotEmpty(t, tab.ID, "created tab should have an ID")
		require.NotNil(t, tab.Root, "created tab should have a Root splitter")
		sessions := tab.Root.Sessions()
		assert.NotEmpty(t, sessions, "created tab should have sessions")

		// Set session name and badge
		if len(sessions) > 0 {
			s := sessions[0]
			require.NoError(t, s.SetName(ctx, "term2go-e2e-createtab-profile"), "SetName")
			require.NoError(t, s.SetBadge(ctx, "e2e-profile"), "SetBadge")
			t.Logf("Session name/badge set: id=%s", s.ID)
		}

		t.Logf("Created tab: id=%s profile=Default sessions=%d", tab.ID, len(sessions))
		return nil
	})
}

// TestE2E_Window_Close finds the highest-numbered window, force-closes it,
// refreshes the hierarchy, and verifies the window count decreased by 1.
// Skips if there is only 1 window to avoid closing the user's last window.
func TestE2E_Window_Close(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "com.term2go.test.window_close")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")

	if len(app.Windows) <= 1 {
		t.Skip("need more than 1 window to test Close (would close user's last window)")
	}

	// Find highest-numbered window
	var target *term2go.Window
	var maxNum int32 = -1
	for _, w := range app.Windows {
		if w.Number > maxNum {
			maxNum = w.Number
			target = w
		}
	}
	require.NotNil(t, target, "should find a window to close")

	oldCount := len(app.Windows)
	t.Logf("Closing window: id=%s number=%d (total windows before: %d)", target.ID, target.Number, oldCount)

	err = target.Close(ctx, term2go.WithCloseForce(true))
	require.NoError(t, err, "Close window")

	err = app.Refresh(ctx)
	require.NoError(t, err, "Refresh after close")

	assert.Equal(t, oldCount-1, len(app.Windows), "window count should decrease by 1")
	t.Logf("Window count after close: %d", len(app.Windows))
}

// TestE2E_Window_CreateWindowVariable creates 2 new windows (one with default
// options, one with WithTabIndex(0)), sets user variables on the first window's
// session, reads them back via model methods, and batch-reads built-in variables
// via the RPC layer. All created windows are cleaned up with defer.
func TestE2E_Window_CreateWindowVariable(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	err := term2go.Run(ctx, "term2go-e2e-window-var", func(caller term2go.Caller) error {
		// Create 2 windows
		resp1, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("CreateTab 1: %w", err)
		}
		resp2, err := term2go.CreateTab(ctx, caller, "", "Default", term2go.WithTabIndex(0))
		if err != nil {
			return fmt.Errorf("CreateTab 2: %w", err)
		}

		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("GetApp: %w", err)
		}

		w1, s1 := findWindowAndSession(t, app, resp1.GetWindowId())
		w2, _ := findWindowAndSession(t, app, resp2.GetWindowId())

		defer func() {
			_ = w1.Close(ctx, term2go.WithCloseForce(true))
			_ = w2.Close(ctx, term2go.WithCloseForce(true))
		}()

		t.Logf("Window 1: id=%s session=%s", w1.ID, s1.ID)
		t.Logf("Window 2: id=%s (WithTabIndex=0)", w2.ID)

		// Set session name and badge on first window
		require.NoError(t, s1.SetName(ctx, "term2go-e2e-var"), "SetName")
		require.NoError(t, s1.SetBadge(ctx, "e2e-var"), "SetBadge")
		t.Logf("Session name/badge set: %s", s1.ID)

		// Set user variables on first window's session
		require.NoError(t, s1.SetVariable(ctx, "user.test1", "val1"), "SetVariable user.test1")
		require.NoError(t, s1.SetVariable(ctx, "user.test2", "val2"), "SetVariable user.test2")

		// Read them back via model method (auto-decoded)
		v1, err := s1.GetVariable(ctx, "user.test1")
		require.NoError(t, err, "GetVariable user.test1")
		assert.Equal(t, "val1", v1)

		v2, err := s1.GetVariable(ctx, "user.test2")
		require.NoError(t, err, "GetVariable user.test2")
		assert.Equal(t, "val2", v2)

		// Read built-in variables via RPC batch
		builtinNames := []string{"jobName", "hostName", "path", "columns", "rows"}
		values, err := term2go.GetVariable(ctx, caller, s1.ID, builtinNames)
		require.NoError(t, err, "GetVariable builtins")
		assert.Len(t, values, len(builtinNames))

		for i, name := range builtinNames {
			assert.NotEmpty(t, values[i], "builtin %s should not be empty", name)
			t.Logf("  %s = %s", name, values[i])
		}

		return nil
	})
	require.NoError(t, err, "Run")
}

// TestE2E_Window_CreateWindowFull creates a new window, sets 8 user variables
// individually and reads them back, batch-reads 10 built-in variables, verifies
// the Tab/Session hierarchy, batch-sets 3 user variables and batch-reads them
// via the RPC layer. Never operates on existing user sessions. Cleanup with defer.
func TestE2E_Window_CreateWindowFull(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	err := term2go.Run(ctx, "term2go-e2e-window-full", func(caller term2go.Caller) error {
		// Create a new window
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}

		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("GetApp: %w", err)
		}

		w, s := findWindowAndSession(t, app, resp.GetWindowId())
		defer func() {
			_ = w.Close(ctx, term2go.WithCloseForce(true))
		}()

		t.Logf("Created window: id=%s session=%s", w.ID, s.ID)

		// Set session name and badge
		require.NoError(t, s.SetName(ctx, "term2go-e2e-full"), "SetName")
		require.NoError(t, s.SetBadge(ctx, "e2e-full"), "SetBadge")
		t.Logf("Session name/badge set: %s", s.ID)

		// --- Set 8 user variables individually ---
		userVars := map[string]string{
			"user.a": "apple",
			"user.b": "banana",
			"user.c": "cherry",
			"user.d": "date",
			"user.e": "elderberry",
			"user.f": "fig",
			"user.g": "grape",
			"user.h": "honeydew",
		}
		for name, val := range userVars {
			if err = s.SetVariable(ctx, name, val); err != nil {
				return fmt.Errorf("SetVariable %s: %w", name, err)
			}
		}
		t.Logf("Set %d user variables individually", len(userVars))

		// --- Read them back via model method (auto-decoded) ---
		for name, expected := range userVars {
			got, err := s.GetVariable(ctx, name)
			require.NoError(t, err, "GetVariable %s", name)
			assert.Equal(t, expected, got, "variable %s", name)
		}
		t.Logf("Read back all %d user variables", len(userVars))

		// --- Batch-read 10 built-in variables via RPC ---
		builtins := []string{
			"jobName",
			"hostName",
			"path",
			"columns",
			"rows",
			"session.terminalFile",
			"session.name",
			"session.tty",
			"session.hostname",
			"session.username",
		}
		values, err := term2go.GetVariable(ctx, caller, s.ID, builtins)
		require.NoError(t, err, "GetVariable builtins batch")
		assert.Len(t, values, len(builtins))

		for i, name := range builtins {
			assert.NotEmpty(t, values[i], "builtin %s should not be empty", name)
			t.Logf("  builtin %s = %s", name, values[i])
		}

		// --- Verify Tab/Session hierarchy ---
		require.NotEmpty(t, w.Tabs, "window should have tabs")
		firstTab := w.Tabs[0]
		assert.NotEmpty(t, firstTab.ID, "tab should have an ID")
		require.NotNil(t, firstTab.Root, "tab should have a Root splitter")
		sessions := firstTab.Root.Sessions()
		assert.NotEmpty(t, sessions, "tab should have sessions")
		assert.Equal(t, s.ID, sessions[0].ID, "first session should match the created session")
		t.Logf("Tab hierarchy: tab.id=%s sessions=%d session.id=%s", firstTab.ID, len(sessions), sessions[0].ID)

		// --- Batch-set 3 user variables via RPC ---
		batchVars := []struct{ name, value string }{
			{"user.batch1", "batch-val-1"},
			{"user.batch2", "batch-val-2"},
			{"user.batch3", "batch-val-3"},
		}
		for _, v := range batchVars {
			if err = term2go.SetVariable(ctx, caller, s.ID, v.name, v.value); err != nil {
				return fmt.Errorf("SetVariable %s: %w", v.name, err)
			}
		}
		t.Logf("Batch-set 3 user variables via RPC")

		// --- Batch-read them via RPC (returns JSON-encoded values) ---
		batchNames := []string{"user.batch1", "user.batch2", "user.batch3"}
		batchValues, err := term2go.GetVariable(ctx, caller, s.ID, batchNames)
		require.NoError(t, err, "GetVariable batch read")
		assert.Len(t, batchValues, len(batchNames))

		for i, name := range batchNames {
			assert.NotEmpty(t, batchValues[i], "batch variable %s should not be empty", name)
			t.Logf("  batch %s = %s", name, batchValues[i])
		}

		return nil
	})
	require.NoError(t, err, "Run")
}
