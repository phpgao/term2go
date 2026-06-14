package e2e

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_RPC_Tmux(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-rpc-tmux", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.TmuxRequest(ctx, caller, &iterm2.TmuxRequest{
			Payload: &iterm2.TmuxRequest_ListConnections_{
				ListConnections: &iterm2.TmuxRequest_ListConnections{},
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		return nil
	})
}

func TestE2E_RPC_Preferences(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-rpc-prefs", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.PreferencesRequest(ctx, caller, &iterm2.PreferencesRequest{
			Requests: []*iterm2.PreferencesRequest_Request{{
				Request: &iterm2.PreferencesRequest_Request_GetPreferenceRequest{
					GetPreferenceRequest: &iterm2.PreferencesRequest_Request_GetPreference{
						Key: proto.String("__iterm2_preferences_version__"),
					},
				},
			}},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		return nil
	})
}

func TestE2E_RPC_SavedArrangement(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-rpc-arrangement", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.SavedArrangementRequest(ctx, caller, &iterm2.SavedArrangementRequest{
			Action: iterm2.SavedArrangementRequest_LIST.Enum(),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		return nil
	})
}

func TestE2E_RPC_InvokeFunction(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-rpc-invoke", func(ctx context.Context, caller term2go.Caller) error {
		resp, err := term2go.InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{
			Invocation: proto.String("window.set_title"),
			Timeout:    proto.Float64(5),
		})
		// May error if no Python API loaded, just verify no panic
		if err != nil {
			t.Logf("InvokeFunction returned error (may be expected): %v", err)
			return nil
		}
		assert.NotNil(t, resp)
		return nil
	})
}

func TestE2E_RPC_ServerOriginatedRPC(t *testing.T) {
	skipIfNoITerm2(t)
	runWithCaller(t, "term2go-e2e-rpc-sorpc", func(ctx context.Context, caller term2go.Caller) error {
		fakeReqID := "term2go-e2e-fake-rpc-id"
		err := term2go.ServerOriginatedRPCResultRequest(ctx, caller, &iterm2.ServerOriginatedRPCResultRequest{
			RequestId: proto.String(fakeReqID),
			Result: &iterm2.ServerOriginatedRPCResultRequest_JsonValue{
				JsonValue: `{"status": "ok"}`,
			},
		})
		// May error if fake ID not recognized, just verify format is valid
		if err != nil {
			t.Logf("ServerOriginatedRPCResultRequest returned error (may be expected): %v", err)
		}
		return nil
	})
}
