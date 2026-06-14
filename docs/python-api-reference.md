# Python-to-Go API Mapping Reference

> Auto-generated from `/Users/jimmy/code/github/iTerm2/api/library/python/iterm2/iterm2/` → `/Users/jimmy/code/github/term2go/`

---

## alert.py
- Purpose: Modal dialogs and alerts
- Types: Alert, TextInputAlert, PolyModalResult, PolyModalAlert
-→ Go: P2 pending

## app.py
- Purpose: iTerm2 application object, app-level functions
- Types: App
- Functions: async_get_app(...), invalidate_app(), async_get_variable(...), async_invoke_function(...)
-→ Go: model.go (App type + GetApp), rpc.go (InvokeFunction)

## arrangement.py
- Purpose: Saved window arrangements
- Types: Arrangement, SavedArrangementException
-→ Go: rpc.go (SavedArrangementRequest), no full types — P2 pending

## auth.py
- Purpose: Cookie/key authentication via env/AppleScript
- Types: AppKitApplescriptRunner, CommandLineApplescriptRunner, LSBackgroundContextManager, AuthenticationException
- Functions: get_script_name(), authenticate(), request_cookie_and_key(...)
-→ Go: conn.go (AuthProvider, EnvAuthProvider, AppleScriptAuthProvider, Connect flow)

## binding.py
- Purpose: Key bindings, global key assignments, paste config
- Types: PasteConfiguration, SnippetIdentifier, BindingAction, KeyBinding, MoveSelectionUnit
- Functions: async_get_global_key_bindings(...), async_set_global_key_bindings(...)
-→ Go: binding.go

## broadcast.py
- Purpose: Input broadcasting domain management
- Types: BroadcastDomain
- Functions: async_set_broadcast_domains(...)
-→ Go: P2 pending

## capabilities.py
- Purpose: iTerm2 version-capability checks
- Types: AppVersionTooOld
- Functions: supports_multiple_set_profile_properties(...), supports_prompt_monitor_modes(...), supports_coprocesses(...), etc.
-→ Go: P2 pending

## color.py
- Purpose: Color representation and ColorSpace enum
- Types: Color, ColorSpace(Enum), MissingDependency
-→ Go: P2 pending

## colorpresets.py
- Purpose: Color preset listing and retrieval
- Types: ColorPreset, ListPresetsException, GetPresetException
-→ Go: P2 pending

## connection.py
- Purpose: WebSocket connection management, run loops
- Types: Connection
- Functions: run_until_complete(...), run_forever(...), add_disconnect_callback(...)
-→ Go: conn.go (Connection type + Connect, dispatchLoop, Call/Send)

## customcontrol.py
- Purpose: Custom escape sequence monitoring
- Types: CustomControlSequenceMonitor
-→ Go: P2 pending

## filepanel.py
- Purpose: File Open/Save panel dialogs
- Types: OpenPanel, SavePanel
-→ Go: P2 pending

## focus.py
- Purpose: Keyboard focus change monitoring
- Types: FocusMonitor, FocusUpdate, FocusUpdateApplicationActive, FocusUpdateWindowChanged, etc.
-→ Go: focus.go

## keyboard.py
- Purpose: Keystroke monitoring and filtering
- Types: Modifier(Enum), Keycode(Enum), Keystroke, KeystrokePattern, KeystrokeMonitor, KeystrokeFilter
-→ Go: keyboard.go

## lifecycle.py
- Purpose: Session lifecycle event monitors
- Types: EachSessionOnceMonitor, SessionTerminationMonitor, LayoutChangeMonitor, NewSessionMonitor
-→ Go: P2 pending

## mainmenu.py
- Purpose: Menu item identification and state
- Types: MenuItemIdentifier, MenuItemState, MainMenu, MenuItemException
-→ Go: P2 pending

## notifications.py
- Purpose: Subscribe/unsubscribe from async iTerm2 notifications
- Types: SubscriptionException
- Functions: async_unsubscribe(...), async_subscribe_to_new_session_notification(...), async_subscribe_to_keystroke_notification(...), async_subscribe_to_screen_update_notification(...), etc.
-→ Go: notify.go

## preferences.py
- Purpose: iTerm2 preference get/set
- Types: PreferenceKey(Enum)
- Functions: async_get_preference(...), async_set_preference(...)
-→ Go: rpc.go (PreferencesRequest) — thin RPC layer, no full types — P2 pending

## profile.py
- Purpose: Session profile properties (huge file: 6800+ lines)
- Types: LocalWriteOnlyProfile, WriteOnlyProfile, Profile, PartialProfile + 12 Enums (CursorType, ThinStrokes, etc.)
-→ Go: P2 pending

