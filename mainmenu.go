package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Menu items use the identifier strings from Python's MenuItemIdentifier.
// The value is the second field (identifier), not the display title.
// Hierarchical names use dots as separators (e.g., "Edit.Find.Find...").

const (
	// ---- iTerm2 ----
	MenuItemAboutITerm2             = "About iTerm2"
	MenuItemShowTipOfTheDay         = "Show Tip of the Day"
	MenuItemCheckForUpdates         = "Check For Updates…"
	MenuItemToggleDebugLogging      = "Toggle Debug Logging"
	MenuItemCopyPerformanceStats    = "Copy Performance Stats"
	MenuItemCaptureGPUFrame         = "Capture Metal Frame"
	MenuItemPreferences             = "Preferences..."
	MenuItemHideITerm2              = "Hide iTerm2"
	MenuItemHideOthers              = "Hide Others"
	MenuItemShowAll                 = "Show All"
	MenuItemSecureKeyboard          = "Secure Keyboard Entry"
	MenuItemMakeITerm2DefaultTerm   = "Make iTerm2 Default Term"
	MenuItemMakeTerminalDefaultTerm = "Make Terminal Default Term"
	MenuItemInstallShellIntegration = "Install Shell Integration"
	MenuItemQuitITerm2              = "Quit iTerm2"

	// ---- Shell ----
	MenuItemNewWindow                           = "New Window"
	MenuItemNewWindowWithCurrentProfile         = "New Window with Current Profile"
	MenuItemNewTab                              = "New Tab"
	MenuItemNewTabWithCurrentProfile            = "New Tab with Current Profile"
	MenuItemDuplicateTab                        = "Duplicate Tab"
	MenuItemSplitHorizontallyWithCurrentProfile = "Split Horizontally with Current Profile"
	MenuItemSplitVerticallyWithCurrentProfile   = "Split Vertically with Current Profile"
	MenuItemSplitHorizontally                   = "Split Horizontally…"
	MenuItemSplitVertically                     = "Split Vertically…"
	MenuItemSaveContents                        = "Log.SaveContents"
	MenuItemSaveSelectedText                    = "Save Selected Text…"
	MenuItemClose                               = "Close"
	MenuItemCloseTerminalWindow                 = "Close Terminal Window"
	MenuItemCloseAllPanesInTab                  = "Close All Panes in Tab"
	MenuItemUndoClose                           = "Undo Close"

	// Shell > BroadcastInput
	MenuItemSendInputToCurrentSessionOnly        = "Broadcast Input.Send Input to Current Session Only"
	MenuItemBroadcastInputToAllPanesInAllTabs    = "Broadcast Input.Broadcast Input to All Panes in All Tabs"
	MenuItemBroadcastInputToAllPanesInCurrentTab = "Broadcast Input.Broadcast Input to All Panes in Current Tab"
	MenuItemToggleBroadcastInputToCurrentSession = "Broadcast Input.Toggle Broadcast Input to Current Session"
	MenuItemShowBackgroundPatternIndicator       = "Broadcast Input.Show Background Pattern Indicator"

	// Shell > tmux
	MenuItemTmuxDetach      = "tmux.Detach"
	MenuItemTmuxForceDetach = "tmux.Force Detach"
	MenuItemTmuxNewWindow   = "tmux.New Tmux Window"
	MenuItemTmuxNewTab      = "tmux.New Tmux Tab"
	MenuItemTmuxPausePane   = "trmux.Pause Pane"
	MenuItemTmuxDashboard   = "tmux.Dashboard"

	// Shell > ssh
	MenuItemSSHDisconnect         = "ssh.Disconnect"
	MenuItemSSHRemoveFileProvider = "ssh.Remove File Provider"
	MenuItemSSHAddFileProvider    = "ssh.Add File Provider"

	// Shell > Print
	MenuItemPageSetup      = "Page Setup..."
	MenuItemPrintScreen    = "Print.Screen"
	MenuItemPrintSelection = "Print.Selection"
	MenuItemPrintBuffer    = "Print.Buffer"

	// ---- Edit ----
	MenuItemUndo                     = "Undo"
	MenuItemRedo                     = "Redo"
	MenuItemCut                      = "Cut"
	MenuItemCopy                     = "Copy"
	MenuItemCopyWithStyles           = "Copy with Styles"
	MenuItemCopyWithControlSequences = "Copy with Control Sequences"
	MenuItemCopyMode                 = "Copy Mode"
	MenuItemPaste                    = "Paste"

	// Edit > PasteSpecial
	MenuItemAdvancedPaste                     = "Paste Special.Advanced Paste…"
	MenuItemPasteSelection                    = "Paste Special.Paste Selection"
	MenuItemPasteFileBase64Encoded            = "Paste Special.Paste File Base64-Encoded"
	MenuItemPasteSlowly                       = "Paste Special.Paste Slowly"
	MenuItemPasteFaster                       = "Paste Special.Paste Faster"
	MenuItemPasteSlowlyFaster                 = "Paste Special.Paste Slowly Faster"
	MenuItemPasteSlower                       = "Paste Special.Paste Slower"
	MenuItemPasteSlowlySlower                 = "Paste Special.Paste Slowly Slower"
	MenuItemWarnBeforeMultilinePaste          = "Paste Special.Warn Before Multi-Line Paste"
	MenuItemPromptConvertTabsToSpacesOnPaste  = "Paste Special.Prompt to Convert Tabs to Spaces when Pasting"
	MenuItemLimitMultilinePasteWarningToShell = "Paste Special.Limit Multi-Line Paste Warning to Shell Prompt"
	MenuItemWarnBeforePastingOneLine          = "Paste Special.Warn Before Pasting One Line Ending in a Newline at Shell Prompt"

	MenuItemRenderSelection                 = "Render Selection Natively"
	MenuItemOpenSelection                   = "Open Selection"
	MenuItemJumpToSelection                 = "Find.Jump to Selection"
	MenuItemSelectAll                       = "Select All"
	MenuItemSelectionRespectsSoftBoundaries = "Selection Respects Soft Boundaries"
	MenuItemSelectOutputOfLastCommand       = "Select Output of Last Command"
	MenuItemSelectCurrentCommand            = "Select Current Command"

	// Edit > Find
	MenuItemFindFind            = "Find.Find..."
	MenuItemFindNext            = "Find.Find Next"
	MenuItemFindPrevious        = "Find.Find Previous"
	MenuItemUseSelectionForFind = "Find.Use Selection for Find"
	MenuItemFindGlobally        = "Find.Find Globally..."
	MenuItemSelectMatches       = "Find.ConvertMatchesToSelections"
	MenuItemFindURLs            = "Find.Find URLs"
	MenuItemFindPickResult      = "Find.Pick Result To Open"
	MenuItemFilter              = "Find.Filter"

	// Edit > MarksAndAnnotations
	MenuItemSetMark               = "Marks and Annotations.Set Mark"
	MenuItemJumpToMark            = "Marks and Annotations.Jump to Mark"
	MenuItemNextMark              = "Marks and Annotations.Next Mark"
	MenuItemPreviousMark          = "Marks and Annotations.Previous Mark"
	MenuItemAddAnnotationAtCursor = "Marks and Annotations.Add Annotation at Cursor"
	MenuItemNextAnnotation        = "Marks and Annotations.Next  Annotation"
	MenuItemPreviousAnnotation    = "Marks and Annotations.Previous  Annotation"

	// Edit > MarksAndAnnotations > Alerts
	MenuItemAlertOnNextMark   = "Marks and Annotations.Alerts.Alert on Next Mark"
	MenuItemShowModalAlertBox = "Marks and Annotations.Alerts.Show Modal Alert Box"
	MenuItemPostNotification  = "Marks and Annotations.Alerts.Post Notification"

	MenuItemClearBuffer             = "Clear Buffer"
	MenuItemClearScrollbackBuffer   = "Clear Scrollback Buffer"
	MenuItemClearToStartOfSelection = "Clear to Start of Selection"
	MenuItemClearToLastMark         = "Clear to Last Mark"

	// ---- View ----
	MenuItemShowTabsInFullscreen               = "Show Tabs in Fullscreen"
	MenuItemToggleFullScreen                   = "Toggle Full Screen"
	MenuItemUseTransparency                    = "Use Transparency"
	MenuItemDisableTransparencyForActiveWindow = "Disable Transparency for Active Window"
	MenuItemZoomInOnSelection                  = "Zoom In on Selection"
	MenuItemZoomOut                            = "Zoom Out"
	MenuItemFindCursor                         = "Find Cursor"
	MenuItemShowCursorGuide                    = "Show Cursor Guide"
	MenuItemShowTimestamps                     = "Show Timestamps"
	MenuItemShowAnnotations                    = "Show Annotations"
	MenuItemShowComposer                       = "Composer"
	MenuItemAutoCommandCompletion              = "Auto Command Completion"
	MenuItemOpenQuickly                        = "Open Quickly"
	MenuItemMaximizeActivePane                 = "Maximize Active Pane"
	MenuItemMakeTextBigger                     = "Make Text Bigger"
	MenuItemMakeTextNormalSize                 = "Make Text Normal Size"
	MenuItemRestoreTextAndSessionSize          = "Restore Text and Session Size"
	MenuItemMakeTextSmaller                    = "Make Text Smaller"
	MenuItemSizeChangesUpdateProfile           = "Size Changes Update Profile"
	MenuItemStartInstantReplay                 = "Start Instant Replay"

	// ---- Session ----
	MenuItemEditSession           = "Edit Session…"
	MenuItemRunCoprocess          = "Run Coprocess…"
	MenuItemStopCoprocess         = "Stop Coprocess"
	MenuItemRestartSession        = "Restart Session"
	MenuItemOpenAutocomplete      = "Open Autocomplete…"
	MenuItemOpenCommandHistory    = "Open Command History…"
	MenuItemOpenRecentDirectories = "Open Recent Directories…"
	MenuItemOpenPasteHistory      = "Open Paste History…"

	// Session > Triggers
	MenuItemAddTrigger                  = "Add Trigger"
	MenuItemEditTriggers                = "Edit Triggers"
	MenuItemEnableTriggersInInteractive = "Enable Triggers in Interactive Apps"
	MenuItemTriggersEnableAll           = "Triggers.Enable All"
	MenuItemTriggersDisableAll          = "Triggers.Disable All"

	MenuItemReset             = "Reset"
	MenuItemResetCharacterSet = "Reset Character Set"

	// Session > Log
	MenuItemLogToggle          = "Log.Toggle"
	MenuItemLogImportRecording = "Log.ImportRecording"
	MenuItemLogExportRecording = "Log.ExportRecording"

	// Session > TerminalState
	MenuItemAlternateScreen          = "Alternate Screen"
	MenuItemFocusReporting           = "Focus Reporting"
	MenuItemMouseReporting           = "Mouse Reporting"
	MenuItemPasteBracketing          = "Paste Bracketing"
	MenuItemApplicationCursor        = "Application Cursor"
	MenuItemApplicationKeypad        = "Application Keypad"
	MenuItemStandardKeyReportingMode = "Terminal State.Standard Key Reporting"
	MenuItemModifyOtherKeysMode1     = "Terminal State.Report Modifiers like xterm 1"
	MenuItemModifyOtherKeysMode2     = "Terminal State.Report Modifiers like xterm 2"
	MenuItemCSIuMode                 = "Terminal State.Report Modifiers with CSI u"
	MenuItemRawKeyReportingMode      = "Terminal State.Raw Key Reporting"
	MenuItemResetTerminalState       = "Reset Terminal State"

	MenuItemBurySession = "Bury Session"

	// ---- Scripts > Manage ----
	MenuItemNewPythonScript       = "New Python Script"
	MenuItemOpenPythonREPL        = "Open Interactive Window"
	MenuItemManageDependencies    = "Manage Dependencies"
	MenuItemInstallPythonRuntime  = "Install Python Runtime"
	MenuItemRevealScriptsInFinder = "Reveal in Finder"
	MenuItemScriptsImport         = "Import Script"
	MenuItemScriptsExport         = "Export Script"
	MenuItemScriptsConsole        = "Script Console"

	// ---- Profiles ----
	MenuItemOpenProfiles            = "Open Profiles…"
	MenuItemPressOptionForNewWindow = "Press Option for New Window"
	MenuItemOpenInNewWindow         = "Open In New Window"

	// ---- Toolbelt ----
	MenuItemShowToolbelt    = "Show Toolbelt"
	MenuItemSetDefaultWidth = "Set Default Width"

	// ---- Window ----
	MenuItemMinimize        = "Minimize"
	MenuItemZoom            = "Zoom"
	MenuItemEditTabTitle    = "Edit Tab Title"
	MenuItemEditWindowTitle = "Edit Window Title"

	// Window > WindowStyle
	MenuItemWindowStyleNormal          = "Window Style.Normal"
	MenuItemWindowStyleFullScreen      = "Window Style.Full Screen"
	MenuItemWindowStyleMaximized       = "Window Style.Maximized"
	MenuItemWindowStyleNoTitleBar      = "Window Style.No Title Bar"
	MenuItemWindowStyleFullWidthBottom = "Window Style.FullWidth Bottom of Screen"
	MenuItemWindowStyleFullWidthTop    = "Window Style.FullWidth Top of Screen"
	MenuItemWindowStyleFullHeightLeft  = "Window Style..FullHeight Left of Screen"
	MenuItemWindowStyleFullHeightRight = "Window Style.FullHeight Right of Screen"
	MenuItemWindowStyleBottom          = "Window Style.Bottom of Screen"
	MenuItemWindowStyleTop             = "Window Style.Top of Screen"
	MenuItemWindowStyleLeft            = "Window Style.Left of Screen"
	MenuItemWindowStyleRight           = "Window Style.Right of Screen"

	MenuItemMergeAllWindows                = "Merge All Windows"
	MenuItemArrangeWindowsHorizontally     = "Arrange Windows Horizontally"
	MenuItemArrangeSplitPanesEvenly        = "Arrange Split Panes Evenly"
	MenuItemMoveSessionToWindow            = "Move Session to Window"
	MenuItemSaveWindowArrangement          = "Save Window Arrangement"
	MenuItemSaveCurrentWindowAsArrangement = "Save Current Window as Arrangement"

	// Window > SelectSplitPane
	MenuItemSelectPaneAbove    = "Select Split Pane.Select Pane Above"
	MenuItemSelectPaneBelow    = "Select Split Pane.Select Pane Below"
	MenuItemSelectPaneLeft     = "Select Split Pane.Select Pane Left"
	MenuItemSelectPaneRight    = "Select Split Pane.Select Pane Right"
	MenuItemSelectNextPane     = "Select Split Pane.Next Pane"
	MenuItemSelectPreviousPane = "Select Split Pane.Previous Pane"

	// Window > ResizeSplitPane
	MenuItemMoveDividerUp    = "Resize Split Pane.Move Divider Up"
	MenuItemMoveDividerDown  = "Resize Split Pane.Move Divider Down"
	MenuItemMoveDividerLeft  = "Resize Split Pane.Move Divider Left"
	MenuItemMoveDividerRight = "Resize Split Pane.Move Divider Right"

	// Window > ResizeWindow
	MenuItemResizeDecreaseHeight = "Resize Window.Decrease Height"
	MenuItemResizeIncreaseHeight = "Resize Window.Increase Height"
	MenuItemResizeDecreaseWidth  = "Resize Window.Decrease Width"
	MenuItemResizeIncreaseWidth  = "Resize Window.Increase Width"

	MenuItemSelectNextTab     = "Select Next Tab"
	MenuItemSelectPreviousTab = "Select Previous Tab"
	MenuItemMoveTabLeft       = "Move Tab Left"
	MenuItemMoveTabRight      = "Move Tab Right"
	MenuItemPasswordManager   = "Password Manager"
	MenuItemPinHotkeyWindow   = "Pin Hotkey Window"
	MenuItemBringAllToFront   = "Bring All To Front"

	// ---- Help ----
	MenuItemITerm2Help              = "iTerm2 Help"
	MenuItemCopyModeShortcuts       = "Copy Mode Shortcuts"
	MenuItemOpenSourceLicenses      = "Open Source Licenses"
	MenuItemGPURendererAvailability = "GPU Renderer Availability"
)

