package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"

	"github.com/ivxivx/demo-blockchain/blockchain/transaction"
	"github.com/ivxivx/demo-blockchain/domain"
)

const (
	Base10      = 10
	PaddingSize = 32
	EthGasLimit = 21000
)

type TransactionBuilder struct {
	client *Client
}

var _ transaction.Builder = (*TransactionBuilder)(nil)

func NewTransactionBuilder(client *Client) *TransactionBuilder {
	return &TransactionBuilder{
		client: client,
	}
}

func (builder *TransactionBuilder) Build(
	ctx context.Context,
	params *transaction.TransferRequest,
) (*transaction.TransferPayload, error) {
	txData, err := builder.build(ctx, params)
	if err != nil {
		return nil, err
	}

	builtTx := types.NewTx(txData)

	bytes, err := Marshal(builtTx)
	if err != nil {
		return nil, err
	}

	return &transaction.TransferPayload{
		Req: params,
		Raw: bytes,
	}, nil
}

func (builder *TransactionBuilder) build(
	ctx context.Context,
	param *transaction.TransferRequest,
) (*types.DynamicFeeTx, error) {
	fromAddr := common.HexToAddress(param.SourceAddress)
	toAddr := common.HexToAddress(param.DestinationAddress)

	chainID, err := builder.client.Delegate.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chain ID: %w", err)
	}

	nonce, err := builder.client.Delegate.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve nonce for address (%s): %w", param.SourceAddress, err)
	}

	gasPrice, err := builder.client.Delegate.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gas price: %w", err)
	}

	var txToAddr common.Address

	var transferAmount *big.Int

	var gasLimit uint64

	var data []byte

	networkCurrency, err := domain.NewNetworkCurrency(param.NetworkCurrencyID)
	if err != nil {

		return nil, err
	}

	convertedAmount := param.Amount.Mul(decimal.NewFromInt(Base10).
		Pow(decimal.NewFromInt(int64(networkCurrency.Scale)))).BigInt()

	if param.NetworkCurrencyID == networkCurrency.Network.NativeToken {
		txToAddr = toAddr
		transferAmount = convertedAmount

		gasLimit = uint64(EthGasLimit)
	} else {
		txToAddr = common.HexToAddress(networkCurrency.Address)
		transferAmount = big.NewInt(0)

		transferFnSignature := []byte("transfer(address,uint256)")
		hash := sha3.NewLegacyKeccak256()
		hash.Write(transferFnSignature)
		methodID := hash.Sum(nil)[:4]

		paddedAddress := common.LeftPadBytes(toAddr.Bytes(), PaddingSize)
		paddedAmount := common.LeftPadBytes(convertedAmount.Bytes(), PaddingSize)

		data = append(data, methodID...)
		data = append(data, paddedAddress...)
		data = append(data, paddedAmount...)

		estimatedGas, err2 := builder.client.Delegate.EstimateGas(ctx, ethereum.CallMsg{
			From: fromAddr,
			To:   &txToAddr,
			Data: data,
		})

		if err2 != nil {
			return nil, fmt.Errorf(
				"failed to estimate gas for currency(%s), from(%s) and to(%s): %w",
				param.NetworkCurrencyID, param.SourceAddress, txToAddr, err2,
			)
		}

		gasLimit = estimatedGas
	}

	return &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &txToAddr,
		Value:     transferAmount,
		Gas:       gasLimit,
		GasTipCap: gasPrice,
		GasFeeCap: gasPrice,
		Data:      data,
	}, nil
}
