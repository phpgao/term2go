package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestWithTabIndex tests WithTabIndex option for setting tab index in CreateTab request.
func TestWithTabIndex(t *testing.T) {
	req := &iterm2.CreateTabRequest{}
	WithTabIndex(5)(req)
	assert.Equal(t, uint32(5), req.GetTabIndex())
}

// TestWithCustomProfileProperties tests WithCustomProfileProperties option for setting custom profile properties.
func TestWithCustomProfileProperties(t *testing.T) {
	props := []*iterm2.ProfileProperty{{JsonValue: strPtr("v")}}
	req := &iterm2.CreateTabRequest{}
	WithCustomProfileProperties(props)(req)
	assert.Len(t, req.GetCustomProfileProperties(), 1)
}

// TestWithSplitPaneCustomProfileProperties tests WithSplitPaneCustomProfileProperties option for setting custom profile properties in SplitPane request.
func TestWithSplitPaneCustomProfileProperties(t *testing.T) {
	props := []*iterm2.ProfileProperty{{JsonValue: strPtr("v")}}
	req := &iterm2.SplitPaneRequest{}
	WithSplitPaneCustomProfileProperties(props)(req)
	assert.Len(t, req.GetCustomProfileProperties(), 1)
}

// TestWithIncludeStyles tests WithIncludeStyles option for including style information in GetBuffer request.
func TestWithIncludeStyles(t *testing.T) {
	req := &iterm2.GetBufferRequest{}
	WithIncludeStyles()(req)
	assert.True(t, req.GetIncludeStyles())
}

// TestWithUniquePromptID tests WithUniquePromptID option for setting unique prompt ID in GetPrompt request.
func TestWithUniquePromptID(t *testing.T) {
	req := &iterm2.GetPromptRequest{}
	WithUniquePromptID("prompt-1")(req)
	assert.Equal(t, "prompt-1", req.GetUniquePromptId())
}

// TestWithFirstUniqueID tests WithFirstUniqueID option for setting first unique ID in ListPrompts request.
func TestWithFirstUniqueID(t *testing.T) {
	req := &iterm2.ListPromptsRequest{}
	WithFirstUniqueID("start")(req)
	assert.Equal(t, "start", req.GetFirstUniqueId())
}

// TestWithLastUniqueID tests WithLastUniqueID option for setting last unique ID in ListPrompts request.
func TestWithLastUniqueID(t *testing.T) {
	req := &iterm2.ListPromptsRequest{}
	WithLastUniqueID("end")(req)
	assert.Equal(t, "end", req.GetLastUniqueId())
}

// TestWithSelectSession tests WithSelectSession option for selecting session in Activate request.
func TestWithSelectSession(t *testing.T) {
	req := &iterm2.ActivateRequest{}
	WithSelectSession()(req)
	assert.True(t, req.GetSelectSession())
}

// TestWithActivateApp tests WithActivateApp option for setting activate app parameters in Activate request.
func TestWithActivateApp(t *testing.T) {
	req := &iterm2.ActivateRequest{}
	WithActivateApp(true, false)(req)
	app := req.GetActivateApp()
	assert.True(t, app.GetRaiseAllWindows())
	assert.False(t, app.GetIgnoringOtherApps())
}
