package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Property_Get gets the "grid_size" property of the first session
// and verifies the response has a non-empty JsonValue.
func TestE2E_Property_Get(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.property_get")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.property_get", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.GetProperty(ctx, caller, s.ID, "grid_size")
		require.NoError(t, err, "GetProperty(grid_size)")
		require.NotNil(t, resp, "GetProperty response should not be nil")
		require.NotNil(t, resp.JsonValue, "JsonValue should not be nil")
		require.NotEmpty(t, *resp.JsonValue, "JsonValue should not be empty")

		t.Logf("grid_size: %s", *resp.JsonValue)
		return nil
	})
}

// TestE2E_Property_SetAndRestore gets the "grid_size" property, saves the
// value, sets the same value back (no net change), and verifies no error.
func TestE2E_Property_SetAndRestore(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.property_set_restore")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.property_set_restore", func(ctx context.Context, caller term2go.Caller) error {
		// Get current value
		resp, err := term2go.GetProperty(ctx, caller, s.ID, "grid_size")
		require.NoError(t, err, "GetProperty(grid_size)")
		require.NotNil(t, resp)
		require.NotNil(t, resp.JsonValue)

		savedValue := *resp.JsonValue
		t.Logf("Saved grid_size: %s", savedValue)

		// Set the same value back (no net change)
		err = term2go.SetProperty(ctx, caller, s.ID, "grid_size", savedValue)
		require.NoError(t, err, "SetProperty(grid_size) with same value should succeed")

		return nil
	})
}

// TestE2E_Property_MultiKey gets both "grid_size" and "frame" properties
// separately and verifies each returns a non-empty result.
func TestE2E_Property_MultiKey(t *testing.T) {
	skipIfNoITerm2(t)

	app := connectAndGetApp(t, "com.term2go.test.property_multikey")
	s := firstSession(t, app)

	runWithCaller(t, "com.term2go.test.property_multikey", func(ctx context.Context, caller term2go.Caller) error {
		for _, name := range []string{"grid_size", "number_of_lines"} {
			resp, err := term2go.GetProperty(ctx, caller, s.ID, name)
			require.NoError(t, err, fmt.Sprintf("GetProperty(%s)", name))
			require.NotNil(t, resp, fmt.Sprintf("GetProperty(%s) response should not be nil", name))
			require.NotNil(t, resp.JsonValue, fmt.Sprintf("GetProperty(%s) JsonValue should not be nil", name))
			require.NotEmpty(t, *resp.JsonValue, fmt.Sprintf("GetProperty(%s) JsonValue should not be empty", name))

			t.Logf("%s: %s", name, *resp.JsonValue)
		}
		return nil
	})
}
