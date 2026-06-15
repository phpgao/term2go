package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// PreferenceKey identifies a preference setting in iTerm2.
type PreferenceKey string

const (
	PrefOpenBookmark                  PreferenceKey = "OpenBookmark"
	PrefOpenArrangementAtStartup      PreferenceKey = "OpenArrangementAtStartup"
	PrefOpenNoWindowsAtStartup        PreferenceKey = "OpenNoWindowsAtStartup"
	PrefQuitWhenAllWindowsClosed      PreferenceKey = "QuitWhenAllWindowsClosed"
	PrefConfirmClosingMultipleTabs    PreferenceKey = "OnlyWhenMoreTabs"
	PrefPromptOnQuit                  PreferenceKey = "PromptOnQuit"
	PrefIRMemory                      PreferenceKey = "IRMemory"
	PrefSavePasteHistory              PreferenceKey = "SavePasteHistory"
	PrefEnableRendezvous              PreferenceKey = "EnableRendezvous"
	PrefSUEnableAutomaticChecks       PreferenceKey = "SUEnableAutomaticChecks"
	PrefCheckTestRelease              PreferenceKey = "CheckTestRelease"
	PrefLoadPrefsFromCustomFolder     PreferenceKey = "LoadPrefsFromCustomFolder"
	PrefPrefsCustomFolder             PreferenceKey = "PrefsCustomFolder"
	PrefCopySelection                 PreferenceKey = "CopySelection"
	PrefCopyLastNewline               PreferenceKey = "CopyLastNewline"
	PrefAllowClipboardAccess          PreferenceKey = "AllowClipboardAccess"
	PrefWordCharacters                PreferenceKey = "WordCharacters"
	PrefSmartPlacement                PreferenceKey = "SmartPlacement"
	PrefAdjustWindowForFontSize       PreferenceKey = "AdjustWindowForFontSizeChange"
	PrefMaxVertically                 PreferenceKey = "MaxVertically"
	PrefUseLionStyleFullscreen        PreferenceKey = "UseLionStyleFullscreen"
	PrefOpenTmuxWindowsIn             PreferenceKey = "OpenTmuxWindowsIn"
	PrefTmuxDashboardLimit            PreferenceKey = "TmuxDashboardLimit"
	PrefAutoHideTmuxClientSession     PreferenceKey = "AutoHideTmuxClientSession"
	PrefTmuxUsesDedicatedProfile      PreferenceKey = "TmuxUsesDedicatedProfile"
	PrefUseMetal                      PreferenceKey = "UseMetal"
	PrefDisableMetalWhenUnplugged     PreferenceKey = "disableMetalWhenUnplugged"
	PrefPreferIntegratedGPU           PreferenceKey = "preferIntegratedGPU"
	PrefMaximizeThroughput            PreferenceKey = "metalMaximizeThroughput"
	PrefTheme                         PreferenceKey = "TabStyleWithAutomaticOption"
	PrefTabViewType                   PreferenceKey = "TabViewType"
	PrefHideTab                       PreferenceKey = "HideTab"
	PrefHideTabNumber                 PreferenceKey = "HideTabNumber"
	PrefHideTabCloseButton            PreferenceKey = "HideTabCloseButton"
	PrefHideActivityIndicator         PreferenceKey = "HideActivityIndicator"
	PrefShowNewOutputIndicator        PreferenceKey = "ShowNewOutputIndicator"
	PrefShowPaneTitles                PreferenceKey = "ShowPaneTitles"
	PrefStretchTabsToFillBar          PreferenceKey = "StretchTabsToFillBar"
	PrefHideMenuBarInFullscreen       PreferenceKey = "HideMenuBarInFullscreen"
	PrefHideFromDockAndAppSwitcher    PreferenceKey = "HideFromDockAndAppSwitcher"
	PrefFlashTabBarInFullscreen       PreferenceKey = "FlashTabBarInFullscreen"
	PrefWindowNumber                  PreferenceKey = "WindowNumber"
	PrefDimOnlyText                   PreferenceKey = "DimOnlyText"
	PrefSplitPaneDimmingAmount        PreferenceKey = "SplitPaneDimmingAmount"
	PrefDimInactiveSplitPanes         PreferenceKey = "DimInactiveSplitPanes"
	PrefUseBorder                     PreferenceKey = "UseBorder"
	PrefHideScrollbar                 PreferenceKey = "HideScrollbar"
	PrefDisableFullscreenTransparency PreferenceKey = "DisableFullscreenTransparency"
	PrefEnableDivisionView            PreferenceKey = "EnableDivisionView"
	PrefEnableProxyIcon               PreferenceKey = "EnableProxyIcon"
	PrefDimBackgroundWindows          PreferenceKey = "DimBackgroundWindows"
	PrefControlRemapping              PreferenceKey = "Control"
	PrefLeftOptionRemapping           PreferenceKey = "LeftOption"
	PrefRightOptionRemapping          PreferenceKey = "RightOption"
	PrefLeftCommandRemapping          PreferenceKey = "LeftCommand"
	PrefRightCommandRemapping         PreferenceKey = "RightCommand"
	PrefSwitchPaneModifier            PreferenceKey = "SwitchPaneModifier"
	PrefSwitchTabModifier             PreferenceKey = "SwitchTabModifier"
	PrefSwitchWindowModifier          PreferenceKey = "SwitchWindowModifier"
	PrefCommandSelection              PreferenceKey = "CommandSelection"
	PrefPassOnControlClick            PreferenceKey = "PassOnControlClick"
	PrefOptionClickMovesCursor        PreferenceKey = "OptionClickMovesCursor"
	PrefThreeFingerEmulates           PreferenceKey = "ThreeFingerEmulates"
	PrefFocusFollowsMouse             PreferenceKey = "FocusFollowsMouse"
	PrefTripleClickSelectsWrapped     PreferenceKey = "TripleClickSelectsFullWrappedLines"
	PrefDoubleClickSmartSelection     PreferenceKey = "DoubleClickPerformsSmartSelection"
	PrefITermVersion                  PreferenceKey = "iTerm Version"
	PrefSizeChangesAffectProfile      PreferenceKey = "Size Changes Affect Profile"
	PrefHTMLTabTitles                 PreferenceKey = "HTMLTabTitles"
	PrefDisableTransparencyForKey     PreferenceKey = "DisableTransparencyForKeyWindow"
)

