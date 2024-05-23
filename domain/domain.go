package domain

import (
	"context"

	"github.com/google/uuid"
)

type WalletRepo interface {
	CreateWallet(context.Context, *CreateWalletPayload) (*Wallet, error)
	GetWallet(context.Context, uuid.UUID) (*Wallet, error)
}

type AddressRepo interface {
	CreateAddress(context.Context, *CreateAddressPayload) (*Address, error)
	GetAddressByValue(context.Context, string, string) (*Address, error)
	GetAddressesByNetwork(context.Context, string) ([]*Address, error)
}
