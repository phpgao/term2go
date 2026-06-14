package term2go

// BindingAction represents an action triggered by a key binding in iTerm2.
// Values match the Python iterm2.BindingAction enum.
type BindingAction int

const (
	ActionNextSession                  BindingAction = 0
	ActionNextWindow                   BindingAction = 1
	ActionPreviousSession              BindingAction = 2
	ActionPreviousWindow               BindingAction = 3
	ActionScrollEnd                    BindingAction = 4
	ActionScrollHome                   BindingAction = 5
	ActionScrollLineDown               BindingAction = 6
	ActionScrollLineUp                 BindingAction = 7
	ActionScrollPageDown               BindingAction = 8
	ActionScrollPageUp                 BindingAction = 9
	ActionEscapeSequence               BindingAction = 10
	ActionHexCode                      BindingAction = 11
	ActionText                         BindingAction = 12
	ActionIgnore                       BindingAction = 13
	ActionIRBackward                   BindingAction = 15
	ActionSendCHBackspace              BindingAction = 16
	ActionSendCQMBackspace             BindingAction = 17
	ActionSelectPaneLeft               BindingAction = 18
	ActionSelectPaneRight              BindingAction = 19
	ActionSelectPaneAbove              BindingAction = 20
	ActionSelectPaneBelow              BindingAction = 21
	ActionDoNotRemapModifiers          BindingAction = 22
	ActionToggleFullscreen             BindingAction = 23
	ActionRemapLocally                 BindingAction = 24
	ActionSelectMenuItem               BindingAction = 25
	ActionNewWindowWithProfile         BindingAction = 26
	ActionNewTabWithProfile            BindingAction = 27
	ActionSplitHorizontallyWithProfile BindingAction = 28
	ActionSplitVerticallyWithProfile   BindingAction = 29
	ActionNextPane                     BindingAction = 30
	ActionPreviousPane                 BindingAction = 31
	ActionNextMRUTab                   BindingAction = 32
	ActionMoveTabLeft                  BindingAction = 33
	ActionMoveTabRight                 BindingAction = 34
	ActionRunCoprocess                 BindingAction = 35
	ActionFindRegex                    BindingAction = 36
	ActionSetProfile                   BindingAction = 37
	ActionVimText                      BindingAction = 38
	ActionPreviousMRUTab               BindingAction = 39
	ActionLoadColorPreset              BindingAction = 40
	ActionPasteSpecial                 BindingAction = 41
	ActionPasteSpecialFromSelection    BindingAction = 42
	ActionToggleHotkeyWindowPinning    BindingAction = 43
	ActionUndo                         BindingAction = 44
	ActionMoveEndOfSelectionLeft       BindingAction = 45
	ActionMoveEndOfSelectionRight      BindingAction = 46
	ActionMoveStartOfSelectionLeft     BindingAction = 47
	ActionMoveStartOfSelectionRight    BindingAction = 48
	ActionDecreaseHeight               BindingAction = 49
	ActionIncreaseHeight               BindingAction = 50
	ActionDecreaseWidth                BindingAction = 51
	ActionIncreaseWidth                BindingAction = 52
	ActionSwapPaneLeft                 BindingAction = 53
	ActionSwapPaneRight                BindingAction = 54
	ActionSwapPaneAbove                BindingAction = 55
	ActionSwapPaneBelow                BindingAction = 56
	ActionFindAgainDown                BindingAction = 57
	ActionFindAgainUp                  BindingAction = 58
	ActionToggleMouseReporting         BindingAction = 59
	ActionInvokeScriptFunction         BindingAction = 60
	ActionDuplicateTab                 BindingAction = 61
	ActionMoveToSplitPane              BindingAction = 62
	ActionSendSnippet                  BindingAction = 63
)
