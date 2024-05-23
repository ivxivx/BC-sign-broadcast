package repo

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/ivxivx/demo-blockchain/domain"
)

var _ domain.WalletRepo = (*WalletRepo)(nil)

type WalletRepo struct {
	storage sync.Map
}

func NewWalletRepo() *WalletRepo {
	return &WalletRepo{
		storage: sync.Map{},
	}
}

func (repo *WalletRepo) CreateWallet(_ context.Context, cwp *domain.CreateWalletPayload) (*domain.Wallet, error) {
	var walletID uuid.UUID

	if cwp.ID != uuid.Nil {
		walletID = cwp.ID
	} else {
		walletID = uuid.Must(uuid.NewV7())
	}

	wallet := &domain.Wallet{
		ID:         walletID,
		ProviderID: cwp.ProviderID,
	}

	repo.storage.Store(wallet.ID, wallet)

	return wallet, nil
}

func (repo *WalletRepo) GetWallet(_ context.Context, walletID uuid.UUID) (*domain.Wallet, error) {
	wallet, okLoad := repo.storage.Load(walletID)
	if !okLoad {
		return nil, domain.WalletNotFoundError{WalletID: walletID}
	}

	walletTyped, ok := wallet.(*domain.Wallet)
	if !ok {
		return nil, domain.WalletNotFoundError{WalletID: walletID}
	}

	return walletTyped, nil
}
