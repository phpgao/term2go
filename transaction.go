package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Transaction is a context-based wrapper around iTerm2's transaction API.
//
// A transaction freezes the terminal state, guaranteeing that nothing changes
// between a sequence of API calls. Use it when you need to read screen
// contents and act on them atomically (e.g. GetLineInfo + GetContents).
//
//	ctx := context.Background()
//	tx := term2go.NewTransaction(conn)
//	if err := tx.Begin(ctx); err != nil {
//	    return err
//	}
//	defer tx.End(ctx)
//
//	info, _ := session.GetLineInfo(ctx)
//	lines, _ := session.GetContents(ctx, int32(info.Overflow), 10)
type Transaction struct {
	caller  Caller
	started bool
}

// NewTransaction creates a Transaction bound to the given Caller.
func NewTransaction(caller Caller) *Transaction {
	return &Transaction{caller: caller}
}

// Begin starts a transaction. The app's main loop will not advance while the
// transaction is active. Call End when done.
func (tx *Transaction) Begin(ctx context.Context) error {
	if tx.started {
		return fmt.Errorf("transaction already started")
	}
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_TransactionRequest{
		TransactionRequest: &iterm2.TransactionRequest{
			Begin: proto.Bool(true),
		},
	}
	resp, err := tx.caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	if err := checkError(resp); err != nil {
		return err
	}
	tx.started = true
	return nil
}

// End commits the transaction, allowing the app to resume.
func (tx *Transaction) End(ctx context.Context) error {
	if !tx.started {
		return fmt.Errorf("transaction not started")
	}
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_TransactionRequest{
		TransactionRequest: &iterm2.TransactionRequest{
			Begin: proto.Bool(false),
		},
	}
	resp, err := tx.caller.Call(ctx, msg)
	if err != nil {
		return err
	}
	tx.started = false
	return checkError(resp)
}
