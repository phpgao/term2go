package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBindingActionValues tests all BindingAction constant values are correct.
func TestBindingActionValues(t *testing.T) {
	tests := []struct {
		name     string
		action   BindingAction
		expected int
	}{
		{"NEXT_SESSION", ActionNextSession, 0},
		{"NEXT_WINDOW", ActionNextWindow, 1},
		{"PREVIOUS_SESSION", ActionPreviousSession, 2},
		{"PREVIOUS_WINDOW", ActionPreviousWindow, 3},
		{"SCROLL_END", ActionScrollEnd, 4},
		{"SCROLL_HOME", ActionScrollHome, 5},
		{"SCROLL_LINE_DOWN", ActionScrollLineDown, 6},
		{"SCROLL_LINE_UP", ActionScrollLineUp, 7},
		{"SCROLL_PAGE_DOWN", ActionScrollPageDown, 8},
		{"SCROLL_PAGE_UP", ActionScrollPageUp, 9},
		{"ESCAPE_SEQUENCE", ActionEscapeSequence, 10},
		{"HEX_CODE", ActionHexCode, 11},
		{"TEXT", ActionText, 12},
		{"IGNORE", ActionIgnore, 13},
		{"IR_BACKWARD", ActionIRBackward, 15},
		{"SEND_C_H_BACKSPACE", ActionSendCHBackspace, 16},
		{"SEND_C_QM_BACKSPACE", ActionSendCQMBackspace, 17},
		{"SELECT_PANE_LEFT", ActionSelectPaneLeft, 18},
		{"SELECT_PANE_RIGHT", ActionSelectPaneRight, 19},
		{"SELECT_PANE_ABOVE", ActionSelectPaneAbove, 20},
		{"SELECT_PANE_BELOW", ActionSelectPaneBelow, 21},
		{"DO_NOT_REMAP_MODIFIERS", ActionDoNotRemapModifiers, 22},
		{"TOGGLE_FULLSCREEN", ActionToggleFullscreen, 23},
		{"REMAP_LOCALLY", ActionRemapLocally, 24},
		{"SELECT_MENU_ITEM", ActionSelectMenuItem, 25},
		{"NEW_WINDOW_WITH_PROFILE", ActionNewWindowWithProfile, 26},
		{"NEW_TAB_WITH_PROFILE", ActionNewTabWithProfile, 27},
		{"SPLIT_HORIZONTALLY_WITH_PROFILE", ActionSplitHorizontallyWithProfile, 28},
		{"SPLIT_VERTICALLY_WITH_PROFILE", ActionSplitVerticallyWithProfile, 29},
		{"NEXT_PANE", ActionNextPane, 30},
		{"PREVIOUS_PANE", ActionPreviousPane, 31},
		{"NEXT_MRU_TAB", ActionNextMRUTab, 32},
		{"MOVE_TAB_LEFT", ActionMoveTabLeft, 33},
		{"MOVE_TAB_RIGHT", ActionMoveTabRight, 34},
		{"RUN_COPROCESS", ActionRunCoprocess, 35},
		{"FIND_REGEX", ActionFindRegex, 36},
		{"SET_PROFILE", ActionSetProfile, 37},
		{"VIM_TEXT", ActionVimText, 38},
		{"PREVIOUS_MRU_TAB", ActionPreviousMRUTab, 39},
		{"LOAD_COLOR_PRESET", ActionLoadColorPreset, 40},
		{"PASTE_SPECIAL", ActionPasteSpecial, 41},
		{"PASTE_SPECIAL_FROM_SELECTION", ActionPasteSpecialFromSelection, 42},
		{"TOGGLE_HOTKEY_WINDOW_PINNING", ActionToggleHotkeyWindowPinning, 43},
		{"UNDO", ActionUndo, 44},
		{"MOVE_END_OF_SELECTION_LEFT", ActionMoveEndOfSelectionLeft, 45},
		{"MOVE_END_OF_SELECTION_RIGHT", ActionMoveEndOfSelectionRight, 46},
		{"MOVE_START_OF_SELECTION_LEFT", ActionMoveStartOfSelectionLeft, 47},
		{"MOVE_START_OF_SELECTION_RIGHT", ActionMoveStartOfSelectionRight, 48},
		{"DECREASE_HEIGHT", ActionDecreaseHeight, 49},
		{"INCREASE_HEIGHT", ActionIncreaseHeight, 50},
		{"DECREASE_WIDTH", ActionDecreaseWidth, 51},
		{"INCREASE_WIDTH", ActionIncreaseWidth, 52},
		{"SWAP_PANE_LEFT", ActionSwapPaneLeft, 53},
		{"SWAP_PANE_RIGHT", ActionSwapPaneRight, 54},
		{"SWAP_PANE_ABOVE", ActionSwapPaneAbove, 55},
		{"SWAP_PANE_BELOW", ActionSwapPaneBelow, 56},
		{"FIND_AGAIN_DOWN", ActionFindAgainDown, 57},
		{"FIND_AGAIN_UP", ActionFindAgainUp, 58},
		{"TOGGLE_MOUSE_REPORTING", ActionToggleMouseReporting, 59},
		{"INVOKE_SCRIPT_FUNCTION", ActionInvokeScriptFunction, 60},
		{"DUPLICATE_TAB", ActionDuplicateTab, 61},
		{"MOVE_TO_SPLIT_PANE", ActionMoveToSplitPane, 62},
		{"SEND_SNIPPET", ActionSendSnippet, 63},
	}

	assert.Equal(t, 63, len(tests), "should have all 63 binding action tests")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, BindingAction(tt.expected), tt.action,
				"%s should have value %d", tt.name, tt.expected)
		})
	}
}
