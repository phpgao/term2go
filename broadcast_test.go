package term2go

import (
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

func TestBroadcastDomainFromProto(t *testing.T) {
	proto := &iterm2.BroadcastDomain{
		SessionIds: []string{"s1", "s2"},
	}
	d := BroadcastDomainFromProto(proto)
	assert.Equal(t, []string{"s1", "s2"}, d.SessionIDs())
}

func TestBroadcastDomainFromProto_Empty(t *testing.T) {
	d := BroadcastDomainFromProto(&iterm2.BroadcastDomain{})
	assert.Empty(t, d.SessionIDs())
}

func TestBroadcastDomain_AddSession(t *testing.T) {
	d := &BroadcastDomain{}
	d.AddSession("s1")
	d.AddSession("s2")
	assert.Equal(t, []string{"s1", "s2"}, d.SessionIDs())
}

func TestBroadcastDomain_RemoveSession(t *testing.T) {
	d := &BroadcastDomain{sessions: []string{"s1", "s2", "s3"}}
	d.RemoveSession("s2")
	assert.Equal(t, []string{"s1", "s3"}, d.SessionIDs())
}

func TestBroadcastDomain_RemoveSession_NotFound(t *testing.T) {
	d := &BroadcastDomain{sessions: []string{"s1", "s2"}}
	d.RemoveSession("s3")
	assert.Equal(t, []string{"s1", "s2"}, d.SessionIDs())
}

func TestBroadcastDomain_Contains(t *testing.T) {
	d := &BroadcastDomain{sessions: []string{"s1", "s2"}}
	assert.True(t, d.Contains("s1"))
	assert.False(t, d.Contains("s3"))
}

func TestBroadcastDomain_ToProto(t *testing.T) {
	d := &BroadcastDomain{sessions: []string{"s1", "s2"}}
	p := d.ToProto()
	assert.Equal(t, []string{"s1", "s2"}, p.GetSessionIds())
}

func TestBroadcastDomain_ToProto_Empty(t *testing.T) {
	d := &BroadcastDomain{}
	p := d.ToProto()
	assert.Empty(t, p.GetSessionIds())
}

func TestBroadcastDomains_FromProto(t *testing.T) {
	proto := []*iterm2.BroadcastDomain{
		{SessionIds: []string{"s1", "s2"}},
		{SessionIds: []string{"s3"}},
	}
	var bds BroadcastDomains
	bds.FromProto(proto)
	require.Len(t, bds, 2)
	assert.Equal(t, []string{"s1", "s2"}, bds[0].SessionIDs())
	assert.Equal(t, []string{"s3"}, bds[1].SessionIDs())
}

func TestBroadcastDomains_FromProto_Empty(t *testing.T) {
	var bds BroadcastDomains
	bds.FromProto(nil)
	assert.Empty(t, bds)
}

func TestBroadcastDomains_ToProto(t *testing.T) {
	bds := BroadcastDomains{
		{sessions: []string{"s1"}},
		{sessions: []string{"s2", "s3"}},
	}
	proto := bds.ToProto()
	require.Len(t, proto, 2)
	assert.Equal(t, []string{"s1"}, proto[0].GetSessionIds())
	assert.Equal(t, []string{"s2", "s3"}, proto[1].GetSessionIds())
}

func TestBroadcastDomains_ToProto_Empty(t *testing.T) {
	var bds BroadcastDomains
	assert.Empty(t, bds.ToProto())
}

func TestBroadcastDomains_FindForSession(t *testing.T) {
	bds := BroadcastDomains{
		{sessions: []string{"s1", "s2"}},
		{sessions: []string{"s3"}},
	}
	assert.NotNil(t, bds.FindForSession("s1"))
	assert.NotNil(t, bds.FindForSession("s3"))
	assert.Nil(t, bds.FindForSession("s4"))
}

func TestBroadcastDomains_Set(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	bds := BroadcastDomains{
		{sessions: []string{"s1", "s2"}},
	}
	err := bds.Set(ctx, mc)
	require.NoError(t, err)
	req := mc.req.GetSetBroadcastDomainsRequest()
	require.NotNil(t, req)
	require.Len(t, req.GetBroadcastDomains(), 1)
}

// App.RefreshBroadcastDomains

func TestApp_RefreshBroadcastDomains(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetBroadcastDomainsResponse{
			BroadcastDomains: []*iterm2.BroadcastDomain{
				{SessionIds: []string{"s1", "s2"}},
			},
		}),
	}
	app := &App{caller: mc}
	err := app.RefreshBroadcastDomains(ctx)
	require.NoError(t, err)
	require.Len(t, app.BroadcastDomains, 1)
	assert.Equal(t, []string{"s1", "s2"}, app.BroadcastDomains[0].SessionIDs())
}

func TestApp_RefreshBroadcastDomains_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("conn error")}
	app := &App{caller: mc}
	err := app.RefreshBroadcastDomains(ctx)
	require.Error(t, err)
}

func TestGetBroadcastDomains_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetBroadcastDomainsResponse{
			BroadcastDomains: []*iterm2.BroadcastDomain{
				{SessionIds: []string{"s1"}},
			},
		}),
	}
	resp, err := GetBroadcastDomains(ctx, mc)
	require.NoError(t, err)
	assert.Len(t, resp.GetBroadcastDomains(), 1)
}

func TestGetBroadcastDomains_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	_, err := GetBroadcastDomains(ctx, mc)
	require.Error(t, err)
}

func TestSetBroadcastDomains_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	err := SetBroadcastDomains(ctx, mc, []*iterm2.BroadcastDomain{
		{SessionIds: []string{"s1", "s2"}},
	})
	require.NoError(t, err)
}

func TestReorderTabs_Success(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	err := ReorderTabs(ctx, mc, []*iterm2.ReorderTabsRequest_Assignment{
		{WindowId: proto.String("w1"), TabIds: []string{"t1", "t2"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "w1", mc.req.GetReorderTabsRequest().GetAssignments()[0].GetWindowId())
}
