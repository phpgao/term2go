package term2go

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// RPCArgs / RPCHandler / RPCRegistration

// RPCArgs holds named arguments from an iTerm2 server-originated RPC invocation.
// Values are JSON-decoded from the notification's argument list.
type RPCArgs map[string]interface{}

// RPCHandler is a function that processes a server-originated RPC.
// It receives the context and parsed arguments, and returns a JSON-serializable
// result or an error. Errors are sent back to iTerm2 as exceptions.
type RPCHandler func(ctx context.Context, args RPCArgs) (interface{}, error)

// RPCRegistration configures how an RPC is registered with iTerm2.
// Corresponds to Python's registration.RPC decorator parameters.
type RPCRegistration struct {
	Name      string            // RPC function name iTerm2 uses to invoke it
	Arguments []string          // argument names in the RPC signature
	Defaults  map[string]string // default key → variable path (like Python's Reference)
	Timeout   float32           // seconds iTerm2 waits; 0 means use default
	// Role-specific fields
	Role        RPCRPCRole // GENERIC / SESSION_TITLE / STATUS_BAR_COMPONENT / CONTEXT_MENU
	DisplayName string     // for SESSION_TITLE / CONTEXT_MENU roles
	UniqueID    string     // unique identifier (reverse DNS), required for non-GENERIC roles
	// StatusBarComponent is embedded in the registration for STATUS_BAR_COMPONENT role.
	StatusBarComponent *StatusBarComponent
}

// RPCRPCRole mirrors iterm2.RPCRegistrationRequest_Role.
type RPCRPCRole int32

const (
	RPCRoleGeneric            RPCRPCRole = 1
	RPCRoleSessionTitle       RPCRPCRole = 2
	RPCRoleStatusBarComponent RPCRPCRole = 3
	RPCRoleContextMenu        RPCRPCRole = 4
)

// RPCRegistry manages registered RPC handlers and dispatches incoming
// ServerOriginatedRPCNotification messages to the correct handler.
//
// Usage:
//
//	reg := NewRPCRegistry(conn)
//	reg.Register(ctx, conn, RPCRegistration{
//	    Name:      "my_function",
//	    Arguments: []string{"arg1"},
//	}, func(ctx context.Context, args RPCArgs) (interface{}, error) {
//	    return "ok", nil
//	})
//	// Block until connection closes:
//	select {}
type RPCRegistry struct {
	conn     *Connection
	handlers map[string]RPCHandler
	mu       sync.RWMutex
	token    NotificationToken
}

// NewRPCRegistry creates a new RPC registry for the given connection.
func NewRPCRegistry(conn *Connection) *RPCRegistry {
	return &RPCRegistry{
		conn:     conn,
		handlers: make(map[string]RPCHandler),
	}
}

