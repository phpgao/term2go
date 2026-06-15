package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// VariableScope describes the scope in which a variable can be evaluated.
type VariableScope int32

const (
	VariableScopeSession VariableScope = 0
	VariableScopeTab     VariableScope = 1
	VariableScopeWindow  VariableScope = 2
	VariableScopeApp     VariableScope = 3
)

// VariableMonitor watches for changes to an iTerm2 variable.
// Use NewVariableMonitor to create, then call Start/Stop to manage the subscription.
type VariableMonitor struct {
	conn       *Connection
	scope      VariableScope
	name       string
	identifier string
	token      NotificationToken
	ch         chan *iterm2.VariableChangedNotification
	started    bool
}

// NewVariableMonitor creates a new VariableMonitor.
//
// Parameters:
//   - conn: The Connection to iTerm2.
//   - scope: The scope (Session/Tab/Window/App).
//   - name: The variable name (e.g., "jobName", "user.myVar").
//   - identifier: A session/tab/window ID, or "all"/"active". Use "" for App scope.
func NewVariableMonitor(conn *Connection, scope VariableScope, name, identifier string) *VariableMonitor {
	return &VariableMonitor{
		conn:       conn,
		scope:      scope,
		name:       name,
		identifier: identifier,
		ch:         make(chan *iterm2.VariableChangedNotification, 8),
	}
}

// Start begins monitoring the variable. Call Stop when done.
func (m *VariableMonitor) Start(ctx context.Context) error {
	if m.started {
		return fmt.Errorf("variable monitor already started")
	}

	if err := m.subscribeWithScope(ctx); err != nil {
		return err
	}

	m.started = true
	return nil
}

func (m *VariableMonitor) subscribeWithScope(ctx context.Context) error {
	var scopeEnum iterm2.VariableScope
	switch m.scope {
	case VariableScopeSession:
		scopeEnum = iterm2.VariableScope_SESSION
	case VariableScopeTab:
		scopeEnum = iterm2.VariableScope_TAB
	case VariableScopeWindow:
		scopeEnum = iterm2.VariableScope_WINDOW
	case VariableScopeApp:
		scopeEnum = iterm2.VariableScope_APP
	}

	key := fmt.Sprintf("variable_change:%d:%s:%s:", scopeEnum, m.identifier, m.name)
	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		if n := msg.GetNotification().GetVariableChangedNotification(); n != nil {
			select {
			case m.ch <- n:
			default:
			}
		}
		return false
	}

	tk := m.conn.storeHandler(key, h)

	if err := notifyRPC(ctx, m.conn.caller(), true,
		iterm2.NotificationType_NOTIFY_ON_VARIABLE_CHANGE, "",
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_VariableMonitorRequest{
				VariableMonitorRequest: &iterm2.VariableMonitorRequest{
					Name:       proto.String(m.name),
					Scope:      scopeEnum.Enum(),
					Identifier: proto.String(m.identifier),
				},
			}
		}); err != nil {
		return err
	}

	m.token = tk
	m.token.nt = iterm2.NotificationType_NOTIFY_ON_VARIABLE_CHANGE
	m.token.sid = m.identifier
	return nil
}

// Changes returns a channel that receives variable change notifications.
func (m *VariableMonitor) Changes() <-chan *iterm2.VariableChangedNotification {
	return m.ch
}

// Stop unsubscribes from variable change notifications.
func (m *VariableMonitor) Stop() {
	if m.started {
		m.conn.Unsubscribe(m.token)
		m.started = false
	}
}

// App Variable (app scope)

// GetVariable fetches an app-level variable.
func (a *App) GetVariable(ctx context.Context, name string) (string, error) {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_App{
				App: true,
			},
			Get: []string{name},
		},
	}
	resp, err := a.caller.Call(ctx, msg)
	if err != nil {
		return "", err
	}
	if err = checkError(resp); err != nil {
		return "", err
	}
	values := resp.GetVariableResponse().GetValues()
	if len(values) > 0 {
		return jsonDecodeForVariable(values[0]), nil
	}
	return "", nil
}

// SetVariable sets an app-level variable.
func (a *App) SetVariable(ctx context.Context, name, value string) error {
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_VariableRequest{
		VariableRequest: &iterm2.VariableRequest{
			Scope: &iterm2.VariableRequest_App{
				App: true,
			},
			Set: []*iterm2.VariableRequest_Set{
				{Name: proto.String(name), Value: proto.String(ensureJSONValue(value))},
			},
		},
	}
	resp, err := a.caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	return checkError(resp)
}
