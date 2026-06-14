package term2go

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// Splitter
// ============================================================================

// SplitChild holds either a Session or a nested Splitter, but never both.
// Use IsSession() or IsSplitter() to determine which field is set.
type SplitChild struct {
	Session  *Session
	Splitter *Splitter
}

// IsSession reports whether this child is a Session.
func (c *SplitChild) IsSession() bool { return c.Session != nil }

// IsSplitter reports whether this child is a Splitter.
func (c *SplitChild) IsSplitter() bool { return c.Splitter != nil }

// SessionOrNil returns the Session if this is a leaf, or nil otherwise.
// This avoids allocating a Splitter when the child is actually a Session.
func (c *SplitChild) SessionOrNil() *Session {
	return c.Session
}

// SplitterOrNil returns the Splitter if this is a node, or nil otherwise.
// This avoids allocating a Session when the child is actually a Splitter.
func (c *SplitChild) SplitterOrNil() *Splitter {
	return c.Splitter
}

// Splitter represents a split pane tree node. It is either a leaf (containing
// a Session) or an interior node with a split direction and child splitters
// or sessions.
type Splitter struct {
	Vertical bool
	Children []SplitChild
}

// ToProto converts the Splitter tree back to a SplitTreeNode for RPC use.
func (s *Splitter) ToProto() *iterm2.SplitTreeNode {
	if s == nil {
		return nil
	}
	node := &iterm2.SplitTreeNode{Vertical: &s.Vertical}
	for _, c := range s.Children {
		link := &iterm2.SplitTreeNode_SplitTreeLink{}
		switch {
		case c.Session != nil:
			link.Child = &iterm2.SplitTreeNode_SplitTreeLink_Session{
				Session: &iterm2.SessionSummary{
					UniqueIdentifier: proto.String(c.Session.ID),
				},
			}
		case c.Splitter != nil:
			link.Child = &iterm2.SplitTreeNode_SplitTreeLink_Node{
				Node: c.Splitter.ToProto(),
			}
		}
		node.Links = append(node.Links, link)
	}
	return node
}

// SplitterFromProto recursively builds a Splitter tree from a proto
// SplitTreeNode. Each link in the node is either a leaf (Session) or a nested
// sub-splitter.
func SplitterFromProto(node *iterm2.SplitTreeNode, conn Caller) *Splitter {
	if node == nil {
		return nil
	}
	s := &Splitter{Vertical: node.GetVertical()}
	for _, link := range node.GetLinks() {
		if sess := link.GetSession(); sess != nil {
			s.Children = append(s.Children, SplitChild{Session: &Session{
				caller: conn,
				ID:     sess.GetUniqueIdentifier(),
			}})
		} else if subNode := link.GetNode(); subNode != nil {
			if child := SplitterFromProto(subNode, conn); child != nil {
				s.Children = append(s.Children, SplitChild{Splitter: child})
			}
		}
	}
	return s
}

// Sessions returns all Session leaf nodes in this splitter tree, including
// those nested in sub-splitters.
func (s *Splitter) Sessions() []*Session {
	var result []*Session
	for _, child := range s.Children {
		if child.Session != nil {
			result = append(result, child.Session)
		} else if child.Splitter != nil {
			result = append(result, child.Splitter.Sessions()...)
		}
	}
	return result
}

// jsonDecodeForVariable decodes a JSON-encoded variable value from iTerm2.
// Returns the decoded string for JSON strings, string representation for
// numbers/bools, or empty string for JSON null / invalid JSON.
func jsonDecodeForVariable(raw string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw // not valid JSON — return as-is (legacy values)
	}
	switch val := v.(type) {
	case string:
		return val
	case nil:
		return ""
	case float64:
		return raw // preserve exact representation (e.g., "42" vs "42.0")
	case bool:
		return raw // preserve "true"/"false"
	default:
		return raw
	}
}