// Register registers an RPC handler with iTerm2 using the full registration config.
// Returns an error if the handler could not be registered.
func (r *RPCRegistry) Register(ctx context.Context, caller Caller, config RPCRegistration, handler RPCHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if config.Name == "" {
		return fmt.Errorf("RPC registration: name is required")
	}
	if handler == nil {
		return fmt.Errorf("RPC registration: handler is required for %q", config.Name)
	}

	// Build the proto registration request
	reg := &iterm2.RPCRegistrationRequest{
		Name: proto.String(config.Name),
		Role: (*iterm2.RPCRegistrationRequest_Role)(proto.Int32(int32(config.Role))).Enum(),
	}
	if len(config.Arguments) > 0 {
		for _, a := range config.Arguments {
			reg.Arguments = append(reg.Arguments, &iterm2.RPCRegistrationRequest_RPCArgumentSignature{Name: proto.String(a)})
		}
	}
	if len(config.Defaults) > 0 {
		for k, v := range config.Defaults {
			reg.Defaults = append(reg.Defaults, &iterm2.RPCRegistrationRequest_RPCArgument{Name: proto.String(k), Path: proto.String(v)})
		}
	}
	if config.Timeout > 0 {
		reg.Timeout = proto.Float32(config.Timeout)
	}

	// Role-specific attributes
	switch config.Role {
	case RPCRoleSessionTitle:
		reg.RoleSpecificAttributes = &iterm2.RPCRegistrationRequest_SessionTitleAttributes_{
			SessionTitleAttributes: &iterm2.RPCRegistrationRequest_SessionTitleAttributes{
				DisplayName:      proto.String(config.DisplayName),
				UniqueIdentifier: proto.String(config.UniqueID),
			},
		}
	case RPCRoleContextMenu:
		reg.RoleSpecificAttributes = &iterm2.RPCRegistrationRequest_ContextMenuAttributes_{
			ContextMenuAttributes: &iterm2.RPCRegistrationRequest_ContextMenuAttributes{
				DisplayName:      proto.String(config.DisplayName),
				UniqueIdentifier: proto.String(config.UniqueID),
			},
		}
	case RPCRoleStatusBarComponent:
		if config.StatusBarComponent != nil {
			attrs := &iterm2.RPCRegistrationRequest_StatusBarComponentAttributes{
				ShortDescription:    proto.String(config.StatusBarComponent.ShortDescription),
				DetailedDescription: proto.String(config.StatusBarComponent.DetailedDescription),
				Exemplar:            proto.String(config.StatusBarComponent.Exemplar),
				UniqueIdentifier:    proto.String(config.StatusBarComponent.Identifier),
			}
			if config.StatusBarComponent.UpdateCadence > 0 {
				attrs.UpdateCadence = proto.Float32(float32(config.StatusBarComponent.UpdateCadence))
			}
			reg.RoleSpecificAttributes = &iterm2.RPCRegistrationRequest_StatusBarComponentAttributes_{
				StatusBarComponentAttributes: attrs,
			}
		}
	}

	// Subscribe to server-originated RPC notifications with the registration
	if err := r.subscribe(ctx, caller, reg); err != nil {
		return fmt.Errorf("RPC registration for %q: %w", config.Name, err)
	}

	// Store the handler
	r.handlers[config.Name] = handler
	return nil
}

func (r *RPCRegistry) subscribe(ctx context.Context, caller Caller, reg *iterm2.RPCRegistrationRequest) error {
	key := "server_originated_rpc:"

	if err := doSubscribe(ctx, caller, iterm2.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC, "",
		func(nr *iterm2.NotificationRequest) {
			nr.Arguments = &iterm2.NotificationRequest_RpcRegistrationRequest{
				RpcRegistrationRequest: reg,
			}
		},
	); err != nil {
		return err
	}

	h := func(msg *iterm2.ServerOriginatedMessage) bool {
		n := msg.GetNotification().GetServerOriginatedRpcNotification()
		if n == nil {
			return false
		}
		r.dispatch(context.Background(), n)
		return false
	}

	r.token = r.conn.storeHandler(key, h)
	r.token.nt = iterm2.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC
	return nil
}

// Stop unsubscribes from RPC notifications.
func (r *RPCRegistry) Stop() {
	if r.conn != nil {
		r.conn.Unsubscribe(r.token)
	}
}

// dispatch handles a single ServerOriginatedRPCNotification.
func (r *RPCRegistry) dispatch(ctx context.Context, n *iterm2.ServerOriginatedRPCNotification) {
	rpc := n.GetRpc()
	if rpc == nil {
		return
	}
	name := rpc.GetName()
	reqID := n.GetRequestId()

	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return
	}

	// Parse arguments
	args := make(RPCArgs)
	for _, arg := range rpc.GetArguments() {
		var v interface{}
		if raw := arg.GetJsonValue(); raw != "" {
			_ = json.Unmarshal([]byte(raw), &v)
		}
		args[arg.GetName()] = v
	}

	// Execute handler
	result, err := handler(ctx, args)

	var respReq iterm2.ServerOriginatedRPCResultRequest
	respReq.RequestId = proto.String(reqID)

	if err != nil {
		exception := map[string]string{
			"reason": err.Error(),
		}
		b, _ := json.Marshal(exception)
		respReq.Result = &iterm2.ServerOriginatedRPCResultRequest_JsonException{
			JsonException: string(b),
		}
	} else {
		b, _ := json.Marshal(result)
		respReq.Result = &iterm2.ServerOriginatedRPCResultRequest_JsonValue{
			JsonValue: string(b),
		}
	}

	// Fire and forget — don't block notification dispatch
	go func() {
		if r.conn != nil {
			_ = ServerOriginatedRPCResultRequest(ctx, r.conn, &respReq)
		}
	}()
}
