package term2go

import (
	"context"
	"fmt"

	iterm2 "github.com/phpgao/term2go/proto"
)

// BroadcastDomain describes a set of sessions that share keyboard input.
// When the user types in one session in the domain, all sessions in the same
// domain receive the input.
//
// Broadcast domains are disjoint — a session may belong to at most one domain.
type BroadcastDomain struct {
	caller   Caller
	sessions []string // session IDs
}

// BroadcastDomainFromProto creates a BroadcastDomain from its proto representation.
func BroadcastDomainFromProto(proto *iterm2.BroadcastDomain) *BroadcastDomain {
	return &BroadcastDomain{
		sessions: proto.GetSessionIds(),
	}
}

// SessionIDs returns the session IDs in this broadcast domain.
func (d *BroadcastDomain) SessionIDs() []string {
	return d.sessions
}

// AddSession adds a session to the broadcast domain.
func (d *BroadcastDomain) AddSession(sessionID string) {
	d.sessions = append(d.sessions, sessionID)
}

// RemoveSession removes a session from the broadcast domain.
func (d *BroadcastDomain) RemoveSession(sessionID string) {
	for i, sid := range d.sessions {
		if sid == sessionID {
			d.sessions = append(d.sessions[:i], d.sessions[i+1:]...)
			return
		}
	}
}

// Contains reports whether the domain contains the given session.
func (d *BroadcastDomain) Contains(sessionID string) bool {
	for _, sid := range d.sessions {
		if sid == sessionID {
			return true
		}
	}
	return false
}

// ToProto returns the protobuf representation for RPC use.
func (d *BroadcastDomain) ToProto() *iterm2.BroadcastDomain {
	return &iterm2.BroadcastDomain{
		SessionIds: d.sessions,
	}
}

// BroadcastDomains (collection)

// BroadcastDomains is a collection of broadcast domains.
type BroadcastDomains []*BroadcastDomain

// FromProto fills this collection from a proto response.
func (bds *BroadcastDomains) FromProto(proto []*iterm2.BroadcastDomain) {
	*bds = make(BroadcastDomains, 0, len(proto))
	for _, p := range proto {
		*bds = append(*bds, BroadcastDomainFromProto(p))
	}
}

// ToProto returns the protobuf representation for RPC use.
func (bds BroadcastDomains) ToProto() []*iterm2.BroadcastDomain {
	result := make([]*iterm2.BroadcastDomain, 0, len(bds))
	for _, d := range bds {
		result = append(result, d.ToProto())
	}
	return result
}

// FindForSession returns the broadcast domain that contains the given session,
// or nil if the session is not in any broadcast domain.
func (bds BroadcastDomains) FindForSession(sessionID string) *BroadcastDomain {
	for _, d := range bds {
		if d.Contains(sessionID) {
			return d
		}
	}
	return nil
}

// Set broadcasts domains to iTerm2.
func (bds BroadcastDomains) Set(ctx context.Context, caller Caller) error {
	return SetBroadcastDomains(ctx, caller, bds.ToProto())
}

// RefreshBroadcastDomains fetches the current broadcast domains from iTerm2
// and updates App.BroadcastDomains.
func (a *App) RefreshBroadcastDomains(ctx context.Context) error {
	resp, err := GetBroadcastDomains(ctx, a.caller)
	if err != nil {
		return fmt.Errorf("get broadcast domains: %w", err)
	}
	var bds BroadcastDomains
	bds.FromProto(resp.GetBroadcastDomains())
	a.BroadcastDomains = bds
	return nil
}
