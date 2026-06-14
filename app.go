package term2go

import (
	"context"
	"fmt"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// App
// ============================================================================

// App represents the iTerm2 application. It holds all terminal windows and
// provides the entry point for navigating the session hierarchy.
type App struct {
	caller  Caller
	Windows []*Window
}

// GetApp retrieves the full iTerm2 session hierarchy by calling ListSessions
// and constructing the object tree from the response.
func GetApp(ctx context.Context, caller Caller) (*App, error) {
	resp, err := ListSessions(ctx, caller)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return appFromListSessionsResponse(caller, resp), nil
}

func appFromListSessionsResponse(caller Caller, resp *iterm2.ListSessionsResponse) *App {
	windows := make([]*Window, 0, len(resp.GetWindows()))
	for _, pw := range resp.GetWindows() {
		if w := windowFromProto(caller, pw); w != nil {
			windows = append(windows, w)
		}
	}
	return &App{caller: caller, Windows: windows}
}

// Refresh reloads the full window/tab/session hierarchy from iTerm2.
func (a *App) Refresh(ctx context.Context) error {
	resp, err := ListSessions(ctx, a.caller)
	if err != nil {
		return fmt.Errorf("refresh: %w", err)
	}
	newApp := appFromListSessionsResponse(a.caller, resp)
	a.Windows = newApp.Windows
	return nil
}
