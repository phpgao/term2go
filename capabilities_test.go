package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeConn(v ProtocolVersion) *Connection {
	c := &Connection{}
	c.SetProtocolVersion(v)
	return c
}

// TestSupportsFeature tests SupportsFeature function version comparison logic.
func TestSupportsFeature(t *testing.T) {
	// v1.0 connection should satisfy v0.69 and v1.0, but not v1.1+
	conn := makeConn(ProtocolVersion{1, 0})
	assert.True(t, SupportsFeature(conn, ProtocolVersion{0, 69}))
	assert.True(t, SupportsFeature(conn, ProtocolVersion{1, 0}))
	assert.False(t, SupportsFeature(conn, ProtocolVersion{1, 1}))
	assert.False(t, SupportsFeature(conn, ProtocolVersion{1, 16}))
}

// TestAllSupportsFalseDefault tests all feature detections return false with default version.
func TestAllSupportsFalseDefault(t *testing.T) {
	// Default version (0,0) should return false for all features
	conn := &Connection{}
	assert.False(t, SupportsMultipleSetProfile(conn))
	assert.False(t, SupportsSelectPaneInDirection(conn))
	assert.False(t, SupportsPromptMonitorModes(conn))
	assert.False(t, SupportsCoprocesses(conn))
	assert.False(t, SupportsGetDefaultProfile(conn))
	assert.False(t, SupportsPromptID(conn))
	assert.False(t, SupportsListSavedArrangements(conn))
	assert.False(t, SupportsContextMenuProviders(conn))
	assert.False(t, SupportsAddAnnotation(conn))
	assert.False(t, SupportsAdvancedKeyNotifications(conn))
	assert.False(t, SupportsFilePanels(conn))
	assert.False(t, SupportsMoveSession(conn))
	assert.False(t, SupportsLoadURL(conn))
	assert.False(t, SupportsMoveSessionToTabOrWindow(conn))
	assert.False(t, SupportsApplyLayout(conn))
	assert.False(t, SupportsPromptExcludedSubranges(conn))
	assert.False(t, SupportsApplyLayoutNewSession(conn))
}

// TestAllSupportsTrueWithLatest tests all feature detections return true with latest version.
func TestAllSupportsTrueWithLatest(t *testing.T) {
	// Latest version should support all features
	conn := makeConn(ProtocolVersion{2, 0})
	assert.True(t, SupportsMultipleSetProfile(conn))
	assert.True(t, SupportsSelectPaneInDirection(conn))
	assert.True(t, SupportsPromptMonitorModes(conn))
	assert.True(t, SupportsCoprocesses(conn))
	assert.True(t, SupportsGetDefaultProfile(conn))
	assert.True(t, SupportsPromptID(conn))
	assert.True(t, SupportsListSavedArrangements(conn))
	assert.True(t, SupportsContextMenuProviders(conn))
	assert.True(t, SupportsAddAnnotation(conn))
	assert.True(t, SupportsAdvancedKeyNotifications(conn))
	assert.True(t, SupportsFilePanels(conn))
	assert.True(t, SupportsMoveSession(conn))
	assert.True(t, SupportsLoadURL(conn))
	assert.True(t, SupportsMoveSessionToTabOrWindow(conn))
	assert.True(t, SupportsApplyLayout(conn))
	assert.True(t, SupportsPromptExcludedSubranges(conn))
	assert.True(t, SupportsApplyLayoutNewSession(conn))
	assert.True(t, SupportsAdvancedKeyUp(conn))
	assert.True(t, SupportsStatusBarUnreadCount(conn))
}
