package transaction

import (
	"context"
)

type GenericTranferor struct {
	Builder     Builder
	Signer      Signer
	Broadcaster Broadcaster
}

var _ Transferor = (*GenericTranferor)(nil)

func NewGenericTransferor(
	builder Builder,
	signer Signer,
	broadcaster Broadcaster,
) *GenericTranferor {
	return &GenericTranferor{
		Builder:     builder,
		Signer:      signer,
		Broadcaster: broadcaster,
	}
}

func (creator *GenericTranferor) Transfer(
	ctx context.Context,
	param *TransferRequest,
) (*TransferPayload, error) {
	payload, err := creator.Builder.Build(ctx, param)
	if err != nil {

		return nil, err
	}

	err = creator.Signer.Sign(ctx, payload)
	if err != nil {

		return nil, err
	}

	err = creator.Broadcaster.Broadcast(ctx, payload)

	return payload, err
}
