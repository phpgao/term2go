# term2go

Go 语言的 [iTerm2](https://iterm2.com/) WebSocket API 客户端库。

[English Documentation](README.md)

## 概述

term2go 是一个 Go 语言客户端库，镜像了官方的 iTerm2 Python API，通过 WebSocket 实现对 iTerm2 的编程控制。它提供了完整的 iTerm2 会话层级、RPC 操作、通知订阅和认证支持。

**Proto 文件**: `proto/api.proto` 文件源自 [官方 iTerm2 Python API](https://github.com/iterm2/iterm2/blob/master/api/library/python/iterm2/api/api.proto)，并与 iTerm2 发布版本保持同步。

## 功能特性

| 类别 | API |
|------|-----|
| **连接** | `Connect`, `Run`, `NewConnection`, AppleScript / 环境变量认证 |
| **层级结构** | `GetApp`, `App.Refresh`, `Window`, `Tab`, `Session`, `Splitter` |
| **会话** | `SendText`, `GetBuffer`, `Inject`, `Activate`, `Close`, `RestartSession` |
| **截图** | `Window.Screenshot`, `Tab.Screenshot`, `Session.Screenshot` — 截取窗口为 PNG |
| **窗格/标签** | `SplitPane` (垂直/水平分割), `CreateTab` (也可创建新窗口) |
| **变量** | `GetVariable`, `SetVariable` |
| **属性** | `GetProperty`, `SetProperty`, `GetProfileProperty`, `SetProfileProperty` |
| **选择** | `SelectionRequest`, `SetSelection` |
| **提示符** | `GetPrompt`, `ListPrompts` |
| **配置文件** | `ListProfiles` |
| **焦点** | `FocusRequest` |
| **对话框** | `ShowAlert`, `ShowTextInputAlert`, `ShowOpenPanel`, `ShowSavePanel`, `PolyModalAlert` |
| **触发器** | `NotificationPostedEventTrigger`, `MarkStopScrolling`, 事件触发器 |
| **快捷键绑定** | `Binding`, `ListBindings`, 快捷键绑定管理 |
| **颜色** | `ListColorPresets`, `GetColorPreset`, 颜色预设管理 |
| **状态栏** | `StatusBarComponent`, `CheckboxKnob`, `StringKnob`, `FloatKnob`, `ColorKnob` |
| **高级功能** | `PreferencesRequest`, `TmuxRequest`, `SavedArrangementRequest`, `InvokeFunction` |
| **通知** | `SubscribeNewSession`, `SubscribeKeystroke`, `SubscribeFocusChange`, `SubscribeVariableChange` 等 |
| **RPC** | `NotificationRequest`, `ServerOriginatedRPCResultRequest`, `RPCRegistry`, `RPCRegistration` |
| **功能检测** | `ProtocolVersion`, `SupportsFeature`, 版本功能检测 |

## 安装

```bash
go get github.com/phpgao/term2go
```

**要求:**
- Go 1.23+
- 启用 Python API 的 iTerm2 (Preferences → General → Magic)

## 快速开始

### 基本用法

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

### 向会话发送文本

```go
app, _ := term2go.GetApp(caller)
session := app.Windows[0].Tabs[0].Root.FindSessionByID("session-id")
session.SendText(ctx, "ls -la\n", false)
```

### 分割窗格

```go
session.SplitPane(ctx, true, false, "Default") // 垂直分割
```

### 窗口截图

```go
app, _ := term2go.GetApp(caller)
w := app.Windows[0]

// 截取整个窗口，保存为 PNG
err := w.Screenshot(ctx, "/tmp/screenshot.png")
```

`Screenshot` 同样适用于 `Tab` 和 `Session`。方法会先激活窗口提到最前，然后通过 macOS `screencapture -l` 按原生窗口 ID 直接截图。

### 订阅事件

```go
token, _ := term2go.SubscribeNewSession(caller, conn,
    func(c term2go.Caller, n *iterm2.NewSessionNotification) {
        fmt.Println("new session:", n.GetSessionId())
    })
defer conn.Unsubscribe(token)
```

## 认证

term2go 首先尝试从环境变量获取凭证，然后回退到 AppleScript:

```bash
# 方式 1: 设置环境变量 (推荐用于脚本)
export ITERM2_COOKIE="..."
export ITERM2_KEY="..."

# 方式 2: 通过 AppleScript 自动检测 (需要 iTerm2 运行)
```

## 对象层级

```
App → Window → Tab → Splitter (递归) → Session

- App:       顶层容器，持有所有窗口
- Window:    一个 iTerm2 窗口，持有标签页
- Tab:       一个标签页，持有 Splitter 根节点
- Splitter:  窗格分割容器 (递归，叶子节点是 Session)
- Session:   一个终端会话 (窗格)
```

## 示例

查看 [`example/`](example/) 获取可运行的演示:

| 示例 | 演示内容 |
|------|---------|
| [`session`](example/session/) | `SendText`, `GetBuffer`, 命令执行 |
| [`variable`](example/variable/) | `GetVariable`, `SetVariable` |
| [`pane`](example/pane/) | `SplitPane`, 遍历分屏结构 |
| [`inject`](example/inject/) | `Inject` 按键注入 |
| [`property`](example/property/) | `SetName`, `SetBadge`, `SetBuried`, `SetGridSize` |
| [`query`](example/query/) | `GetApp`, 遍历层级结构 |
| [`notification`](example/notification/) | 事件订阅 |
| [`prompt`](example/prompt/) | 交互式提示符 / 用户输入 |
| [`live`](example/live/) | 持续输出 + 轮询增量追踪 |

运行示例:

```bash
go run ./example/session
```

## 文档

```bash
go doc github.com/phpgao/term2go
```

或查看源文件中的内联文档。

## 开发

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make cover

# 运行 linter
make lint

# 重新生成 protobuf 文件
make proto

# 构建
make build
```

## 许可证

[MIT](LICENSE)
