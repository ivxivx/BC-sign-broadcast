package repo

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/ivxivx/demo-blockchain/domain"
)

var _ domain.AddressRepo = (*AddressRepo)(nil)

type AddressRepo struct {
	storage sync.Map
}

func NewAddressRepo() *AddressRepo {
	return &AddressRepo{
		storage: sync.Map{},
	}
}

func (repo *AddressRepo) CreateAddress(_ context.Context, cdp *domain.CreateAddressPayload) (*domain.Address, error) {
	address := &domain.Address{
		ID:          uuid.Must(uuid.NewV7()),
		Address:     cdp.Address,
		NetworkCode: cdp.NetworkCode,
		WalletID:    cdp.WalletID,
	}

	repo.storage.Store(address.ID, address)

	return address, nil
}

func (repo *AddressRepo) GetAddressByValue(
	_ context.Context,
	address string,
	networkCode string,
) (*domain.Address, error) {
	var addrTyped *domain.Address

	repo.storage.Range(func(_, value interface{}) bool {
		addr, ok := value.(*domain.Address)
		if !ok {
			return false
		}

		if addr.Address == address && addr.NetworkCode == networkCode {
			addrTyped = addr

			return false
		}

		return true
	})

	if addrTyped == nil {
		return nil, domain.AddressNotFoundError{Address: address}
	}

	return addrTyped, nil
}

func (repo *AddressRepo) GetAddressesByNetwork(
	_ context.Context,
	networkCode string,
) ([]*domain.Address, error) {
	var addresses []*domain.Address

	repo.storage.Range(func(_, value interface{}) bool {
		addr, ok := value.(*domain.Address)
		if !ok {
			return false
		}

		if addr.NetworkCode == networkCode {
			addresses = append(addresses, addr)
		}

		return true
	})

	return addresses, nil
}
