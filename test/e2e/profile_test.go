package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phpgao/term2go"
)

// TestE2E_Profile_List verifies that ListProfiles returns at least one profile
// and that each profile has non-empty Name and Guid properties.
func TestE2E_Profile_List(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-profile-list", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.ListProfiles(ctx, caller, []string{"Name", "Guid"}, nil)
		if err != nil {
			return err
		}
		profiles := resp.GetProfiles()
		require.NotEmpty(t, profiles, "expected at least one profile")

		for _, p := range profiles {
			var name, guid string
			for _, prop := range p.GetProperties() {
				switch prop.GetKey() {
				case "Name":
					name = prop.GetJsonValue()
				case "Guid":
					guid = prop.GetJsonValue()
				}
			}
			assert.NotEmpty(t, name, "profile should have a Name")
			assert.NotEmpty(t, guid, "profile should have a Guid")
		}
		return nil
	})
}

// TestE2E_Profile_ListFiltered lists all profiles, picks the Guid of the first
// one, then calls ListProfiles filtered by that Guid and expects exactly one
// profile.
func TestE2E_Profile_ListFiltered(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-profile-listfiltered", func(ctx context.Context, caller term2go.Caller) error {
		// List all profiles to get one Guid.
		resp, err := term2go.ListProfiles(ctx, caller, []string{"Guid"}, nil)
		if err != nil {
			return err
		}
		profiles := resp.GetProfiles()
		require.NotEmpty(t, profiles, "expected at least one profile")

		// Extract first profile's Guid (JsonValue is JSON-encoded).
		var firstGUID string
		for _, prop := range profiles[0].GetProperties() {
			if prop.GetKey() == "Guid" {
				_ = json.Unmarshal([]byte(prop.GetJsonValue()), &firstGUID)
				break
			}
		}
		require.NotEmpty(t, firstGUID, "first profile should have a Guid")

		// Filter by that Guid.
		filtered, err := term2go.ListProfiles(ctx, caller, nil, []string{firstGUID})
		if err != nil {
			return err
		}
		require.Len(t, filtered.GetProfiles(), 1, "expected exactly 1 profile when filtering by Guid")
		return nil
	})
}

// TestE2E_Profile_GetProperty reads the Name and Guid profile properties from
// the first session and verifies both are non-empty.
func TestE2E_Profile_GetProperty(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-profile-getproperty")
	require.NoError(t, err)
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err)
	s := firstSession(t, app)

	resp, err := term2go.GetProfileProperty(ctx, conn, s.ID, []string{"Name", "Guid"})
	require.NoError(t, err)

	var name, guid string
	for _, prop := range resp.GetProperties() {
		switch prop.GetKey() {
		case "Name":
			name = prop.GetJsonValue()
		case "Guid":
			guid = prop.GetJsonValue()
		}
	}
	assert.NotEmpty(t, name, "profile Name should not be empty")
	assert.NotEmpty(t, guid, "profile Guid should not be empty")
}

// TestE2E_Profile_SetAndRestore reads the current Name profile property and
// sets it back to the same value (no net change), verifying the operation
// completes without error.
func TestE2E_Profile_SetAndRestore(t *testing.T) {
	skipIfNoITerm2(t)
	ctx := context.Background()
	conn, err := term2go.Connect(ctx, "term2go-e2e-profile-setrestore")
	require.NoError(t, err)
	defer conn.Close()

	app, err := term2go.GetApp(ctx, conn)
	require.NoError(t, err)
	s := firstSession(t, app)

	// Get current Name value.
	resp, err := term2go.GetProfileProperty(ctx, conn, s.ID, []string{"Name"})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetProperties(), "expected Name property")

	jsonValue := resp.GetProperties()[0].GetJsonValue()
	require.NotEmpty(t, jsonValue, "profile Name value should not be empty")

	// Set the same value back (no net change).
	err = term2go.SetProfileProperty(ctx, conn, s.ID, "Name", jsonValue)
	require.NoError(t, err)
}
