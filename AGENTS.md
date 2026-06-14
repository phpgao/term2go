# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> Agent behavior rules are in **AGENTS.md**. This file focuses on codebase-specific working knowledge.

---

## Commands

| Purpose | Command |
|---------|---------|
| Build | `make build` |
| Test (with race detector) | `make test` |
| Coverage report | `make cover` |
| Lint | `make lint` (requires `golangci-lint`) |
| Regenerate protobuf | `make proto` (requires `protoc`) |
| Run a single test | `go test -v -run TestXxx ./...` |
| Run tests in current package | `go test -v -run TestXxx -count=1 .` |

---

## Architecture

### Object Hierarchy

```
┌─[●][●][●]──────────────────────────────────── Window ────────────────┐
│  ┌─────────┐ ┌───────────┐                                             │
│  │ Tab "A" │ │  Tab "B"  │  ← Tab bar                                │
│  └─────────┘ └───────────┘                                             │
│ ┌──────────────────────┬──────────────────┐                            │
│ │                      │                  │                            │
│ │      Session         │     Session      │                            │
│ │   · ID: "uuid-1"     │   · ID: "uuid-2" │                            │
│ │   · SendText()       │   · GetBuffer()  │                            │
│ │   · GetVariable()    │   · SplitPane()  │                            │
│ │                      │                  │                            │
│ │   $ echo hi          │   $ ls           │                            │
│ │   hi                 │   src/  main.go  │                            │
│ │   $ █               │   $ █          │                            │
│ │                      │                  │                            │
│ └──────────────────────┴──────────────────┘                            │
│  ↑ Tab.Root: Splitter(Vertical=true)                                  │
│    ├─ SplitChild → Session "uuid-1"                                   │
│    └─ SplitChild → Session "uuid-2"                                   │
└────────────────────────────────────────────────────────────────────────┘

┌─[●][●][●]──────────────────────────────────── Window ────────────────┐
│ ┌─────────┐                                                            │
│ │ Tab "C" │                                                            │
│ └─────────┘                                                            │
│ ┌──────────────┬─────────────────────────┐                             │
│ │              │ ┌─────────────────────┐ │                             │
│ │   Session    │ │   Session           │ │                             │
│ │   "uuid-3"   │ │   "uuid-4"          │ │                             │
│ │              │ │                     │ │                             │
│ │   $ vim      │ │   $ top             │ │                             │
│ │   ~          │ │   PID  CPU MEM      │ │                             │
│ │   ~          │ │   123  5.2 1.3%     │ │                             │
│ │              │ │                     │ │                             │
│ └──────────────┴─────────────────────────┘                             │
│  ↑ Tab.Root: Splitter(Vertical=false)                                 │
│    ├─ SplitChild → Session "uuid-3"                                   │
│    └─ SplitChild → Splitter(Vertical=true)                            │
│          ├─ SplitChild → Session "uuid-4"                             │
│          └─ SplitChild → Session ...                                  │
└────────────────────────────────────────────────────────────────────────┘

App.Windows[]
  ├─ Window[0]
  │    ├─ Tab "A" → Splitter ─┬─ Session "uuid-1"
  │    │                      └─ Session "uuid-2"
  │    └─ Tab "B" → Splitter → Session ...
  └─ Window[1]
       └─ Tab "C" → Splitter ─┬─ Session "uuid-3"
                              └─ Splitter
                                   ├─ Session "uuid-4"
                                   └─ Session ...

App → Window → Tab → Splitter → SplitChild → Session (leaf)
                                    └──────── Splitter (nested)
```

`GetApp(caller)` → `ListSessions` RPC → proto response → tree built in `model.go`.

### Key Files

