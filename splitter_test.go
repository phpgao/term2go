package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestSplitChild_IsSession tests SplitChild.IsSession method for various cases.
func TestSplitChild_IsSession(t *testing.T) {
	t.Run("with Session", func(t *testing.T) {
		caller := &mockCaller{}
		sc := SplitChild{Session: &Session{caller: caller, ID: "s1"}}
		assert.True(t, sc.IsSession())
	})
	t.Run("with Splitter", func(t *testing.T) {
		sc := SplitChild{Splitter: &Splitter{}}
		assert.False(t, sc.IsSession())
	})
	t.Run("with neither", func(t *testing.T) {
		var sc SplitChild
		assert.False(t, sc.IsSession())
	})
}

// TestSplitChild_IsSplitter tests SplitChild.IsSplitter method for various cases.
func TestSplitChild_IsSplitter(t *testing.T) {
	t.Run("with Splitter", func(t *testing.T) {
		sc := SplitChild{Splitter: &Splitter{}}
		assert.True(t, sc.IsSplitter())
	})
	t.Run("with Session", func(t *testing.T) {
		caller := &mockCaller{}
		sc := SplitChild{Session: &Session{caller: caller, ID: "s1"}}
		assert.False(t, sc.IsSplitter())
	})
	t.Run("with neither", func(t *testing.T) {
		var sc SplitChild
		assert.False(t, sc.IsSplitter())
	})
}

// TestSplitChild_SessionOrNil tests SplitChild.SessionOrNil method for various cases.
func TestSplitChild_SessionOrNil(t *testing.T) {
	t.Run("returns Session when set", func(t *testing.T) {
		caller := &mockCaller{}
		sc := SplitChild{Session: &Session{caller: caller, ID: "s1"}}
		sess := sc.SessionOrNil()
		require.NotNil(t, sess)
		assert.Equal(t, "s1", sess.ID)
	})
	t.Run("returns nil when Splitter set", func(t *testing.T) {
		sc := SplitChild{Splitter: &Splitter{}}
		assert.Nil(t, sc.SessionOrNil())
	})
	t.Run("returns nil when neither set", func(t *testing.T) {
		var sc SplitChild
		assert.Nil(t, sc.SessionOrNil())
	})
}

// TestSplitChild_SplitterOrNil tests SplitChild.SplitterOrNil method for various cases.
func TestSplitChild_SplitterOrNil(t *testing.T) {
	t.Run("returns Splitter when set", func(t *testing.T) {
		sc := SplitChild{Splitter: &Splitter{Vertical: true}}
		s := sc.SplitterOrNil()
		require.NotNil(t, s)
		assert.True(t, s.Vertical)
	})
	t.Run("returns nil when Session set", func(t *testing.T) {
		caller := &mockCaller{}
		sc := SplitChild{Session: &Session{caller: caller, ID: "s1"}}
		assert.Nil(t, sc.SplitterOrNil())
	})
	t.Run("returns nil when neither set", func(t *testing.T) {
		var sc SplitChild
		assert.Nil(t, sc.SplitterOrNil())
	})
}

// TestSplitter_ToProto tests Splitter.ToProto method for various cases.
func TestSplitter_ToProto(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, (*Splitter)(nil).ToProto())
	})
	t.Run("leaf with Session", func(t *testing.T) {
		caller := &mockCaller{}
		s := &Splitter{
			Vertical: false,
			Children: []SplitChild{
				{Session: &Session{caller: caller, ID: "sess-1"}},
			},
		}
		node := s.ToProto()
		require.NotNil(t, node)
		assert.False(t, node.GetVertical())
		require.Len(t, node.GetLinks(), 1)
		assert.Equal(t, "sess-1", node.GetLinks()[0].GetSession().GetUniqueIdentifier())
	})
	t.Run("with nested Splitter", func(t *testing.T) {
		caller := &mockCaller{}
		nested := &Splitter{
			Vertical: true,
			Children: []SplitChild{
				{Session: &Session{caller: caller, ID: "sess-2"}},
			},
		}
		s := &Splitter{
			Vertical: false,
			Children: []SplitChild{
				{Session: &Session{caller: caller, ID: "sess-1"}},
				{Splitter: nested},
			},
		}
		node := s.ToProto()
		require.NotNil(t, node)
		require.Len(t, node.GetLinks(), 2)
		// First link is a session
		assert.Equal(t, "sess-1", node.GetLinks()[0].GetSession().GetUniqueIdentifier())
		// Second link is a nested node
		subNode := node.GetLinks()[1].GetNode()
		require.NotNil(t, subNode)
		assert.True(t, subNode.GetVertical())
	})
	t.Run("empty Splitter", func(t *testing.T) {
		s := &Splitter{Vertical: true, Children: nil}
		node := s.ToProto()
		require.NotNil(t, node)
		assert.True(t, node.GetVertical())
		assert.Empty(t, node.GetLinks())
	})
}

// TestSplitterFromProto tests SplitterFromProto function for various cases.
func TestSplitterFromProto(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, SplitterFromProto(nil, nil))
	})
	t.Run("leaf with Session", func(t *testing.T) {
		caller := &mockCaller{}
		node := &iterm2.SplitTreeNode{
			Vertical: proto.Bool(false),
			Links: []*iterm2.SplitTreeNode_SplitTreeLink{
				{
					Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{
						Session: &iterm2.SessionSummary{
							UniqueIdentifier: proto.String("sess-1"),
						},
					},
				},
			},
		}
		s := SplitterFromProto(node, caller)
		require.NotNil(t, s)
		assert.False(t, s.Vertical)
		require.Len(t, s.Children, 1)
		assert.True(t, s.Children[0].IsSession())
		assert.Equal(t, "sess-1", s.Children[0].Session.ID)
	})
	t.Run("with nested Splitter", func(t *testing.T) {
		caller := &mockCaller{}
		subNode := &iterm2.SplitTreeNode{
			Vertical: proto.Bool(true),
			Links: []*iterm2.SplitTreeNode_SplitTreeLink{
				{
					Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{
						Session: &iterm2.SessionSummary{
							UniqueIdentifier: proto.String("sess-2"),
						},
					},
				},
			},
		}
		node := &iterm2.SplitTreeNode{
			Vertical: proto.Bool(false),
			Links: []*iterm2.SplitTreeNode_SplitTreeLink{
				{
					Child: &iterm2.SplitTreeNode_SplitTreeLink_Session{
						Session: &iterm2.SessionSummary{
							UniqueIdentifier: proto.String("sess-1"),
						},
					},
				},
				{
					Child: &iterm2.SplitTreeNode_SplitTreeLink_Node{Node: subNode},
				},
			},
		}
		s := SplitterFromProto(node, caller)
		require.NotNil(t, s)
		require.Len(t, s.Children, 2)
		// First child is a session
		assert.True(t, s.Children[0].IsSession())
		assert.Equal(t, "sess-1", s.Children[0].Session.ID)
		// Second child is a nested splitter
		assert.True(t, s.Children[1].IsSplitter())
		nestedSplitter := s.Children[1].Splitter
		require.NotNil(t, nestedSplitter)
		assert.True(t, nestedSplitter.Vertical)
		require.Len(t, nestedSplitter.Children, 1)
	})
	t.Run("empty Links", func(t *testing.T) {
		caller := &mockCaller{}
		node := &iterm2.SplitTreeNode{
			Vertical: proto.Bool(true),
			Links:    nil,
		}
		s := SplitterFromProto(node, caller)
		require.NotNil(t, s)
		assert.True(t, s.Vertical)
		assert.Empty(t, s.Children)
	})
}