// SelectMenuItem selects a menu item by its identifier string.
func SelectMenuItem(ctx context.Context, caller Caller, identifier string) error {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_MenuItemRequest{
		MenuItemRequest: &iterm2.MenuItemRequest{
			Identifier: proto.String(identifier),
			QueryOnly:  proto.Bool(false),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return err
	}
	if err = checkError(resp); err != nil {
		return err
	}
	menuResp := resp.GetMenuItemResponse()
	if menuResp == nil {
		return fmt.Errorf("menu item response is nil")
	}
	switch menuResp.GetStatus() {
	case iterm2.MenuItemResponse_OK:
		return nil
	case iterm2.MenuItemResponse_DISABLED:
		return fmt.Errorf("menu item %q is disabled", identifier)
	default:
		return fmt.Errorf("menu item %q: unknown status %v", identifier, menuResp.GetStatus())
	}
}

// MenuItemState describes the current state of a menu item.
type MenuItemState struct {
	Checked bool
	Enabled bool
}

// GetMenuItemState queries the state of a menu item by its identifier string.
func GetMenuItemState(ctx context.Context, caller Caller, identifier string) (*MenuItemState, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_MenuItemRequest{
		MenuItemRequest: &iterm2.MenuItemRequest{
			Identifier: proto.String(identifier),
			QueryOnly:  proto.Bool(true),
		},
	}
	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	menuResp := resp.GetMenuItemResponse()
	if menuResp == nil {
		return nil, fmt.Errorf("menu item response is nil")
	}
	if menuResp.GetStatus() != iterm2.MenuItemResponse_OK {
		return nil, fmt.Errorf("menu item %q: status %v", identifier, menuResp.GetStatus())
	}
	return &MenuItemState{
		Checked: menuResp.GetChecked(),
		Enabled: menuResp.GetEnabled(),
	}, nil
}