| File | Role |
|------|------|
| `conn.go` | WebSocket connection, `Caller`/`Notifier` interfaces, `dispatchLoop`, auth (env/AppleScript), `Connection` struct |
| `model.go` | App/Window/Tab/Splitter/Session types and convenience methods |
| `rpc.go` | 30+ iTerm2 RPC functions (`ListSessions`, `SendText`, `SplitPane`, …) |
| `notify.go` | Notification subscribe/unsubscribe (`SubscribeNewSession`, `SubscribeKeystroke`, …) |
| `screen.go` | Terminal screen reading (`GetBuffer`, `GetScreenContents`, …) |
| `option.go` | `WithCallTimeout`, `WithReadTimeout`, `WithHandshakeTimeout`; `OnDisconnect()` method on `Connection` |
| `lifecycle.go` | `EachSessionOnce` / `EachSessionOnceCtx` for iterating all sessions |
| `keyboard.go` | `KeystrokeMonitor`, `KeystrokeFilter` for keystroke events |
| `alert.go` | Modal alerts, text input, file panels, poly modal alerts |
| `trigger.go` | Trigger management (`GetTriggers`, `SetTriggers`, trigger types/factory) |
| `screenshot.go` | Window/Tab/Session screenshot capture via AppleScript + screencapture |
| `binding.go` | Key binding types and actions (`BindingAction`) |
| `capabilities.go` | Protocol version check (`SupportsFeature`, `SupportsMultipleSetProfile`, …) |
| `registration.go` | Server-originated RPC registration (`RPCRegistration`, `RegisterRPC`) |
| `tmux.go` | tmux integration (`TmuxConnection`, `SendCommand`) |
| `proto/api.pb.go` | Generated protobuf types — do not edit directly |

### Connection Flow

```
Connect(ctx, scriptName)
  │
  ├─ GetCookieOrCreate()
  │     ├─ Try ITERM2_COOKIE / ITERM2_KEY env vars first
  │     └─ Fall back to osascript (AppleScript) to obtain credentials
  │
  ├─ Try Unix socket: ~/Library/Application Support/iTerm2/private/socket
  └─ Fall back to TCP: localhost:1912
       └─ Start dispatchLoop goroutine
             ├─ msg.Id > 0  → deliver to pending[id] channel  (Call response)
             └─ msg.Id == 0 → fan out to registered NotificationHandlers
```

### Caller Interface (central abstraction)

```
Client                                    iTerm2
──────                                    ──────
Call(ctx, req) ──req(id=1)──▶        ──resp(id=1)──▶  Unwrap, match pending[1], return
Send()         ──req(id=2)──▶        No reply expected


                    dispatchLoop (single goroutine continuously reads WebSocket)
                    ┌─────────────────────────────────────────────┐
                    │ On receiving msg:                            │
                    │   msg.Id > 0 ──▶ Reply to Call/Send          │
                    │                   → Deliver to pending[id] channel │
                    │   msg.Id == 0 ──▶ iTerm2 pushed notification  │
                    │                   → Broadcast to all Handlers │
                    └─────────────────────────────────────────────┘
```

```go
type Caller interface {
    Call(ctx context.Context, req *ClientOriginatedMessage) (*ServerOriginatedMessage, error)
    Send(req *ClientOriginatedMessage) error
}
```

| Method | Behavior | Use Case |
|------|------|---------|
| `Call(ctx, req)` | Send request → **block waiting** for response; can return early on ctx cancellation | All RPCs: `GetApp`, `SendText`, `GetBuffer`... |
| `Send(req)` | Send request → **no reply**, returns immediately | Fire-and-forget: notification subscriptions, etc. |

**Core Mechanism**

1. Each `Call` / `Send` is assigned a unique incrementing `req.Id`
2. Outgoing requests remember `pending[id] = responseChannel`
3. `dispatchLoop` single goroutine reads WebSocket, routes by `msg.Id`:
   - `>0` → Deliver to `pending[id]` (Call gets result back)
   - `==0` → Broadcast to all `NotificationHandler` (event push)
4. All RPC functions and model methods depend on `Caller` interface → easy to mock

- `Connection` implements both `Caller` and `Notifier`

### Testing Pattern

- `wsConn` interface in `conn.go` allows injecting a mock WebSocket via `ConnectWithWS`
- `connection.dialFunc` field can replace the dialer in tests
- `connection.notifyCaller` field can inject a mock Caller for notification RPCs only
- Unit tests: root-level `*_test.go`; E2E tests: `test/e2e/` (requires a running iTerm2)

**Mock Pattern**:

```go
// 1. Mock Caller for notification RPCs
type mockCaller struct{}
func (m *mockCaller) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
    // return canned response
}
func (m *mockCaller) Send(req *iterm2.ClientOriginatedMessage) error {
    return nil
}

// 2. Inject via notifyCaller for notification tests
conn.notifyCaller = &mockCaller{}

// 3. Mock WebSocket via ConnectWithWS
conn.ConnectWithWS(mockWS)
```

---

## Example Conventions

All examples under `example/` that perform **write operations** (`SendText`, `Inject`, `SetVariable`, `SplitPane`, etc.) must:

