# term2go

Go client library for the [iTerm2](https://iterm2.com/) WebSocket API.

[中文文档](README_zh.md)

## Overview

term2go is a Go client library that mirrors the official iTerm2 Python API, enabling programmatic control of iTerm2 via WebSocket. It provides comprehensive support for the full iTerm2 session hierarchy, RPC operations, notification subscriptions, and authentication.

**Proto files**: The `proto/api.proto` file is derived from the [official iTerm2 Python API](https://github.com/iterm2/iterm2/blob/master/api/library/python/iterm2/api/api.proto) and is kept in sync with iTerm2 releases.

## Features

| Category | APIs |
|----------|------|
| **Connection** | `Connect`, `Run`, `NewConnection`, AppleScript / env auth |
| **Hierarchy** | `GetApp`, `App.Refresh`, `Window`, `Tab`, `Session`, `Splitter` |
| **Session** | `SendText`, `GetBuffer`, `Inject`, `Activate`, `Close`, `RestartSession` |
| **Screenshot** | `Window.Screenshot`, `Tab.Screenshot`, `Session.Screenshot` — capture window as PNG |
| **Pane / Tab** | `SplitPane` (vertical/horizontal), `CreateTab` (also creates new windows) |
| **Variable** | `GetVariable`, `SetVariable` |
| **Property** | `GetProperty`, `SetProperty`, `GetProfileProperty`, `SetProfileProperty` |
| **Selection** | `SelectionRequest`, `SetSelection` |
| **Prompt** | `GetPrompt`, `ListPrompts` |
| **Profile** | `ListProfiles` |
| **Focus** | `FocusRequest` |
| **Alert** | `ShowAlert`, `ShowTextInputAlert`, `ShowOpenPanel`, `ShowSavePanel`, `PolyModalAlert` |
| **Trigger** | `NotificationPostedEventTrigger`, `MarkStopScrolling`, event-based triggers |
| **Binding** | `Binding`, `ListBindings`, keyboard binding management |
| **Color** | `ListColorPresets`, `GetColorPreset`, color preset management |
| **Status Bar** | `StatusBarComponent`, `CheckboxKnob`, `StringKnob`, `FloatKnob`, `ColorKnob` |
| **Advanced** | `PreferencesRequest`, `TmuxRequest`, `SavedArrangementRequest`, `InvokeFunction` |
| **Notification** | `SubscribeNewSession`, `SubscribeKeystroke`, `SubscribeFocusChange`, `SubscribeVariableChange`, and more |
| **RPC** | `NotificationRequest`, `ServerOriginatedRPCResultRequest`, `RPCRegistry`, `RPCRegistration` |
| **Capability** | `ProtocolVersion`, `SupportsFeature`, version feature detection |

## Installation

```bash
go get github.com/phpgao/term2go
```

**Requirements:**
- Go 1.23+
- iTerm2 with the Python API enabled (Preferences → General → Magic)

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/phpgao/term2go"
)

func main() {
    ctx := context.Background()
    term2go.Run(ctx, "my-app", func(caller term2go.Caller) error {
        app, _ := term2go.GetApp(caller)
        for _, w := range app.Windows {
            fmt.Printf("Window %s: %d tabs\n", w.ID, len(w.Tabs))
        }
        return nil
    })
}
```

### Send Text to Session

```go
app, _ := term2go.GetApp(caller)
session := app.Windows[0].Tabs[0].Root.FindSessionByID("session-id")
session.SendText(ctx, "ls -la\n", false)
```

### Split Pane

```go
session.SplitPane(ctx, true, false, "Default") // Vertical split
```

### Screenshot Window

```go
app, _ := term2go.GetApp(caller)
w := app.Windows[0]

// Captures the entire window as a PNG file
err := w.Screenshot(ctx, "/tmp/screenshot.png")
```

`Screenshot` is also available on `Tab` and `Session`. It activates the window, then uses macOS `screencapture -l` to capture the window directly by its native window ID.

### Subscribe to Events

```go
token, _ := term2go.SubscribeNewSession(caller, conn,
    func(c term2go.Caller, n *iterm2.NewSessionNotification) {
        fmt.Println("new session:", n.GetSessionId())
    })
defer conn.Unsubscribe(token)
```

## Authentication

term2go tries environment variables first, then falls back to AppleScript:

```bash
# Option 1: Set env vars (recommended for scripts)
export ITERM2_COOKIE="..."
export ITERM2_KEY="..."

# Option 2: Auto-detect via AppleScript (needs iTerm2 running)
```

## Object Hierarchy

```
App → Window → Tab → Splitter (recursive) → Session

- App:       top-level container, holds all windows
- Window:    an iTerm2 window, holds tabs
- Tab:       a tab, holds a Splitter root
- Splitter:  a pane-split container (recursive, leaf is Session)
- Session:   a terminal session (pane)
```

## Examples

See [`example/`](example/) for runnable demos:

| Example | What it demonstrates |
|---------|---------------------|
| [`session`](example/session/) | `SendText`, `GetBuffer`, command execution |
| [`variable`](example/variable/) | `GetVariable`, `SetVariable` |
| [`pane`](example/pane/) | `SplitPane`, traverse split structures |
| [`inject`](example/inject/) | `Inject` keystrokes into session |
| [`property`](example/property/) | `SetName`, `SetBadge`, `SetBuried`, `SetGridSize` |
| [`query`](example/query/) | `GetApp`, traverse hierarchy |
| [`notification`](example/notification/) | Event subscriptions |
| [`prompt`](example/prompt/) | Interactive prompt / user input |
| [`live`](example/live/) | Continuous output with poll-based delta tracking |

Run an example:

```bash
go run ./example/session
```

## Documentation

```bash
go doc github.com/phpgao/term2go
```

Or check the inline documentation in the source files.

## Development

```bash
# Run all tests
make test

# Run with race detector
make test

# Generate coverage report
make cover

# Run linter
make lint

# Regenerate protobuf files
make proto

# Build
make build
```

## License

[MIT](LICENSE)
