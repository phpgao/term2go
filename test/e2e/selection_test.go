package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

// TestE2E_Selection_Get calls SelectionRequest and verifies the response is
// valid. It logs the selection state but does not assert on content since
// the terminal may or may not have an active text selection.
func TestE2E_Selection_Get(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.selection_get")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.selection_get", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.SelectionRequest(ctx, caller, s.ID)
		require.NoError(t, err, "SelectionRequest")
		require.NotNil(t, resp, "SelectionResponse should not be nil")

		t.Logf("Selection status: %v", resp.GetStatus())

		getResp := resp.GetGetSelectionResponse()
		if getResp != nil {
			sel := getResp.GetSelection()
			t.Logf("GetSelectionResponse: selection=%v", sel)
		}

		return nil
	})
}

// TestE2E_Selection_SetAndRestore gets the current selection via GetSelection,
// then clears it by setting an empty selection. It does not try to restore
// the original selection since clearing is safe and generic.
func TestE2E_Selection_SetAndRestore(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.selection_setrestore")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.selection_setrestore", func(ctx context.Context, caller term2go.Caller) error {
		// Get current selection state (for logging only).
		getResp, err := term2go.GetSelection(ctx, caller, s.ID)
		require.NoError(t, err, "GetSelection")

		if getResp != nil {
			original := getResp.GetSelection()
			t.Logf("Original selection: %v", original)
		} else {
			t.Log("GetSelection returned nil (no active selection)")
		}

		// Clear the selection by setting an empty one.
		err = term2go.SetSelection(ctx, caller, s.ID, &iterm2.Selection{})
		assert.NoError(t, err, "SetSelection (clear)")
		t.Log("Selection cleared successfully")

		return nil
	})
}