1. Create a new window via `term2go.CreateTab(ctx, caller, "", "Default")` — pass empty windowID
2. Never operate on existing user sessions
3. Set session name and badge for identification: `SetName` + `SetBadge`
4. Clean up with `defer exampleWindow.Close(ctx)`

```go
resp, err := term2go.CreateTab(ctx, caller, "", "Default")
windowID := resp.GetWindowId()
sessionID := resp.GetSessionId()
// locate exampleWindow + newSession via GetApp + ID match
defer exampleWindow.Close(ctx)
// SetName + SetBadge for identification
// operate on newSession only
```

Read-only examples (`query`, `notification`) do not need a new window.

---

## Proto

- Edit `proto/api.proto` → run `make proto` to regenerate `proto/api.pb.go`
- Never edit `proto/api.pb.go` directly

---

## Variable JSON Encoding

iTerm2's `VariableRequest.Set.value` field (`proto/api.proto:613`) requires values to be **JSON encoded**:

```protobuf
message VariableRequest {
  message Set {
    optional string name = 1;
    optional string value = 2;  // JSON encoded
  }
}
```

**Python Reference Implementation** (`~/code/github/iterm2/api/library/python/iterm2/session.py:643-690`):

```
Set:  json.dumps(value)  → proto
Get:  proto → json.loads(raw_value)
```

**Go Implementation (two-layer symmetric)**

| Direction | Location | Function | Behavior |
|------|------|------|------|
| Set | `rpc.go` | `ensureJSONValue(v)` | If v is valid JSON → pass through; otherwise `json.Marshal(v)` encode |
| Get | `model.go` | `jsonDecodeForVariable(raw)` | `json.Unmarshal` → string extract content / number keep as-is / null → `""` |

**Note the distinction between RPC layer and model layer:**

- `rpc.GetVariable()` returns `[]string` — **raw** JSON strings, matching proto semantics
- `model.Session.GetVariable()` returns `(string, error)` — **decoded**, user gets the final value directly

**Example flow:**

```
SetVariable("user.test", "hello")
  → ensureJSONValue("hello") = "\"hello\""  (JSON encoded)
  → ...stored in iTerm2...

GetVariable("user.test")
  → iTerm2 returns "\"hello\"" (JSON string)
  → jsonDecodeForVariable("\"hello\"") = "hello"  (decoded)
```

**Related tests:**

| Test | File |
|------|------|
| `TestEnsureJSONValue` (14 cases) | `rpc_test.go` |
| `TestJsonDecodeForVariable` (9 cases) | `model_test.go` |
| `TestE2E_CreateWindow_Variable` | `test/e2e/create_window_variable_test.go` |

---

## Key Types

| Type | File | Description |
|------|------|-------------|
| `App` | `model.go` | Top-level container: `Windows []Window` |
| `Window` | `model.go` | Has `ID`, `Tabs []Tab`, methods: `Close`, `Screenshot` |
| `Tab` | `model.go` | Has `ID`, `Root *Splitter`, methods: `Select`, `Screenshot` |
| `Splitter` | `splitter.go` | Binary tree: `Vertical bool`, `SplitChild []interface{}` |
| `Session` | `session.go` | Has `ID`, `caller`; methods: `SendText`, `GetBuffer`, `SplitPane`, `Close`, … |
| `Connection` | `conn.go` | WebSocket connection; implements `Caller` + `Notifier` |
| `NotificationToken` | `notify.go` | Opaque token for unsubscribe |
| `KeystrokeMonitor` | `keyboard.go` | Streams keystroke events |
| `KeystrokeFilter` | `keyboard.go` | Intercepts keystrokes |
| `Trigger` | `trigger.go` | Unified trigger representation |
| `PolyModalAlert` | `alert.go` | Builder for complex modal alerts |
| `ProtocolVersion` | `capabilities.go` | iTerm2 protocol version check |
| `TmuxConnection` | `tmux.go` | tmux integration |

---

## Notes

- This project is **Go-only** (no Python dependency at runtime).
- All iTerm2 communication is over a **local WebSocket** (Unix socket or TCP `localhost:1912`).
- The `proto/` directory contains the single `.proto` file; generated Go code is `proto/api.pb.go`.
- Examples in `example/` are runnable; read-only ones need only a running iTerm2.
- E2E tests (`test/e2e/`) require a running iTerm2 and will create/close windows/tabs.
