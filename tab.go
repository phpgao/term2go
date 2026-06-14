package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// Tab
// ============================================================================

// Tab represents an iTerm2 tab, which contains a tree of split panes.
type Tab struct {
	caller Caller
	ID     string
	Root   *Splitter
}

func tabFromProto(caller Caller, pt *iterm2.ListSessionsResponse_Tab) *Tab {
	return &Tab{
		caller: caller,
		ID:     pt.GetTabId(),
		Root:   SplitterFromProto(pt.GetRoot(), caller),
	}
}

// Select makes this tab the active tab.
func (t *Tab) Select(ctx context.Context) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ActivateRequest{
		ActivateRequest: &iterm2.ActivateRequest{
			Identifier: &iterm2.ActivateRequest_TabId{
				TabId: t.ID,
			},
			SelectTab:        proto.Bool(true),
			OrderWindowFront: proto.Bool(true),
		},
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("select tab: %w", err)
	}
	return checkError(resp)
}

// Close closes the tab.
func (t *Tab) Close(ctx context.Context, opts ...CloseOption) error {
	req := newRequest()
	closeReq := &iterm2.CloseRequest{
		Target: &iterm2.CloseRequest_Tabs{
			Tabs: &iterm2.CloseRequest_CloseTabs{
				TabIds: []string{t.ID},
			},
		},
	}
	for _, o := range opts {
		o(closeReq)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_CloseRequest{
		CloseRequest: closeReq,
	}
	resp, err := t.caller.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("close tab: %w", err)
	}
	return checkError(resp)
}

// UpdateLayout sends the current split-pane layout to iTerm2 to adjust sizes.
func (t *Tab) UpdateLayout(ctx context.Context) error {
	return SetTabLayout(ctx, t.caller, t.ID, t.Root.ToProto())
}
