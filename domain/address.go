package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type AddressNotFoundError struct {
	Address     string
	NetworkCode string
}

func (e AddressNotFoundError) Error() string {
	if e.NetworkCode != "" {
		return fmt.Sprintf("address not found for network %s", e.NetworkCode)
	}

	return fmt.Sprintf("address %s not found", e.Address)
}

type Address struct {
	ID          uuid.UUID `json:"id"`
	Address     string    `json:"address"`
	NetworkCode string    `json:"network_code"`
	WalletID    uuid.UUID `json:"wallet_id"`
}

type CreateAddressPayload struct {
	Address     string    `json:"address"`
	NetworkCode string    `json:"network_code"`
	WalletID    uuid.UUID `json:"wallet_id"`
}
