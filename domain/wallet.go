package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type WalletNotFoundError struct {
	WalletID uuid.UUID
}

func (e WalletNotFoundError) Error() string {
	return fmt.Sprintf("wallet %s not found", e.WalletID)
}

type Wallet struct {
	ID               uuid.UUID `json:"id"`
	ProviderID       string    `json:"provider_id"`
}

type CreateWalletPayload struct {
	ID               uuid.UUID `json:"id"`
	ProviderID       string    `json:"provider_id"`
}
