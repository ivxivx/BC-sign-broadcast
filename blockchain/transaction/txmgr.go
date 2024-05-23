package transaction

import (
	"context"
	"fmt"

	"github.com/ivxivx/demo-blockchain/domain"
)

type TransferorNotFoundError struct {
	ProviderID string
}

func (e TransferorNotFoundError) Error() string {
	return "transferor not found for provider " + e.ProviderID
}

type Manager struct {
	addressRepo domain.AddressRepo
	walletRepo  domain.WalletRepo

	transferorMap map[string]Transferor
}

func NewManager(
	addressRepo domain.AddressRepo,
	walletRepo domain.WalletRepo,
	transferorMap map[string]Transferor,
) *Manager {
	return &Manager{
		addressRepo:   addressRepo,
		walletRepo:    walletRepo,
		transferorMap: transferorMap,
	}
}

func (txmgr *Manager) Transfer(ctx context.Context, param *TransferRequest) (*TransferPayload, error) {
	networkCurrency, err := domain.NewNetworkCurrency(param.NetworkCurrencyID)
	if err != nil {

		return nil, err
	}

	if param.SourceAddress == "" {
		return nil, fmt.Errorf("source address is not provided")
	}

	address, err := txmgr.addressRepo.GetAddressByValue(ctx, param.SourceAddress, networkCurrency.Network.Code)
	if err != nil {

		return nil, err
	}

	wallet, err := txmgr.walletRepo.GetWallet(ctx, address.WalletID)
	if err != nil {

		return nil, err
	}

	transferor, ok := txmgr.transferorMap[wallet.ProviderID]
	if !ok {
		return nil, TransferorNotFoundError{ProviderID: wallet.ProviderID}
	}

	payload, err := transferor.Transfer(ctx, param)
	if err != nil {

		return nil, err
	}

	payload.ProviderID = wallet.ProviderID

	return payload, nil
}