// GetPreference reads a single preference value.
// Returns the raw JSON value as a string (e.g., "true", "42", or `"hello"`).
func GetPreference(ctx context.Context, caller Caller, key PreferenceKey) (string, error) {
	req := &iterm2.PreferencesRequest{
		Requests: []*iterm2.PreferencesRequest_Request{
			{
				Request: &iterm2.PreferencesRequest_Request_GetPreferenceRequest{
					GetPreferenceRequest: &iterm2.PreferencesRequest_Request_GetPreference{
						Key: proto.String(string(key)),
					},
				},
			},
		},
	}
	resp, err := PreferencesRequest(ctx, caller, req)
	if err != nil {
		return "", fmt.Errorf("get preference %s: %w", key, err)
	}
	results := resp.GetResults()
	if len(results) == 0 {
		return "", nil
	}
	r := results[0]
	if gp := r.GetGetPreferenceResult(); gp != nil {
		return gp.GetJsonValue(), nil
	}
	return "", nil
}

// SetPreference writes a single preference value.
// The jsonValue should be a JSON-encoded value matching the preference type
// (e.g., "true" for booleans, "42" for numbers, `"hello"` for strings).
func SetPreference(ctx context.Context, caller Caller, key PreferenceKey, jsonValue string) error {
	req := &iterm2.PreferencesRequest{
		Requests: []*iterm2.PreferencesRequest_Request{
			{
				Request: &iterm2.PreferencesRequest_Request_SetPreferenceRequest{
					SetPreferenceRequest: &iterm2.PreferencesRequest_Request_SetPreference{
						Key:       proto.String(string(key)),
						JsonValue: proto.String(jsonValue),
					},
				},
			},
		},
	}
	_, err := PreferencesRequest(ctx, caller, req)
	if err != nil {
		return fmt.Errorf("set preference %s: %w", key, err)
	}
	return nil
}

// SetPreferenceJSON is like SetPreference but accepts a Go value and JSON-encodes it.
func SetPreferenceJSON(ctx context.Context, caller Caller, key PreferenceKey, value interface{}) error {
	v, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json encode preference %s: %w", key, err)
	}
	return SetPreference(ctx, caller, key, string(v))
}
