package transaction

import (
	"context"

	"github.com/shopspring/decimal"
)

type TransferRequest struct {
	SourceAddress      string
	DestinationAddress string
	Amount             decimal.Decimal
	NetworkCurrencyID  string
}

type TransferPayload struct {
	Req            *TransferRequest
	SourceWalletID string
	ProviderID     string
	ID             string
	Raw            []byte
	Signed         []byte
}

type Builder interface {
	Build(ctx context.Context, param *TransferRequest) (*TransferPayload, error)
}

type Signer interface {
	Sign(ctx context.Context, payload *TransferPayload) error
}

type Broadcaster interface {
	Broadcast(ctx context.Context, payload *TransferPayload) error
}

type Transferor interface {
	Transfer(ctx context.Context, param *TransferRequest) (*TransferPayload, error)
}
