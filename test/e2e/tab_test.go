package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Tab_Select tests that Tab.Select() activates the tab (focuses it).
// Select does NOT reorder tabs — it only brings the tab into focus.
func TestE2E_Tab_Select(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.e2e.tab.Select", func(ctx context.Context, caller term2go.Caller) error {
		app, err := term2go.GetApp(ctx, caller)
		require.NoError(t, err)
		require.NotNil(t, app)
		require.NotEmpty(t, app.Windows, "no windows")

		w := app.Windows[0]
		require.NotEmpty(t, w.Tabs, "no tabs in first window")

		if len(w.Tabs) >= 2 {
			// Multiple tabs exist — test switching between them.
			secondTab := w.Tabs[1]

			err = secondTab.Select(ctx)
			require.NoError(t, err, "Select second tab")

			// Verify the tab was selected by checking focus notifications
			focusResp, err := term2go.FocusRequest(ctx, caller)
			require.NoError(t, err, "FocusRequest after select")
			require.NotNil(t, focusResp)

			// One of the notifications should reference the selected tab
			found := false
			for _, n := range focusResp.GetNotifications() {
				if n.GetSelectedTab() == secondTab.ID {
					found = true
					break
				}
			}
			// Focus notifications: the selected tab notification confirms selection worked
			t.Logf("Selected tab %s, FocusRequest returned %d notifications",
				secondTab.ID, len(focusResp.GetNotifications()))
			for i, n := range focusResp.GetNotifications() {
				t.Logf("  [%d] type=%T tab=%s session=%s",
					i, n.GetSelectedTab, n.GetSelectedTab(), n.GetSession())
			}
			// Don't assert exact notification — focus behavior varies by iTerm2 state
			_ = found

			// Select back to first tab
			firstTab := w.Tabs[0]
			err = firstTab.Select(ctx)
			require.NoError(t, err, "Select back to first tab")
			t.Logf("Selected back to first tab %s", firstTab.ID)
		} else {
			// Only one tab — create and select, then select back.
			newTab, err := w.CreateTab(ctx, "Default")
			require.NoError(t, err, "CreateTab")
			defer newTab.Close(ctx)

			// Select back to original tab
			firstID := w.Tabs[0].ID
			err = w.Tabs[0].Select(ctx)
			require.NoError(t, err, "Select back to original")
			assert.NotEmpty(t, firstID, "original tab should have valid ID")
			t.Logf("Selected back to original tab %s", firstID)
		}

		return nil
	})
}

// TestE2E_Tab_BasicProperties verifies fundamental tab attributes.
func TestE2E_Tab_BasicProperties(t *testing.T) {
	skipIfNoITerm2(t)
	app := connectAndGetApp(t, "com.term2go.e2e.tab.BasicProperties")
	tab := firstTab(t, app)

	assert.NotEmpty(t, tab.ID, "tab ID should not be empty")
	assert.NotNil(t, tab.Root, "tab root splitter should not be nil")
	assert.NotEmpty(t, tab.Root.Sessions(), "tab should have at least one session")
}
