package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

// TestE2E_App_GetApp retrieves the full App hierarchy and verifies that
// Windows is non-empty. It also logs every window/tab/session for visibility.
func TestE2E_App_GetApp(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.app_getapp")
	require.NotEmpty(t, app.Windows, "expected at least one window")

	for i, w := range app.Windows {
		t.Logf("Window %d: id=%s number=%d tabs=%d", i, w.ID, w.Number, len(w.Tabs))
		for j, tb := range w.Tabs {
			t.Logf("  Tab %d: id=%s", j, tb.ID)
			sessions := tb.Root.Sessions()
			t.Logf("    Sessions: %d", len(sessions))
			for k, s := range sessions {
				t.Logf("      Session %d: id=%s", k, s.ID)
			}
		}
	}
}

// TestE2E_App_Refresh calls Refresh and verifies the window count is
// unchanged after refreshing.
func TestE2E_App_Refresh(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "com.term2go.test.app_refresh")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")

	count := len(app.Windows)
	require.NotZero(t, count, "expected at least one window before refresh")

	err = app.Refresh(ctx)
	require.NoError(t, err, "Refresh")

	assert.Equal(t, count, len(app.Windows), "window count should not change after Refresh")
}

// TestE2E_App_WindowProps validates the properties of the first Window:
// ID is non-empty, Number is non-negative, Frame is non-nil, Tabs is non-empty.
func TestE2E_App_WindowProps(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.app_windowprops")
	w := firstWindow(t, app)

	assert.NotEmpty(t, w.ID, "Window.ID should not be empty")
	assert.GreaterOrEqual(t, w.Number, int32(0), "Window.Number should be >= 0")
	assert.NotNil(t, w.Frame, "Window.Frame should not be nil")
	assert.NotEmpty(t, w.Tabs, "Window.Tabs should not be empty")

	assert.Greater(t, w.Frame.Size.Width, int32(0), "Frame.Size.Width should be > 0")
	assert.Greater(t, w.Frame.Size.Height, int32(0), "Frame.Size.Height should be > 0")

	t.Logf("Window: id=%s number=%d tabs=%d frame=(%d,%d %dx%d)",
		w.ID, w.Number, len(w.Tabs),
		w.Frame.Origin.X, w.Frame.Origin.Y,
		w.Frame.Size.Width, w.Frame.Size.Height)

	// At least one session exists
	firstTab := w.Tabs[0]
	sessions := firstTab.Root.Sessions()
	require.NotEmpty(t, sessions, "first tab must have at least one session")
	t.Logf("First session: id=%s", sessions[0].ID)
}

// TestE2E_App_ListSessions calls the raw ListSessions RPC and verifies
// the response contains windows, tabs, and sessions.
func TestE2E_App_ListSessions(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.app_listsessions", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.ListSessions(ctx, caller)
		require.NoError(t, err, "ListSessions")
		require.NotNil(t, resp, "ListSessions response should not be nil")

		windows := resp.GetWindows()
		require.NotEmpty(t, windows, "expected at least one window in ListSessions response")

		totalTabs := 0
		totalSessions := 0

		for i, w := range windows {
			t.Logf("Window %d: id=%s number=%d", i, w.GetWindowId(), w.GetNumber())
			tabs := w.GetTabs()
			require.NotEmpty(t, tabs, fmt.Sprintf("window %s should have at least one tab", w.GetWindowId()))

			for j, tb := range tabs {
				totalTabs++
				t.Logf("  Tab %d: id=%s", j, tb.GetTabId())
				require.NotNil(t, tb.Root, fmt.Sprintf("tab %s should have root", tb.GetTabId()))

				sessions := collectSessionsFromSplitTree(tb.Root)
				require.NotEmpty(t, sessions, fmt.Sprintf("tab %s should have at least one session", tb.GetTabId()))

				for k, s := range sessions {
					totalSessions++
					t.Logf("    Session %d: id=%s title=%s", k, s.GetUniqueIdentifier(), s.GetTitle())
				}
			}
		}

		t.Logf("Total: %d windows, %d tabs, %d sessions", len(windows), totalTabs, totalSessions)
		assert.Greater(t, totalTabs, 0, "expected at least one tab")
		assert.Greater(t, totalSessions, 0, "expected at least one session")

		return nil
	})
}

// collectSessionsFromSplitTree recursively extracts SessionSummary entries
// from a SplitTreeNode's link tree.
func collectSessionsFromSplitTree(node *iterm2.SplitTreeNode) []*iterm2.SessionSummary {
	var result []*iterm2.SessionSummary
	for _, link := range node.GetLinks() {
		if s := link.GetSession(); s != nil {
			result = append(result, s)
		}
		if n := link.GetNode(); n != nil {
			result = append(result, collectSessionsFromSplitTree(n)...)
		}
	}
	return result
}
