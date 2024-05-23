package evm

import (
	"context"
	"fmt"

	"github.com/ivxivx/demo-blockchain/blockchain/transaction"
)

type TransactionBroadcaster struct {
	client *Client
}

var _ transaction.Broadcaster = (*TransactionBroadcaster)(nil)

func NewTransactionBroadcaster(client *Client) *TransactionBroadcaster {
	return &TransactionBroadcaster{
		client: client,
	}
}

func (broadcaster *TransactionBroadcaster) Broadcast(
	ctx context.Context,
	payload *transaction.TransferPayload,
) error {
	txn, err := Unmarshal(payload.Signed)
	if err != nil {
		return err
	}

	err = broadcaster.client.Delegate.SendTransaction(ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction (%s): %w", payload.ID, err)
	}

	return nil
}
