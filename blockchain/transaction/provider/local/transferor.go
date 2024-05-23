package local

import (
	"context"

	"github.com/ivxivx/demo-blockchain/blockchain"
	"github.com/ivxivx/demo-blockchain/blockchain/transaction"
	"github.com/ivxivx/demo-blockchain/domain"
)

type TransactionTransferor struct {
	delegates map[string]transaction.Transferor
}

var _ transaction.Transferor = (*TransactionTransferor)(nil)

func NewTransactionTranferor(delegates map[string]transaction.Transferor) *TransactionTransferor {
	return &TransactionTransferor{
		delegates: delegates,
	}
}

func (ttf *TransactionTransferor) Transfer(
	ctx context.Context,
	param *transaction.TransferRequest,
) (*transaction.TransferPayload, error) {
	networkCurrency, err := domain.NewNetworkCurrency(param.NetworkCurrencyID)
	if err != nil {

		return nil, err
	}

	ctr := ttf.delegates[networkCurrency.Network.Code]
	if ctr == nil {
		return nil, &blockchain.NetworkNotSupportedError{NetworkCode: networkCurrency.Network.Code}
	}

	payload, err := ctr.Transfer(ctx, param)
	if err != nil {

		return nil, err
	}

	return payload, nil
}
