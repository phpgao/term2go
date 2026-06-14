package term2go

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// PromptState
// ============================================================================

// PromptState describes the lifecycle of a shell prompt.
type PromptState int

const (
	PromptEditing  PromptState = 0 // Command is being edited
	PromptRunning  PromptState = 1 // Command is executing
	PromptFinished PromptState = 2 // Command has completed
)

// ============================================================================
// Prompt
// ============================================================================

// Prompt wraps GetPromptResponse from a shell prompt.
type Prompt struct {
	raw *iterm2.GetPromptResponse
}

// NewPrompt creates a Prompt from a proto response.
func NewPrompt(raw *iterm2.GetPromptResponse) *Prompt {
	return &Prompt{raw: raw}
}

// PromptRange returns the coordinates of the prompt text.
func (p *Prompt) PromptRange() CoordRange {
	if p.raw == nil {
		return CoordRange{}
	}
	return CoordRangeFromProto(p.raw.GetPromptRange())
}

// CommandRange returns the coordinates of the command typed by the user.
func (p *Prompt) CommandRange() CoordRange {
	if p.raw == nil {
		return CoordRange{}
	}
	return CoordRangeFromProto(p.raw.GetCommandRange())
}

// OutputRange returns the coordinates of the command output.
func (p *Prompt) OutputRange() CoordRange {
	if p.raw == nil {
		return CoordRange{}
	}
	return CoordRangeFromProto(p.raw.GetOutputRange())
}

// WorkingDirectory returns the working directory when the command ran.
func (p *Prompt) WorkingDirectory() string {
	return p.raw.GetWorkingDirectory()
}

// Command returns the text of the command.
func (p *Prompt) Command() string {
	return p.raw.GetCommand()
}

// State returns the prompt state.
func (p *Prompt) State() PromptState {
	// proto state maps 1:1 to our PromptState.
	return PromptState(p.raw.GetPromptState())
}

// ExitStatus returns the command exit code (only valid when state==Finished).
func (p *Prompt) ExitStatus() uint32 {
	return p.raw.GetExitStatus()
}

// UniqueID returns the unique prompt identifier, or "" if unavailable.
func (p *Prompt) UniqueID() string {
	return p.raw.GetUniquePromptId()
}

// ExcludedSubranges returns ranges inside the command that are not
// part of user input (e.g. copy-mode paste bracketed regions).
func (p *Prompt) ExcludedSubranges() []CoordRange {
	raw := p.raw.GetExcludedSubranges()
	out := make([]CoordRange, len(raw))
	for i, r := range raw {
		out[i] = CoordRangeFromProto(r)
	}
	return out
}

// Raw returns the underlying proto response.
func (p *Prompt) Raw() *iterm2.GetPromptResponse { return p.raw }

// ============================================================================
// Convenience functions
// ============================================================================

