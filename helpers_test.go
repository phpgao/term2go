package term2go

import (
	"context"
	"time"
)

// testCtx returns a context with 5 second timeout and its cancel function.
// Caller should use defer cancel().
func testCtx() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
