package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

func TestE2E_Session_GetBuffer(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-session-getbuffer")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")
	s := firstSession(t, app)

	buf, err := s.GetBuffer(ctx, nil)
	require.NoError(t, err, "GetBuffer")

	if buf == nil {
		t.Log("GetBuffer returned nil (may need active content in session)")
	} else {
		sc := term2go.NewScreenContents(buf)
		require.NotNil(t, sc, "ScreenContents should not be nil")
		lines := sc.LineCount()
		t.Logf("Buffer lines: %d", lines)
		assert.Greater(t, lines, 0, "buffer should contain at least one line")
	}
}

func TestE2E_Session_SplitPane(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-session-splitpane")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")
	s := firstSession(t, app)

	// Note: iTerm2 may not always return session IDs in SplitPane response
	// (documented as known behavior for tmux integration).
	resp, err := term2go.SplitPane(ctx, conn, s.ID, true, false, "")
	require.NoError(t, err, "SplitPane RPC should succeed")

	ids := resp.GetSessionId()
	if len(ids) > 0 {
		newSessionID := ids[0]
		t.Logf("New session: %s", newSessionID)
		assert.NotEqual(t, s.ID, newSessionID, "new split session should have unique ID")

		// Clean up
		err = term2go.Close(ctx, conn, newSessionID)
		assert.NoError(t, err, "Close new session")
	} else {
		t.Log("SplitPane did not return session IDs (known iTerm2 quirk)")
	}
}

func TestE2E_Session_Activate(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-session-activate")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")
	s := firstSession(t, app)

	err = term2go.Activate(ctx, conn, s.GetID(), true, true)
	require.NoError(t, err, "Activate")

	focusResp, err := term2go.FocusRequest(ctx, conn)
	require.NoError(t, err, "FocusRequest")
	require.NotNil(t, focusResp, "FocusRequest should return non-nil response after activation")
}

func TestE2E_Session_BasicInfo(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-session-basicinfo")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")
	s := firstSession(t, app)

	assert.NotEmpty(t, s.ID, "session ID should not be empty")
	assert.Equal(t, s.ID, s.GetID(), "GetID() should match ID field")
}

func TestE2E_Session_Restart(t *testing.T) {
	skipIfNoITerm2(t)

	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-session-restart")
	require.NoError(t, err, "Connect")
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err, "GetApp")
	s := firstSession(t, app)

	resp, err := term2go.SplitPane(ctx, conn, s.ID, true, false, "")
	require.NoError(t, err, "SplitPane")

	ids := resp.GetSessionId()
	if len(ids) == 0 {
		t.Skip("SplitPane did not return session IDs, cannot test Restart")
	}

	newSessionID := ids[0]
	defer func() {
		_ = term2go.Close(ctx, conn, newSessionID)
	}()

	// RestartSessionIfExited is a no-op when the session is still running.
	err = term2go.RestartSessionIfExited(ctx, conn, newSessionID)
	require.NoError(t, err, "RestartSessionIfExited should succeed on a running session")
}
