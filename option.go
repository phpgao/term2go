package term2go

import "time"

// Option configures a Connection.
type Option func(*Connection)

// WithCallTimeout sets the deadline for each Call operation.
// Zero means no timeout (use with caution — a missing response blocks forever).
// Default: 30s.
func WithCallTimeout(d time.Duration) Option {
	return func(c *Connection) {
		c.callTimeout = d
	}
}

// WithReadTimeout sets the read deadline for each WebSocket read.
// This controls how long dispatchLoop waits for the next message before timing out.
// Default: 60s.
func WithReadTimeout(d time.Duration) Option {
	return func(c *Connection) {
		c.readTimeout = d
	}
}

// WithHandshakeTimeout sets the WebSocket handshake timeout for dialing.
// Default: 45s.
func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *Connection) {
		c.handshakeTimeout = d
	}
}
