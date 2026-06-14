package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Prompt_Get retrieves the current prompt info for the first session
// and verifies the response is non-nil. It logs command, working directory,
// and exit status for visibility.
func TestE2E_Prompt_Get(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.prompt_get")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.prompt_get", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.GetPrompt(ctx, caller, s.ID)
		require.NoError(t, err, "GetPrompt")
		require.NotNil(t, resp, "GetPrompt response should not be nil")

		if resp.Command != nil {
			t.Logf("Command: %s", *resp.Command)
		}
		if resp.WorkingDirectory != nil {
			t.Logf("Working Directory: %s", *resp.WorkingDirectory)
		}
		if resp.ExitStatus != nil {
			t.Logf("Exit Status: %d", *resp.ExitStatus)
		}
		if resp.UniquePromptId != nil {
			t.Logf("Unique Prompt ID: %s", *resp.UniquePromptId)
		}
		if resp.PromptState != nil {
			t.Logf("Prompt State: %v", *resp.PromptState)
		}

		return nil
	})
}

// TestE2E_Prompt_List lists all historical prompts for the first session
// and verifies the response is non-nil and contains at least one prompt.
func TestE2E_Prompt_List(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.prompt_list")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.prompt_list", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.ListPrompts(ctx, caller, s.ID)
		require.NoError(t, err, "ListPrompts")
		require.NotNil(t, resp, "ListPrompts response should not be nil")

		promptIDs := resp.GetUniquePromptId()
		require.NotEmpty(t, promptIDs, "expected at least one historical prompt")
		t.Logf("Total prompts: %d", len(promptIDs))
		for i, id := range promptIDs {
			t.Logf("  Prompt %d: id=%s", i, id)
		}

		return nil
	})
}
