package term2go

import (
	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// CreateTabOption is an option for CreateTab.
type CreateTabOption func(*iterm2.CreateTabRequest)

// WithTabIndex sets the desired index of the new tab.
// Only valid if the tab is being created in an existing window (windowID is set).
func WithTabIndex(idx uint32) CreateTabOption {
	return func(req *iterm2.CreateTabRequest) {
		req.TabIndex = proto.Uint32(idx)
	}
}

// WithCustomProfileProperties modifies the profile to customize its behavior just for this session.
func WithCustomProfileProperties(props []*iterm2.ProfileProperty) CreateTabOption {
	return func(req *iterm2.CreateTabRequest) {
		req.CustomProfileProperties = props
	}
}

// SplitPaneOption is an option for SplitPane.
type SplitPaneOption func(*iterm2.SplitPaneRequest)

// WithSplitPaneCustomProfileProperties modifies the profile for the split pane session.
func WithSplitPaneCustomProfileProperties(props []*iterm2.ProfileProperty) SplitPaneOption {
	return func(req *iterm2.SplitPaneRequest) {
		req.CustomProfileProperties = props
	}
}

// GetBufferOption is an option for GetBuffer.
type GetBufferOption func(*iterm2.GetBufferRequest)

// WithIncludeStyles populates the style field of LineContents in the response.
func WithIncludeStyles() GetBufferOption {
	return func(req *iterm2.GetBufferRequest) {
		req.IncludeStyles = proto.Bool(true)
	}
}

// GetPromptOption is an option for GetPrompt.
type GetPromptOption func(*iterm2.GetPromptRequest)

// WithUniquePromptID returns the prompt with the given ID instead of the last one.
func WithUniquePromptID(id string) GetPromptOption {
	return func(req *iterm2.GetPromptRequest) {
		req.UniquePromptId = proto.String(id)
	}
}

// ListPromptsOption is an option for ListPrompts.
type ListPromptsOption func(*iterm2.ListPromptsRequest)

// WithFirstUniqueID starts listing prompts from the given ID (exclusive).
func WithFirstUniqueID(id string) ListPromptsOption {
	return func(req *iterm2.ListPromptsRequest) {
		req.FirstUniqueId = proto.String(id)
	}
}

// WithLastUniqueID ends listing prompts at the given ID (inclusive).
func WithLastUniqueID(id string) ListPromptsOption {
	return func(req *iterm2.ListPromptsRequest) {
		req.LastUniqueId = proto.String(id)
	}
}

// ActivateOption is an option for Activate.
type ActivateOption func(*iterm2.ActivateRequest)

// WithSelectSession selects the session in addition to the tab.
func WithSelectSession() ActivateOption {
	return func(req *iterm2.ActivateRequest) {
		req.SelectSession = proto.Bool(true)
	}
}

// WithActivateApp also activates the app.
func WithActivateApp(raiseAllWindows, ignoringOtherApps bool) ActivateOption {
	return func(req *iterm2.ActivateRequest) {
		req.ActivateApp = &iterm2.ActivateRequest_App{
			RaiseAllWindows:   proto.Bool(raiseAllWindows),
			IgnoringOtherApps: proto.Bool(ignoringOtherApps),
		}
	}
}

// SendTextOption is an option for SendText.
type SendTextOption func(*iterm2.SendTextRequest)

// WithSendTextSuppressBroadcast prevents broadcast when broadcasting is on.
func WithSendTextSuppressBroadcast(suppress bool) SendTextOption {
	return func(req *iterm2.SendTextRequest) {
		req.SuppressBroadcast = proto.Bool(suppress)
	}
}

// CloseOption is an option for Close.
type CloseOption func(*iterm2.CloseRequest)

// WithCloseForce forces the close without confirmation.
func WithCloseForce(force bool) CloseOption {
	return func(req *iterm2.CloseRequest) {
		req.Force = proto.Bool(force)
	}
}

// WithCloseTabs closes tabs instead of sessions.
func WithCloseTabs(tabIDs []string) CloseOption {
	return func(req *iterm2.CloseRequest) {
		req.Target = &iterm2.CloseRequest_Tabs{
			Tabs: &iterm2.CloseRequest_CloseTabs{
				TabIds: tabIDs,
			},
		}
	}
}

// WithCloseWindows closes windows instead of sessions.
func WithCloseWindows(windowIDs []string) CloseOption {
	return func(req *iterm2.CloseRequest) {
		req.Target = &iterm2.CloseRequest_Windows{
			Windows: &iterm2.CloseRequest_CloseWindows{
				WindowIds: windowIDs,
			},
		}
	}
}

// RestartSessionOption is an option for RestartSession.
type RestartSessionOption func(*iterm2.RestartSessionRequest)

// WithRestartOnlyIfExited only restarts if the session has exited.
func WithRestartOnlyIfExited(onlyIfExited bool) RestartSessionOption {
	return func(req *iterm2.RestartSessionRequest) {
		req.OnlyIfExited = proto.Bool(onlyIfExited)
	}
}
