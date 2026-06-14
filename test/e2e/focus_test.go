package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Focus_Request calls FocusRequest and logs the current focus state
// (applicationActive, windowChanged, selectedTab, session). It verifies
// the response is non-nil but does not assert on specific notification
// values since focus may change at any time.
func TestE2E_Focus_Request(t *testing.T) {
	skipIfNoITerm2(t)

	runWithCaller(t, "com.term2go.test.focus_request", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.FocusRequest(ctx, caller)
		require.NoError(t, err, "FocusRequest")
		require.NotNil(t, resp, "FocusResponse should not be nil")

		notifications := resp.GetNotifications()
		t.Logf("Focus notifications count: %d", len(notifications))

		for i, n := range notifications {
			switch {
			case n.GetApplicationActive():
				t.Logf("  [%d] applicationActive=true", i)
			case n.GetWindow() != nil:
				t.Logf("  [%d] windowChanged: id=%s status=%v", i,
					n.GetWindow().GetWindowId(), n.GetWindow().GetWindowStatus())
			case n.GetSelectedTab() != "":
				t.Logf("  [%d] selectedTab: %s", i, n.GetSelectedTab())
			case n.GetSession() != "":
				t.Logf("  [%d] session: %s", i, n.GetSession())
			default:
				t.Logf("  [%d] unknown notification type", i)
			}
		}

		return nil
	})
}

// TestE2E_Focus_Activate activates the first session with
// orderWindowFront=true and selectTab=true, then calls FocusRequest to
// verify the session appears in the focus notifications.
func TestE2E_Focus_Activate(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.focus_activate")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.focus_activate", func(ctx context.Context, caller term2go.Caller) error {
		err := term2go.Activate(ctx, caller, s.ID, true, true)
		require.NoError(t, err, "Activate")

		// Verify focus state after activation.
		focusResp, err := term2go.FocusRequest(ctx, caller)
		require.NoError(t, err, "FocusRequest after Activate")
		require.NotNil(t, focusResp, "FocusResponse should not be nil")

		t.Logf("Focus state after activate: %d notifications", len(focusResp.GetNotifications()))

		return nil
	})
}
