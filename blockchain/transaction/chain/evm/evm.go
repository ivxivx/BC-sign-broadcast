package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Chain struct{}

type Client struct {
	Delegate *ethclient.Client
}

func NewClient(ctx context.Context, url string) (*Client, error) {
	clnt, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create client (%s): %w", url, err)
	}

	client := &Client{
		Delegate: clnt,
	}

	return client, nil
}

func (client *Client) Close() {
	client.Delegate.Close()
}

func Marshal(txn *types.Transaction) ([]byte, error) {
	bytes, err := txn.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	return bytes, nil
}

func Unmarshal(bytes []byte) (*types.Transaction, error) {
	txn := &types.Transaction{}

	err := txn.UnmarshalBinary(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return txn, nil
}
