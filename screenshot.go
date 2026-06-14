package term2go

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// screenshot captures the frontmost iTerm2 window to a PNG file.
// activate is called first to bring the target window to the front.
func screenshot(ctx context.Context, path string, activate func(ctx context.Context) error) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".png") {
		return fmt.Errorf("output path must end with .png: %s", path)
	}

	// Bring the target window to the front
	if err = activate(ctx); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	// Wait for window to settle at the front of the z-order
	time.Sleep(200 * time.Millisecond)

	// The activated window is now window 1 in iTerm2's AppleScript hierarchy.
	// iTerm2's AppleScript 'id of window' returns the native CGWindowID,
	// which is exactly what screencapture -l expects.
	var wid string
	wid, err = getITerm2WindowID(ctx, 1)
	if err != nil {
		return fmt.Errorf("get window ID: %w", err)
	}

	// Capture by CGWindowID (-o = no window shadow)
	cmd := exec.CommandContext(ctx, "screencapture", "-l"+wid, "-x", "-o", abs)
	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("screencapture: %w\n%s", err, string(output))
	}
	return nil
}

// getITerm2WindowID returns the native CGWindowID of the n-th iTerm2 window.
// n=1 is the frontmost (activated) window. Uses iTerm2's AppleScript support.
func getITerm2WindowID(ctx context.Context, n int) (string, error) {
	script := fmt.Sprintf(`tell application "iTerm2" to id of window %d`, n)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("osascript: %w", err)
	}
	wid := strings.TrimSpace(string(output))
	if wid == "" {
		return "", fmt.Errorf("no iTerm2 window found")
	}
	return wid, nil
}

// ---------------------------------------------------------------
// Window.Screenshot
// ---------------------------------------------------------------

// Screenshot captures this window and saves it as a PNG file.
func (w *Window) Screenshot(ctx context.Context, path string) error {
	app, err := GetApp(ctx, w.caller)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	for _, win := range app.Windows {
		if win.ID == w.ID {
			if len(win.Tabs) == 0 {
				return fmt.Errorf("window has no tabs: %s", w.ID)
			}
			return screenshot(ctx, path, func(ctx context.Context) error {
				return win.Tabs[0].Select(ctx)
			})
		}
	}
	return fmt.Errorf("window not found: %s", w.ID)
}

// ---------------------------------------------------------------
// Tab.Screenshot
// ---------------------------------------------------------------

// Screenshot captures this tab's containing window and saves it as a PNG file.
func (t *Tab) Screenshot(ctx context.Context, path string) error {
	app, err := GetApp(ctx, t.caller)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	for _, w := range app.Windows {
		for _, tab := range w.Tabs {
			if tab.ID == t.ID {
				return screenshot(ctx, path, func(ctx context.Context) error {
					return tab.Select(ctx)
				})
			}
		}
	}
	return fmt.Errorf("tab not found in any window: %s", t.ID)
}

// ---------------------------------------------------------------
// Session.Screenshot
// ---------------------------------------------------------------

// Screenshot captures this session's containing window and saves it as a PNG file.
func (s *Session) Screenshot(ctx context.Context, path string) error {
	app, err := GetApp(ctx, s.caller)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	for _, w := range app.Windows {
		for _, tab := range w.Tabs {
			for _, session := range tab.Root.Sessions() {
				if session.ID == s.ID {
					return screenshot(ctx, path, func(ctx context.Context) error {
						return tab.Select(ctx)
					})
				}
			}
		}
	}
	return fmt.Errorf("session not found in any window: %s", s.ID)
}
