package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

func TestE2E_Screenshot_Window(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.screenshot_window", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}

		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("GetApp: %w", err)
		}

		w, s := findWindowAndSession(t, app, resp.GetWindowId())
		defer func() { _ = w.Close(ctx, term2go.WithCloseForce(true)) }()

		require.NoError(t, s.SetName(ctx, "term2go-e2e-shot"), "SetName")
		require.NoError(t, s.SetBadge(ctx, "E2E"), "SetBadge")
		require.NoError(t, s.SendText(ctx, "echo Hello_from_Screenshot_Window\n"), "SendText")
		time.Sleep(500 * time.Millisecond)

		path := filepath.Join(t.TempDir(), "window.png")
		err = w.Screenshot(ctx, path)
		require.NoError(t, err, "Window.Screenshot")

		info, err := os.Stat(path)
		require.NoError(t, err, "os.Stat")
		assert.Greater(t, info.Size(), int64(1024), "screenshot should be > 1KB")
		t.Logf("Window screenshot: %s (%d bytes)", path, info.Size())
		return nil
	})
}

func TestE2E_Screenshot_Tab(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.screenshot_tab", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}

		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("GetApp: %w", err)
		}

		w, s := findWindowAndSession(t, app, resp.GetWindowId())
		defer func() { _ = w.Close(ctx, term2go.WithCloseForce(true)) }()

		require.NoError(t, s.SetName(ctx, "term2go-e2e-shot-tab"), "SetName")
		require.NoError(t, s.SetBadge(ctx, "TAB"), "SetBadge")
		require.NoError(t, s.SendText(ctx, "echo Tab_Screenshot\n"), "SendText")
		time.Sleep(300 * time.Millisecond)

		tab := w.Tabs[0]
		path := filepath.Join(t.TempDir(), "tab.png")
		err = tab.Screenshot(ctx, path)
		require.NoError(t, err, "Tab.Screenshot")

		info, err := os.Stat(path)
		require.NoError(t, err, "os.Stat")
		assert.Greater(t, info.Size(), int64(1024), "screenshot should be > 1KB")
		t.Logf("Tab screenshot: %s (%d bytes)", path, info.Size())
		return nil
	})
}

func TestE2E_Screenshot_Session(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.screenshot_session", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.CreateTab(ctx, caller, "", "Default")
		if err != nil {
			return fmt.Errorf("CreateTab: %w", err)
		}

		app, err := term2go.GetApp(ctx, caller)
		if err != nil {
			return fmt.Errorf("GetApp: %w", err)
		}

		w, s := findWindowAndSession(t, app, resp.GetWindowId())
		defer func() { _ = w.Close(ctx, term2go.WithCloseForce(true)) }()

		require.NoError(t, s.SetName(ctx, "term2go-e2e-shot-sess"), "SetName")
		require.NoError(t, s.SetBadge(ctx, "SES"), "SetBadge")
		require.NoError(t, s.SendText(ctx, "echo Session_Screenshot\n"), "SendText")
		time.Sleep(300 * time.Millisecond)

		path := filepath.Join(t.TempDir(), "tab.png")
		err = s.Screenshot(ctx, path)
		require.NoError(t, err, "Session.Screenshot")

		info, err := os.Stat(path)
		require.NoError(t, err, "os.Stat")
		assert.Greater(t, info.Size(), int64(1024), "screenshot should be > 1KB")
		t.Logf("Session screenshot: %s (%d bytes)", path, info.Size())
		return nil
	})
}
