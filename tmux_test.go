package term2go

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestTmuxConnection_SendCommand tests TmuxConnection.SendCommand method.
func TestTmuxConnection_SendCommand(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_SendCommand_{
						SendCommand: &iterm2.TmuxResponse_SendCommand{
							Output: proto.String("list-windows output"),
						},
					},
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "conn-1"}

	output, err := tc.SendCommand(ctx, "list-windows")
	require.NoError(t, err)
	assert.Equal(t, "list-windows output", output)

	req := mc.req.GetTmuxRequest()
	sc := req.GetSendCommand()
	assert.Equal(t, "conn-1", sc.GetConnectionId())
	assert.Equal(t, "list-windows", sc.GetCommand())
}

// TestTmuxConnection_SendCommand_InvalidConn tests TmuxConnection.SendCommand with invalid connection ID.
func TestTmuxConnection_SendCommand_InvalidConn(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status:  iterm2.TmuxResponse_INVALID_CONNECTION_ID.Enum(),
					Payload: &iterm2.TmuxResponse_SendCommand_{},
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "bad-conn"}
	_, err := tc.SendCommand(ctx, "ls")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad-conn")
}

// TestTmuxConnection_SetWindowVisible tests TmuxConnection.SetWindowVisible method.
func TestTmuxConnection_SetWindowVisible(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_OK.Enum(),
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "conn-1"}

	err := tc.SetWindowVisible(ctx, "win-1", true)
	require.NoError(t, err)

	req := mc.req.GetTmuxRequest()
	swv := req.GetSetWindowVisible()
	assert.Equal(t, "conn-1", swv.GetConnectionId())
	assert.Equal(t, "win-1", swv.GetWindowId())
	assert.True(t, swv.GetVisible())
}

// TestTmuxConnection_SetWindowVisible_InvalidWindow tests TmuxConnection.SetWindowVisible with invalid window ID.
func TestTmuxConnection_SetWindowVisible_InvalidWindow(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_INVALID_WINDOW_ID.Enum(),
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "conn-1"}
	err := tc.SetWindowVisible(ctx, "bad-win", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad-win")
}

// TestTmuxConnection_CreateWindow tests TmuxConnection.CreateWindow method.
func TestTmuxConnection_CreateWindow(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_CreateWindow_{
						CreateWindow: &iterm2.TmuxResponse_CreateWindow{
							TabId: proto.String("tab-new"),
						},
					},
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "conn-1"}

	tabID, err := tc.CreateWindow(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "tab-new", tabID)

	req := mc.req.GetTmuxRequest()
	cw := req.GetCreateWindow()
	assert.Equal(t, "conn-1", cw.GetConnectionId())
	assert.Empty(t, cw.GetAffinity())
}

// TestTmuxConnection_CreateWindow_WithAffinity tests TmuxConnection.CreateWindow with session affinity.
func TestTmuxConnection_CreateWindow_WithAffinity(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status:  iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_CreateWindow_{},
				},
			},
		},
	}
	tc := &TmuxConnection{caller: mc, ConnectionID: "conn-1"}

	_, err := tc.CreateWindow(ctx, "target-session")
	require.NoError(t, err)
	assert.Equal(t, "target-session", mc.req.GetTmuxRequest().GetCreateWindow().GetAffinity())
}

// TestGetTmuxConnections tests GetTmuxConnections function.
func TestGetTmuxConnections(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_ListConnections_{
						ListConnections: &iterm2.TmuxResponse_ListConnections{
							Connections: []*iterm2.TmuxResponse_ListConnections_Connection{
								{ConnectionId: proto.String("c1"), OwningSessionId: proto.String("s1")},
								{ConnectionId: proto.String("c2"), OwningSessionId: proto.String("s2")},
							},
						},
					},
				},
			},
		},
	}

	conns, err := GetTmuxConnections(ctx, mc)
	require.NoError(t, err)
	require.Len(t, conns, 2)
	assert.Equal(t, "c1", conns[0].ConnectionID)
	assert.Equal(t, "s1", conns[0].OwningSessionID)
	assert.Equal(t, "c2", conns[1].ConnectionID)
}

// TestGetTmuxConnections_Empty tests GetTmuxConnections function returns empty list when no connections.
func TestGetTmuxConnections_Empty(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status:  iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_ListConnections_{},
				},
			},
		},
	}
	conns, err := GetTmuxConnections(ctx, mc)
	require.NoError(t, err)
	assert.Empty(t, conns)
}

// TestGetTmuxConnectionByID_Found tests GetTmuxConnectionByID function finds a connection.
func TestGetTmuxConnectionByID_Found(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status: iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_ListConnections_{
						ListConnections: &iterm2.TmuxResponse_ListConnections{
							Connections: []*iterm2.TmuxResponse_ListConnections_Connection{
								{ConnectionId: proto.String("target"), OwningSessionId: proto.String("s1")},
							},
						},
					},
				},
			},
		},
	}
	tc, err := GetTmuxConnectionByID(ctx, mc, "target")
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Equal(t, "target", tc.ConnectionID)
}

// TestGetTmuxConnectionByID_NotFound tests GetTmuxConnectionByID function returns nil when connection not found.
func TestGetTmuxConnectionByID_NotFound(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: &iterm2.ServerOriginatedMessage{
			Submessage: &iterm2.ServerOriginatedMessage_TmuxResponse{
				TmuxResponse: &iterm2.TmuxResponse{
					Status:  iterm2.TmuxResponse_OK.Enum(),
					Payload: &iterm2.TmuxResponse_ListConnections_{},
				},
			},
		},
	}
	tc, err := GetTmuxConnectionByID(ctx, mc, "nope")
	require.NoError(t, err)
	assert.Nil(t, tc)
}