## prompt.py
- Purpose: Shell prompt detection and information
- Types: Prompt, PromptState(Enum), PromptMonitor
- Functions: async_get_last_prompt(...), async_get_prompt_by_id(...), async_list_prompts(...)
-→ Go: prompt.go

## registration.py
- Purpose: RPC function registration decorators
- Types: Reference
- Functions: RPC(func), ContextMenuProviderRPC(func), TitleProviderRPC(func), StatusBarRPC(func), generic_handle_rpc(...)
-→ Go: P2 pending

## rpc.py
- Purpose: Low-level RPC request builders for all iTerm2 operations
- Types: RPCException
- Functions: async_list_sessions(...), async_send_text(...), async_split_pane(...), async_create_tab(...), async_get_screen_contents(...), async_start_transaction(...), etc.
-→ Go: rpc.go (ListSessions, SendText, GetBuffer, CreateTab, SplitPane, etc.)

## screen.py
- Purpose: Screen content and streaming
- Types: CellStyle, LineContents, ScreenContents, ScreenStreamer
-→ Go: screen.go

## selection.py
- Purpose: Text selection representation and interaction
- Types: SelectionMode(Enum), SubSelection, Selection
-→ Go: rpc.go (SelectionRequest, GetSelection) — thin RPC layer only — P2 pending

## session.py
- Purpose: Session representation and operations (1100+ lines)
- Types: Session, Splitter, ProxySession, SessionLineInfo, InvalidSessionId, SplitPaneException
-→ Go: model.go (Session type + Splitter + SplitChild)

## statusbar.py
- Purpose: Status bar component customization knobs
- Types: StatusBarComponent, BaseKnob, Knob, CheckboxKnob, StringKnob, PositiveFloatingPointKnob, ColorKnob
-→ Go: P2 pending

## tab.py
- Purpose: Tab navigation and management
- Types: Tab, NavigationDirection(Enum)
-→ Go: model.go (Tab type + Select, Close, UpdateLayout)

## tmux.py
- Purpose: Tmux integration connections
- Types: TmuxConnection, TmuxException, Delegate(ABC)
- Functions: async_get_tmux_connections(...), async_get_tmux_connection_by_connection_id(...)
-→ Go: tmux.go

## tool.py
- Purpose: WebView tool registration (thin wrapper)
- Functions: async_register_web_view_tool(...)
-→ Go: P2 pending

## transaction.py
- Purpose: Atomic transaction scope
- Types: Transaction
-→ Go: P2 pending (rpc.go has StartTransaction/EndTransaction RPCs, no Transaction type)

## triggers.py
- Purpose: Trigger definitions for pattern matching (22 trigger types)
- Types: MatchType(Enum), Trigger, AlertTrigger, AnnotateTrigger, BellTrigger, BounceTrigger, BufferInputTrigger, RPCTrigger, CaptureTrigger, SetNamedMarkTrigger, SGRTrigger, FoldTrigger, InjectTrigger, HighlightLineTrigger, UserNotificationTrigger, SetUserVariableTrigger, ShellPromptTrigger, SetTitleTrigger, SendTextTrigger, RunCommandTrigger
- Functions: decode_trigger(...)
-→ Go: trigger.go

## util.py
- Purpose: Geometry types (Size, Point, Frame) and encoding helpers
- Types: Size, Point, Frame, CoordRange, Range, WindowedCoordRange
- Functions: frame_str(...), size_str(...), point_str(...), distance(...), iterm2_encode(...), invocation_string(...)
-→ Go: model.go (Point, Size, WindowFrame) + util.go (Coord, CoordRange)

## variables.py
- Purpose: Variable monitoring (user-defined and system variables)
- Types: VariableScopes(Enum), VariableMonitor
-→ Go: model.go (GetVariable/SetVariable on Session) — thin, no monitor type — P2 pending

## window.py
- Purpose: Window management (create/close tabs, properties)
- Types: Window, CreateTabException, CreateWindowException, SetPropertyException, GetPropertyException
-→ Go: model.go (Window type + CreateTab, Close)

---

## Summary

| Status | Count | Files |
|--------|-------|-------|
| Implemented | 16 | app, auth, binding, connection, focus, keyboard, notifications, prompt, rpc, screen, session, tab, tmux, triggers, util, window |
| P2 pending | 18 | alert, arrangement, broadcast, capabilities, color, colorpresets, customcontrol, filepanel, lifecycle, mainmenu, preferences, profile, registration, selection, statusbar, tool, transaction, variables |
