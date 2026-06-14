package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Variable_GetBuiltins reads built-in session variables (jobName,
// columns, rows) via term2go.GetVariable and verifies they are all non-empty.
func TestE2E_Variable_GetBuiltins(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-variable-getbuiltins")
	require.NoError(t, err)
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err)
	s := firstSession(t, app)

	values, err := term2go.GetVariable(ctx, conn, s.ID, []string{"jobName", "columns", "rows"})
	require.NoError(t, err)
	require.Len(t, values, 3)

	assert.NotEmpty(t, values[0], "jobName should not be empty")
	assert.NotEmpty(t, values[1], "columns should not be empty")
	assert.NotEmpty(t, values[2], "rows should not be empty")
}

// TestE2E_Variable_SetAndRestore sets a user variable, reads it back, and
// restores the original value so the test is side-effect free.
func TestE2E_Variable_SetAndRestore(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-variable-setrestore")
	require.NoError(t, err)
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err)
	s := firstSession(t, app)

	// Read original value (may be empty if never set).
	orig, _ := s.GetVariable(ctx, "user.e2e_test")

	const newVal = "e2e-test-value"
	err = s.SetVariable(ctx, "user.e2e_test", newVal)
	require.NoError(t, err)

	got, err := s.GetVariable(ctx, "user.e2e_test")
	require.NoError(t, err)
	assert.Equal(t, newVal, got)

	// Restore original value if it existed.
	if orig != "" {
		err = s.SetVariable(ctx, "user.e2e_test", orig)
		require.NoError(t, err)
	}
}

// TestE2E_Variable_ModelMethod verifies that the Session.GetVariable model
// method returns a decoded (non-empty) value for a built-in variable.
func TestE2E_Variable_ModelMethod(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-variable-model")
	require.NoError(t, err)
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err)
	s := firstSession(t, app)

	name, err := s.GetVariable(ctx, "jobName")
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}
