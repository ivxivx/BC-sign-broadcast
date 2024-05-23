package evm

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ivxivx/demo-blockchain/blockchain/transaction"
)

type PrivKeyTransactionSigner struct {
	privateKey string
}

var _ transaction.Signer = (*PrivKeyTransactionSigner)(nil)

func NewPrvKeyTransactionSigner(privateKey string) *PrivKeyTransactionSigner {
	return &PrivKeyTransactionSigner{
		privateKey: privateKey,
	}
}

func (signer *PrivKeyTransactionSigner) Sign(
	_ context.Context,
	payload *transaction.TransferPayload,
) error {
	txn, err := Unmarshal(payload.Raw)
	if err != nil {
		return err
	}

	privateKey := signer.privateKey
	if strings.HasPrefix(privateKey, "0x") || strings.HasPrefix(privateKey, "0X") {
		privateKey = privateKey[2:]
	}

	ecdsaPrivateKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return fmt.Errorf("failed to convert key: %w", err)
	}

	signedTx, err := types.SignTx(txn, types.NewLondonSigner(txn.ChainId()), ecdsaPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	signedTxBytes, err := Marshal(signedTx)
	if err != nil {
		return err
	}

	if payload.ID == "" {
		payload.ID = signedTx.Hash().Hex()
	}

	payload.Signed = signedTxBytes

	return nil
}
