package term2go

// ProtocolVersion represents an iTerm2 protocol version.
type ProtocolVersion struct {
	Major int
	Minor int
}

// SupportsFeature checks if the connected iTerm2 supports a feature
// requiring at least the given protocol version.
func SupportsFeature(conn *Connection, min ProtocolVersion) bool {
	v := conn.ProtocolVersion()
	return v.Major > min.Major || (v.Major == min.Major && v.Minor >= min.Minor)
}

// SupportsMultipleSetProfile checks if multiple profile properties can
// be set in a single call (requires proto version >= 0.69).
func SupportsMultipleSetProfile(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{0, 69})
}

// SupportsSelectPaneInDirection checks if pane direction selection
// (left/right/up/down) is available (requires proto version >= 1.0).
func SupportsSelectPaneInDirection(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 0})
}

// SupportsPromptMonitorModes checks if different prompt monitor modes
// are available (requires proto version >= 1.1).
func SupportsPromptMonitorModes(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 1})
}

// SupportsStatusBarUnreadCount checks if the status bar can show
// an unread count (requires proto version >= 1.2).
func SupportsStatusBarUnreadCount(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 2})
}

// SupportsCoprocesses checks if coprocess manipulation is available
// (requires proto version >= 1.3).
func SupportsCoprocesses(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 3})
}

// SupportsGetDefaultProfile checks if the default profile can be
// retrieved (requires proto version >= 1.4).
func SupportsGetDefaultProfile(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 4})
}

// SupportsPromptID checks if prompts can be listed or fetched by ID
// (requires proto version >= 1.5).
func SupportsPromptID(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 5})
}

// SupportsListSavedArrangements checks if saved arrangements can be
// listed (requires proto version >= 1.6).
func SupportsListSavedArrangements(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 6})
}

// SupportsContextMenuProviders checks if context menu providers
// can be registered (requires proto version >= 1.7).
func SupportsContextMenuProviders(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 7})
}

// SupportsAddAnnotation checks if annotations can be added
// (requires proto version >= 1.8).
func SupportsAddAnnotation(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 8})
}

// SupportsAdvancedKeyNotifications checks if advanced keystroke
// notifications (key-up, flags-changed) are available
// (requires proto version >= 1.9).
func SupportsAdvancedKeyNotifications(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 9})
}

// SupportsAdvancedKeyUp is an alias for SupportsAdvancedKeyNotifications.
func SupportsAdvancedKeyUp(conn *Connection) bool {
	return SupportsAdvancedKeyNotifications(conn)
}

// SupportsFilePanels checks if open/save panels can be used
// (requires proto version >= 1.10).
func SupportsFilePanels(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 10})
}

// SupportsMoveSession checks if sessions can be moved to split panes
// (requires proto version >= 1.11).
func SupportsMoveSession(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 11})
}

// SupportsLoadURL checks if URLs can be loaded in browser sessions
// (requires proto version >= 1.12).
func SupportsLoadURL(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 12})
}

// SupportsMoveSessionToTabOrWindow checks if sessions can be moved
// to new tabs or windows (requires proto version >= 1.13).
func SupportsMoveSessionToTabOrWindow(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 13})
}

// SupportsApplyLayout checks if App.apply_layout() is available
// (requires proto version >= 1.14).
func SupportsApplyLayout(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 14})
}

// SupportsPromptExcludedSubranges checks if prompt responses include
// excluded subranges (PS2 prefixes, right-prompt cells)
// (requires proto version >= 1.15).
func SupportsPromptExcludedSubranges(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 15})
}

// SupportsApplyLayoutNewSession checks if apply_layout can create new
// sessions inline via new_session leaves
// (requires proto version >= 1.16).
func SupportsApplyLayoutNewSession(conn *Connection) bool {
	return SupportsFeature(conn, ProtocolVersion{1, 16})
}
