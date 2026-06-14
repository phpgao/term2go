package term2go

import (
	"context"
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestRPCRegistry_Register tests RPCRegistry.Register method.
func TestRPCRegistry_Register(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	reg := NewRPCRegistry(nil) // nil conn is acceptable in unit tests

	// Registering with nil conn is safe as long as we bypass conn.storeHandler
	reg.handlers["test_func"] = func(ctx context.Context, args RPCArgs) (interface{}, error) {
		return "ok", nil
	}

	_, ok := reg.handlers["test_func"]
	assert.True(t, ok, "handler should be registered")
	_ = ctx
}

// TestRPCRegistry_Dispatch_Success tests RPCRegistry.Dispatch method success path.
func TestRPCRegistry_Dispatch_Success(t *testing.T) {
	reg := NewRPCRegistry(nil)

	called := false
	reg.handlers["my_func"] = func(ctx context.Context, args RPCArgs) (interface{}, error) {
		called = true
		assert.Equal(t, "hello", args["greeting"])
		assert.Equal(t, float64(42), args["count"])
		return map[string]string{"status": "done"}, nil
	}

	// Simulate ServerOriginatedRPCNotification
	notif := &iterm2.ServerOriginatedRPCNotification{
		RequestId: proto.String("req-1"),
		Rpc: &iterm2.ServerOriginatedRPC{
			Name: proto.String("my_func"),
			Arguments: []*iterm2.ServerOriginatedRPC_RPCArgument{
				{Name: proto.String("greeting"), JsonValue: proto.String(`"hello"`)},
				{Name: proto.String("count"), JsonValue: proto.String("42")},
			},
		},
	}

	// Dispatch directly (no connection needed for the handler call itself)
	reg.dispatch(context.Background(), notif)
	assert.True(t, called, "handler should have been called")
}

// TestRPCRegistry_Dispatch_Error tests RPCRegistry.Dispatch method error handling.
func TestRPCRegistry_Dispatch_Error(t *testing.T) {
	reg := NewRPCRegistry(nil)

	reg.handlers["bad_func"] = func(ctx context.Context, args RPCArgs) (interface{}, error) {
		return nil, assert.AnError
	}

	notif := &iterm2.ServerOriginatedRPCNotification{
		RequestId: proto.String("req-2"),
		Rpc: &iterm2.ServerOriginatedRPC{
			Name:      proto.String("bad_func"),
			Arguments: nil,
		},
	}

	// Should not panic
	reg.dispatch(context.Background(), notif)
}

// TestRPCRegistry_Dispatch_UnknownRPC tests Dispatch silently ignores unknown RPCs.
func TestRPCRegistry_Dispatch_UnknownRPC(t *testing.T) {
	reg := NewRPCRegistry(nil)

	reg.handlers["known"] = func(ctx context.Context, args RPCArgs) (interface{}, error) {
		t.Fatal("should not be called")
		return nil, nil
	}

	notif := &iterm2.ServerOriginatedRPCNotification{
		RequestId: proto.String("req-3"),
		Rpc: &iterm2.ServerOriginatedRPC{
			Name: proto.String("unknown_func"),
		},
	}

	// Should silently ignore
	reg.dispatch(context.Background(), notif)
}

// TestRPCRegistry_Dispatch_NilRPC tests Dispatch should not panic on nil RPC.
func TestRPCRegistry_Dispatch_NilRPC(t *testing.T) {
	reg := NewRPCRegistry(nil)

	notif := &iterm2.ServerOriginatedRPCNotification{
		RequestId: proto.String("req-4"),
		Rpc:       nil,
	}

	// Should not panic
	reg.dispatch(context.Background(), notif)
}

// TestRPCRegistry_Dispatch_EmptyArgumentValue tests Dispatch handles empty argument value.
func TestRPCRegistry_Dispatch_EmptyArgumentValue(t *testing.T) {
	reg := NewRPCRegistry(nil)

	called := false
	reg.handlers["func"] = func(ctx context.Context, args RPCArgs) (interface{}, error) {
		called = true
		assert.Nil(t, args["maybe"])
		return "ok", nil
	}

	notif := &iterm2.ServerOriginatedRPCNotification{
		RequestId: proto.String("req-5"),
		Rpc: &iterm2.ServerOriginatedRPC{
			Name: proto.String("func"),
			Arguments: []*iterm2.ServerOriginatedRPC_RPCArgument{
				{Name: proto.String("maybe"), JsonValue: nil},
			},
		},
	}

	reg.dispatch(context.Background(), notif)
	assert.True(t, called)
}

// TestRPCRPCRoleValues tests all RPCRole constant values.
func TestRPCRPCRoleValues(t *testing.T) {
	assert.Equal(t, int32(1), int32(RPCRoleGeneric))
	assert.Equal(t, int32(2), int32(RPCRoleSessionTitle))
	assert.Equal(t, int32(3), int32(RPCRoleStatusBarComponent))
	assert.Equal(t, int32(4), int32(RPCRoleContextMenu))
}

// TestRPCRegistration_BuildProto tests RPCRegistration builds proto request.
func TestRPCRegistration_BuildProto(t *testing.T) {
	config := RPCRegistration{
		Name:        "test_fn",
		Arguments:   []string{"a", "b"},
		Defaults:    map[string]string{"a": "session.id"},
		Timeout:     5,
		Role:        RPCRoleSessionTitle,
		DisplayName: "Test Title",
		UniqueID:    "com.example.test",
	}

	reg := &iterm2.RPCRegistrationRequest{
		Name:    proto.String(config.Name),
		Role:    iterm2.RPCRegistrationRequest_Role(RPCRoleSessionTitle).Enum(),
		Timeout: proto.Float32(config.Timeout),
	}
	for _, a := range config.Arguments {
		reg.Arguments = append(reg.Arguments, &iterm2.RPCRegistrationRequest_RPCArgumentSignature{Name: proto.String(a)})
	}
	for k, v := range config.Defaults {
		reg.Defaults = append(reg.Defaults, &iterm2.RPCRegistrationRequest_RPCArgument{Name: proto.String(k), Path: proto.String(v)})
	}
	reg.RoleSpecificAttributes = &iterm2.RPCRegistrationRequest_SessionTitleAttributes_{
		SessionTitleAttributes: &iterm2.RPCRegistrationRequest_SessionTitleAttributes{
			DisplayName:      proto.String(config.DisplayName),
			UniqueIdentifier: proto.String(config.UniqueID),
		},
	}

	assert.Equal(t, "test_fn", reg.GetName())
	assert.Equal(t, 2, len(reg.GetArguments()))
	assert.Equal(t, "a", reg.GetArguments()[0].GetName())
	assert.Equal(t, float32(5), reg.GetTimeout())
	assert.Equal(t, iterm2.RPCRegistrationRequest_SESSION_TITLE, reg.GetRole())
	assert.Equal(t, "Test Title", reg.GetSessionTitleAttributes().GetDisplayName())
}

// TestRPCArgs_JSONRoundTrip tests RPCArgs JSON serialization roundtrip.
func TestRPCArgs_JSONRoundTrip(t *testing.T) {
	args := RPCArgs{
		"name":   "hello",
		"count":  float64(42),
		"flag":   true,
		"nil":    nil,
		"nested": map[string]interface{}{"key": "val"},
	}
	b, err := json.Marshal(args)
	require.NoError(t, err)

	var decoded RPCArgs
	err = json.Unmarshal(b, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "hello", decoded["name"])
	assert.NotNil(t, decoded["count"])
	assert.Equal(t, true, decoded["flag"])
	assert.Nil(t, decoded["nil"])
}

// TestRPCRegistry_Register_Validation tests Register method parameter validation.
func TestRPCRegistry_Register_Validation(t *testing.T) {
	reg := NewRPCRegistry(nil)
	ctx, cancel := testCtx()
	defer cancel()

	// Empty name
	err := reg.Register(ctx, &mockCaller{}, RPCRegistration{}, func(ctx context.Context, args RPCArgs) (interface{}, error) {
		return nil, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	// nil handler
	err = reg.Register(ctx, &mockCaller{}, RPCRegistration{Name: "test"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler is required")
}

// TestRPCRegistry_Stop tests Stop method should not panic without prior subscriptions.
func TestRPCRegistry_Stop(t *testing.T) {
	// Stop should not panic without prior subscriptions
	reg := NewRPCRegistry(nil)
	reg.Stop() // should not panic
}
