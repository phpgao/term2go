package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

// getSessionID connects, gets the first session ID, and returns it.
func getSessionID(t *testing.T, scriptName string) string {
	t.Helper()
	app := connectAndGetApp(t, scriptName)
	return firstSession(t, app).ID
}

// ============================================================================
// TestE2E_Notify_NewSession — trigger: create a new tab
// ============================================================================

func TestE2E_Notify_NewSession(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetNewSessionNotification(); n != nil {
			t.Logf("new session notification: %s", n.GetSessionId())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeNewSession(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.NewSessionNotification) {
		t.Logf("callback: new session %s", n.GetSessionId())
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: create a new tab
	app, _ := term2go.GetApp(ctx, conn)
	if app != nil && len(app.Windows) > 0 {
		tab, err := app.Windows[0].CreateTab(ctx, "Default")
		if err == nil {
			defer tab.Close(ctx)
			t.Logf("trigger: created tab %s", tab.ID)
		}
	}

	select {
	case <-received:
		t.Log("received new session notification")
	case <-time.After(time.Second):
		t.Log("no new session notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_TerminateSession — trigger: create + close a tab
// ============================================================================

func TestE2E_Notify_TerminateSession(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetTerminateSessionNotification(); n != nil {
			t.Logf("terminate session notification: %s", n.GetSessionId())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeTerminateSession(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.TerminateSessionNotification) {
		t.Logf("callback: terminate %s", n.GetSessionId())
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: create and immediately close a tab
	app, _ := term2go.GetApp(ctx, conn)
	if app != nil && len(app.Windows) > 0 {
		tab, err := app.Windows[0].CreateTab(ctx, "Default")
		if err == nil {
			_ = tab.Close(ctx)
			t.Logf("trigger: created and closed tab %s", tab.ID)
		}
	}

	select {
	case <-received:
		t.Log("received terminate session notification")
	case <-time.After(time.Second):
		t.Log("no terminate session notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_Keystroke — trigger: send text
// ============================================================================

func TestE2E_Notify_Keystroke(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetKeystrokeNotification(); n != nil {
			t.Logf("keystroke: %q", n.GetCharacters())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeKeystroke(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.KeystrokeNotification) {
		t.Logf("callback: keystroke %q", n.GetCharacters())
	}, sessionID)
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: send a key
	_ = term2go.SendText(ctx, conn, sessionID, "x")

	select {
	case <-received:
		t.Log("received keystroke notification")
	case <-time.After(time.Second):
		t.Log("no keystroke notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_ScreenUpdate — trigger: send text triggers screen update
// ============================================================================

func TestE2E_Notify_ScreenUpdate(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetScreenUpdateNotification(); n != nil {
			t.Logf("screen update: session=%s", n.GetSession())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeScreenUpdate(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.ScreenUpdateNotification) {
		t.Logf("callback: screen update %s", n.GetSession())
	}, sessionID)
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: send text causes screen to update
	_ = term2go.SendText(ctx, conn, sessionID, "y")

	select {
	case <-received:
		t.Log("received screen update notification")
	case <-time.After(time.Second):
		t.Log("no screen update notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_Prompt — trigger: send a command that produces a prompt
// ============================================================================

func TestE2E_Notify_Prompt(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetPromptNotification(); n != nil {
			t.Logf("prompt notification: session=%s", n.GetSession())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribePrompt(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.PromptNotification) {
		t.Logf("callback: prompt %s", n.GetSession())
	}, sessionID)
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: run a command that ends with a prompt (echo empty line)
	// Note: this may produce content in the user's session; keep it minimal
	_ = term2go.SendText(ctx, conn, sessionID, "echo -n\n")

	select {
	case <-received:
		t.Log("received prompt notification")
	case <-time.After(2 * time.Second):
		t.Log("no prompt notification within 2s")
	}
}

// ============================================================================
// TestE2E_Notify_CustomEscapeSequence — hard to trigger, short wait
// ============================================================================

func TestE2E_Notify_CustomEscapeSequence(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	tk, err := term2go.SubscribeCustomEscapeSequence(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.CustomEscapeSequenceNotification) {
		t.Logf("callback: custom escape %s", n.GetSession())
	}, sessionID)
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	t.Log("subscribe/unsubscribe succeeded")
}

// ============================================================================
// TestE2E_Notify_VariableChange — trigger: set a variable
// ============================================================================

func TestE2E_Notify_VariableChange(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetVariableChangedNotification(); n != nil {
			t.Logf("variable change: name=%s", n.GetName())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeVariableChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.VariableChangedNotification) {
		t.Logf("callback: variable %s changed", n.GetName())
	}, sessionID, "user.e2e_notify_test")
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: set a variable
	_ = term2go.SetVariable(ctx, conn, sessionID, "user.e2e_notify_test", "triggered")

	select {
	case <-received:
		t.Log("received variable change notification")
	case <-time.After(time.Second):
		t.Log("no variable change notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_LayoutChange — trigger: split pane
// ============================================================================

func TestE2E_Notify_LayoutChange(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetLayoutChangedNotification(); n != nil {
			t.Logf("layout change notification")
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeLayoutChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.LayoutChangedNotification) {
		t.Logf("callback: layout change")
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: split pane (no session ID needed in response)
	resp, _ := term2go.SplitPane(ctx, conn, sessionID, true, false, "")
	if ids := resp.GetSessionId(); len(ids) > 0 {
		defer term2go.Close(ctx, conn, ids[0])
		t.Logf("trigger: split pane created %s", ids[0])
	}

	select {
	case <-received:
		t.Log("received layout change notification")
	case <-time.After(time.Second):
		t.Log("no layout change notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_FocusChange — trigger: activate a session
// ============================================================================

func TestE2E_Notify_FocusChange(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetFocusChangedNotification(); n != nil {
			t.Logf("focus change notification")
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeFocusChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.FocusChangedNotification) {
		t.Logf("callback: focus change")
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: activate the session
	_ = term2go.Activate(ctx, conn, sessionID, true, true)
	t.Logf("trigger: activated session %s", sessionID)

	select {
	case <-received:
		t.Log("received focus change notification")
	case <-time.After(time.Second):
		t.Log("no focus change notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_ServerOriginatedRPC — no trigger, verify subscribe only
// ============================================================================

func TestE2E_Notify_ServerOriginatedRPC(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	tk, err := term2go.SubscribeServerOriginatedRPC(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.ServerOriginatedRPCNotification) {
		t.Logf("callback: server RPC %s", n.GetRequestId())
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	t.Log("subscribe/unsubscribe succeeded")
}

// ============================================================================
// TestE2E_Notify_BroadcastChange — no trigger, verify subscribe only
// ============================================================================

func TestE2E_Notify_BroadcastChange(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	tk, err := term2go.SubscribeBroadcastChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.BroadcastDomainsChangedNotification) {
		t.Logf("callback: broadcast change")
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	t.Log("subscribe/unsubscribe succeeded")
}

// ============================================================================
// TestE2E_Notify_ProfileChange — trigger: set profile property
// ============================================================================

func TestE2E_Notify_ProfileChange(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	sessionID := getSessionID(t, scriptName)
	t.Logf("session: %s", sessionID)

	received := make(chan struct{}, 1)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetProfileChangedNotification(); n != nil {
			t.Logf("profile change: guid=%s", n.GetGuid())
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk, err := term2go.SubscribeProfileChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.ProfileChangedNotification) {
		t.Logf("callback: profile %s changed", n.GetGuid())
	})
	require.NoError(t, err)
	defer conn.Unsubscribe(tk)

	// Trigger: read profile name and set it back (minor profile change)
	resp, _ := term2go.GetProfileProperty(ctx, conn, sessionID, []string{"Name"})
	if resp != nil && len(resp.GetProperties()) > 0 {
		name := resp.GetProperties()[0].GetJsonValue()
		_ = term2go.SetProfileProperty(ctx, conn, sessionID, "Name", name)
		t.Logf("trigger: set profile Name to current value")
	}

	select {
	case <-received:
		t.Log("received profile change notification")
	case <-time.After(time.Second):
		t.Log("no profile change notification within 1s")
	}
}

// ============================================================================
// TestE2E_Notify_MultipleSubscribe — subscribe 3, short wait
// ============================================================================

func TestE2E_Notify_MultipleSubscribe(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	scriptName := "term2go-e2e-notify-" + strings.ToLower(strings.ReplaceAll(t.Name(), "_", "-"))

	conn := connectForNotifications(t, scriptName)
	defer conn.Close()

	received := make(chan string, 10)
	dispatch := func(msg *iterm2.ServerOriginatedMessage) bool {
		n := msg.GetNotification()
		switch {
		case n.GetNewSessionNotification() != nil:
			select {
			case received <- "new_session":
			default:
			}
		case n.GetLayoutChangedNotification() != nil:
			select {
			case received <- "layout_change":
			default:
			}
		case n.GetFocusChangedNotification() != nil:
			select {
			case received <- "focus_change":
			default:
			}
		}
		return true
	}
	conn.RegisterHandler(dispatch)

	tk1, err := term2go.SubscribeNewSession(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.NewSessionNotification) {})
	require.NoError(t, err)

	tk2, err := term2go.SubscribeLayoutChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.LayoutChangedNotification) {})
	require.NoError(t, err)

	tk3, err := term2go.SubscribeFocusChange(ctx, conn, conn, func(caller term2go.Caller, n *iterm2.FocusChangedNotification) {})
	require.NoError(t, err)

	// Collect notifications for a short time.
	var notifications []string
	timeout := time.After(time.Second)
collectLoop:
	for {
		select {
		case name := <-received:
			notifications = append(notifications, name)
		case <-timeout:
			break collectLoop
		}
	}
	t.Logf("notifications: %v", notifications)

	conn.Unsubscribe(tk1)
	conn.Unsubscribe(tk2)
	conn.Unsubscribe(tk3)
	t.Log("all unsubscribed")
}