// GetLastPrompt retrieves the most recent prompt for a session.
// Returns nil if PROMPT_UNAVAILABLE.
func GetLastPrompt(ctx context.Context, caller Caller, sessionID string) (*Prompt, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_GetPromptRequest{
		GetPromptRequest: &iterm2.GetPromptRequest{
			Session: proto.String(sessionID),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	pr := resp.GetGetPromptResponse()
	if pr.GetStatus() == iterm2.GetPromptResponse_PROMPT_UNAVAILABLE {
		return nil, nil
	}
	return NewPrompt(pr), nil
}

// GetPromptByID retrieves a specific prompt by its unique ID.
func GetPromptByID(ctx context.Context, caller Caller, sessionID, promptID string) (*Prompt, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_GetPromptRequest{
		GetPromptRequest: &iterm2.GetPromptRequest{
			Session:        proto.String(sessionID),
			UniquePromptId: proto.String(promptID),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	pr := resp.GetGetPromptResponse()
	if pr.GetStatus() == iterm2.GetPromptResponse_PROMPT_UNAVAILABLE {
		return nil, nil
	}
	return NewPrompt(pr), nil
}

// ListPromptIDs returns a list of prompt IDs for a session, optionally
// bounded by first/last.
func ListPromptIDs(ctx context.Context, caller Caller, sessionID, first, last string) ([]string, error) {
	req := newRequest()
	lpr := &iterm2.ListPromptsRequest{
		Session: proto.String(sessionID),
	}
	if first != "" {
		lpr.FirstUniqueId = proto.String(first)
	}
	if last != "" {
		lpr.LastUniqueId = proto.String(last)
	}
	req.Submessage = &iterm2.ClientOriginatedMessage_ListPromptsRequest{
		ListPromptsRequest: lpr,
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	return resp.GetListPromptsResponse().GetUniquePromptId(), nil
}

// ============================================================================
// PromptEvent
// ============================================================================

// PromptEvent is produced by PromptMonitor on each prompt-state change.
type PromptEvent struct {
	Mode    iterm2.PromptMonitorMode // PROMPT / COMMAND_START / COMMAND_END
	Prompt  *Prompt                  // non-nil when Mode==PROMPT
	Command string                   // non-empty when Mode==COMMAND_START
	Status  int32                    // valid when Mode==COMMAND_END

	// UniquePromptID is set when the notification includes it.
	UniquePromptID string
}

// ============================================================================
// PromptMonitor
// ============================================================================

// PromptMonitor streams prompt lifecycle events.  Create one with
// NewPromptMonitor, iterate over Chan(), and call Close() when finished.
//
// Unlike ScreenStreamer, PromptMonitor does not need a run goroutine — the
// prompt data is embedded directly in the notification, so no extra RPC
// calls are needed.
type PromptMonitor struct {
	conn            *Connection
	ch              chan PromptEvent
	done            chan struct{}
	token           NotificationToken
	dispatchHandler NotificationHandler
	once            sync.Once
}

// NewPromptMonitor subscribes to prompt notifications with the given modes.
// If modes is nil, defaults to [PromptMonitorMode_PROMPT].
//
// Usage:
//
//	pm, err := NewPromptMonitor(conn, sessionID, []PromptMonitorMode{PROMPT, COMMAND_END})
//	if err != nil { ... }
//	defer pm.Close()
//	for ev := range pm.Chan() {
//	    switch ev.Mode {
//	    case PROMPT:
//	        fmt.Println("prompt:", ev.Prompt.Command())
//	    case COMMAND_END:
//	        fmt.Println("exit:", ev.Status)
//	    }
//	}
func NewPromptMonitor(conn *Connection, sessionID string, modes []iterm2.PromptMonitorMode) (*PromptMonitor, error) {
	if len(modes) == 0 {
		modes = []iterm2.PromptMonitorMode{iterm2.PromptMonitorMode_PROMPT}
	}

	pm := &PromptMonitor{
		conn: conn,
		ch:   make(chan PromptEvent, 8),
		done: make(chan struct{}),
	}

	key := "prompt:" + sessionID
	if err := notifyRPC(context.Background(), conn, true, iterm2.NotificationType_NOTIFY_ON_PROMPT, sessionID,
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_PromptMonitorRequest{
				PromptMonitorRequest: &iterm2.PromptMonitorRequest{
					Modes: modes,
				},
			}
		},
	); err != nil {
		return nil, fmt.Errorf("subscribe prompt: %w", err)
	}

	cb := func(msg *iterm2.ServerOriginatedMessage) bool {
		n := msg.GetNotification().GetPromptNotification()
		if n == nil {
			return false
		}
		ev := PromptEvent{UniquePromptID: n.GetUniquePromptId()}
		switch e := n.Event.(type) {
		case *iterm2.PromptNotification_Prompt:
			ev.Mode = iterm2.PromptMonitorMode_PROMPT
			if pp := e.Prompt; pp != nil && pp.GetPrompt() != nil {
				ev.Prompt = NewPrompt(pp.GetPrompt())
			}
		case *iterm2.PromptNotification_CommandStart:
			ev.Mode = iterm2.PromptMonitorMode_COMMAND_START
			if cs := e.CommandStart; cs != nil {
				ev.Command = cs.GetCommand()
			}
		case *iterm2.PromptNotification_CommandEnd:
			ev.Mode = iterm2.PromptMonitorMode_COMMAND_END
			if ce := e.CommandEnd; ce != nil {
				ev.Status = ce.GetStatus()
			}
		default:
			return false
		}
		select {
		case pm.ch <- ev:
		case <-pm.done:
		}
		return false
	}
	pm.token = conn.storeHandler(key, cb)
	pm.token.nt = iterm2.NotificationType_NOTIFY_ON_PROMPT
	pm.token.sid = sessionID

	// Bridge: dispatchLoop calls handlers, not conn.Dispatch, so we need
	// a dispatchHandler to route notifications to our notifyMap handler.
	pm.dispatchHandler = func(msg *iterm2.ServerOriginatedMessage) bool {
		conn.Dispatch(msg)
		return false
	}
	conn.RegisterHandler(pm.dispatchHandler)

	// Auto-close on disconnect.
	conn.OnDisconnect(func() { pm.Close() })

	return pm, nil
}

// Chan returns a receive-only channel of PromptEvents.
func (pm *PromptMonitor) Chan() <-chan PromptEvent { return pm.ch }

// Close stops the monitor and unsubscribes.  Safe to call multiple times.
func (pm *PromptMonitor) Close() {
	pm.once.Do(func() {
		close(pm.done)
		pm.conn.Unsubscribe(pm.token)
		if pm.dispatchHandler != nil {
			pm.conn.UnregisterHandler(pm.dispatchHandler)
		}
		close(pm.ch)
	})
}
