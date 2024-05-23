package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ivxivx/demo-blockchain/blockchain/transaction"
	"github.com/ivxivx/demo-blockchain/blockchain/transaction/chain/evm"
	"github.com/ivxivx/demo-blockchain/blockchain/transaction/provider/local"
	"github.com/ivxivx/demo-blockchain/domain"
	"github.com/ivxivx/demo-blockchain/repo"
)

const (
	providerIDLocal = "Local"

	walletIDEvmLocalTestnet = "018ee4c9-5161-7fa2-b280-20573311aab4"

	paramEvmLocalTestnetURL = "evm_local_testnet_url"
)

type Currency struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Address struct {
	ID          uuid.UUID `json:"id"`
	Address     string    `json:"address"`
	NetworkCode string    `json:"network_code"`
	WalletID    uuid.UUID `json:"wallet_id"`
	ProviderID  string    `json:"provider_id"`
}

type Provider struct {
	ID     string            `json:"id"`
	Params map[string]string `json:"params"`
}

type Wallet struct {
	ID         uuid.UUID `json:"id"`
	ProviderID string    `json:"provider_id"`
	PrivateKey string    `json:"private_key"`
}

type Network struct {
	Currencies []*Currency `json:"currencies"`
	Addresses  []*Address  `json:"addresses"`
}

type Transferor struct {
	ProviderID  string                 `json:"provider_id"`
	Description string                 `json:"description"`
	Delegate    transaction.Transferor `json:"-"`
}

type Transaction struct {
	ID                    string `json:"id"`
	URL                   string `json:"url"`
	TransferorDescription string `json:"creator_description"`
}

type DemoConfig struct {
	Addresses []*domain.Address `json:"addresses"`
	Providers []*Provider       `json:"providers"`
	Wallets   []*Wallet         `json:"wallets"`
}

type DemoContext struct {
	txmgr       *transaction.Manager
	walletRepo  domain.WalletRepo
	addressRepo domain.AddressRepo
}

type ProviderNotFoundError struct {
	ProviderID string
}

func (e ProviderNotFoundError) Error() string {
	return "provider not found for ID " + e.ProviderID
}

//go:embed index.html
var indexHtml embed.FS

//go:embed config.json
var configFile embed.FS

var testEthClient *evm.Client

func main() {
	ctx := context.Background()

	contentFS, err := fs.Sub(indexHtml, ".")
	if err != nil {
		log.Fatal(err)
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	demoContext, err := newTransferors(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(contentFS)))

	http.HandleFunc("GET /demo/networks", getNetwork(config))
	http.HandleFunc("POST /demo/payouts", createPayout(config, demoContext))
	http.HandleFunc("GET /demo/transactions", getTransaction)

	const readerHeaderTimeout = 5 * time.Second

	server := &http.Server{
		Addr:              ":9111",
		ReadHeaderTimeout: readerHeaderTimeout,
	}

	slog.Log(ctx, slog.LevelInfo, "Listening on :9111...")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() (*DemoConfig, error) {
	content, err := configFile.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config DemoConfig

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func getNetwork(config *DemoConfig) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		networkCode := req.FormValue("network")

		networkCurrencies, err := domain.GetNetworkCurrencies(networkCode)
		if err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to retrieve currencies:", err)
			http.Error(resp, "failed to retrieve currencies", http.StatusInternalServerError)

			return
		}

		currencies := make([]*Currency, len(networkCurrencies))

		for index, networkCurrency := range networkCurrencies {
			currencies[index] = &Currency{
				ID:    networkCurrency.ID,
				Label: networkCurrency.Currency.Code,
			}
		}

		addrs := getAddresses(config, networkCode)
		if len(addrs) == 0 {
			slog.Log(req.Context(), slog.LevelError, "failed to get provider address:",
				slog.Any("NetworkCode", networkCode))
			http.Error(resp, fmt.Sprintf("failed to get provider address: %s", err), http.StatusInternalServerError)

			return
		}

		addresses := make([]*Address, len(addrs))

		for index, addr := range addrs {
			wallet, errW := getWallet(config, addr.WalletID)
			if errW != nil {
				slog.Log(req.Context(), slog.LevelError, "failed to retrieve wallet:", errW)
				http.Error(resp, "failed to retrieve wallet", http.StatusInternalServerError)

				return
			}

			addresses[index] = &Address{
				ID:          addr.ID,
				Address:     addr.Address,
				NetworkCode: addr.NetworkCode,
				WalletID:    addr.WalletID,
				ProviderID:  wallet.ProviderID,
			}
		}

		networkRes := Network{
			Currencies: currencies,
			Addresses:  addresses,
		}

		res, err := json.MarshalIndent(networkRes, "", "  ")
		if err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to marshall response:", err)
			http.Error(resp, "failed to marshall response", http.StatusInternalServerError)

			return
		}

		resp.Header().Set("Content-Type", "application/json")

		if _, err := resp.Write(res); err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to writeg response:", err)
			http.Error(resp, "failed to writeg response", http.StatusInternalServerError)

			return
		}
	}
}

func getTransaction(resp http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	transactionID := req.FormValue("tid")

	txn, _, err := testEthClient.Delegate.TransactionByHash(ctx, common.HexToHash(transactionID))
	if err != nil {
		slog.Log(req.Context(), slog.LevelError, "failed to retrieve transaction:", err)
		http.Error(resp, "failed to retrieve transaction", http.StatusInternalServerError)

		return
	}

	txRes := map[string]interface{}{
		"id":        transactionID,
		"chain_id":  txn.ChainId(),
		"time":      txn.Time(),
		"gas":       txn.Gas(),
		"gas_price": txn.GasPrice(),
		"nonce":     txn.Nonce(),
	}

	res, err := json.MarshalIndent(txRes, "", "  ")
	if err != nil {
		slog.Log(req.Context(), slog.LevelError, "failed to marshall response:", err)
		http.Error(resp, "failed to marshall response", http.StatusInternalServerError)

		return
	}

	resp.Header().Set("Content-Type", "application/json")

	if _, err := resp.Write(res); err != nil {
		slog.Log(req.Context(), slog.LevelError, "failed to writeg response:", err)
		http.Error(resp, "failed to writeg response", http.StatusInternalServerError)

		return
	}
}

func createPayout(
	_ *DemoConfig,
	demoContext *DemoContext,
) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := context.Background()

		if err := req.ParseForm(); err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to parse form:", err)
			http.Error(resp, fmt.Sprintf("failed to parse form: %s", err), http.StatusInternalServerError)

			return
		}

		fromAddress := req.FormValue("from")
		toAddress := req.FormValue("to")
		amount := req.FormValue("amount")
		currencyCode := req.FormValue("currency")

		amountDecimal, err := decimal.NewFromString(amount)
		if err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to parse amount:", err)
			http.Error(resp, fmt.Sprintf("failed to parse amount: %s", err), http.StatusInternalServerError)

			return
		}

		networkCurrency, err := domain.NewNetworkCurrency(currencyCode)
		if err != nil {

			return
		}

		param := &transaction.TransferRequest{
			SourceAddress:      fromAddress,
			DestinationAddress: toAddress,
			Amount:             amountDecimal,
			NetworkCurrencyID:  networkCurrency.ID,
		}

		payload, err := demoContext.txmgr.Transfer(ctx, param)
		if err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to create transfer:", err)
			http.Error(resp, fmt.Sprintf("failed to create transfer: %s", err), http.StatusInternalServerError)

			return
		}

		var txExplorerURLPrefix string

		switch networkCurrency.Network.Code {
		case domain.TestEth:
			txExplorerURLPrefix = "http://localhost:9111/demo/transactions?tid="
		}

		txExplorerURL := txExplorerURLPrefix + payload.ID

		txRes := Transaction{
			ID:                    payload.ID,
			URL:                   txExplorerURL,
			TransferorDescription: payload.ProviderID,
		}

		res, err := json.MarshalIndent(txRes, "", "  ")
		if err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to marshall response:", err)

			http.Error(resp, "failed to marshall response", http.StatusInternalServerError)

			return
		}

		resp.Header().Set("Content-Type", "application/json")

		if _, err := resp.Write(res); err != nil {
			slog.Log(req.Context(), slog.LevelError, "failed to writeg response:", err)
			http.Error(resp, "failed to writeg response", http.StatusInternalServerError)

			return
		}
	}
}

func getAddresses(config *DemoConfig, network string) []*domain.Address {
	results := make([]*domain.Address, 0)

	for _, address := range config.Addresses {
		if address.NetworkCode == network {
			addr := address
			results = append(results, addr)
		}
	}

	return results
}

func getProvider(config *DemoConfig, providerID string) (*Provider, error) {
	for _, provider := range config.Providers {
		if provider.ID == providerID {
			return provider, nil
		}
	}

	return nil, &ProviderNotFoundError{ProviderID: providerID}
}

func getWallet(config *DemoConfig, walletID uuid.UUID) (*Wallet, error) {
	for _, wallet := range config.Wallets {
		if wallet.ID == walletID {
			return wallet, nil
		}
	}

	return nil, &domain.WalletNotFoundError{WalletID: walletID}
}

func newLocalEvmTransferor(
	ctx context.Context,
	config *DemoConfig,
	walletID string,
	nodeURL string,
) (*evm.Client, transaction.Transferor, transaction.Builder, error) {
	wallet, err := getWallet(config, uuid.MustParse(walletID))
	if err != nil {
		return nil, nil, nil, err
	}

	provider, err := getProvider(config, wallet.ProviderID)
	if err != nil {
		return nil, nil, nil, err
	}

	url := provider.Params[nodeURL]

	client, err := evm.NewClient(ctx, url)
	if err != nil {

		return nil, nil, nil, err
	}

	builder := evm.NewTransactionBuilder(client)
	signer := evm.NewPrvKeyTransactionSigner(wallet.PrivateKey)
	broadcaster := evm.NewTransactionBroadcaster(client)
	transferor := transaction.NewGenericTransferor(builder, signer, broadcaster)

	return client, transferor, builder, nil
}

func newTransferors(
	ctx context.Context,
	config *DemoConfig,
) (*DemoContext, error) {
	walletRepo := repo.NewWalletRepo()

	for _, wallet := range config.Wallets {
		walletPayload := &domain.CreateWalletPayload{
			ID:         wallet.ID,
			ProviderID: wallet.ProviderID,
		}

		_, err := walletRepo.CreateWallet(ctx, walletPayload)
		if err != nil {

			return nil, err
		}
	}

	addressRepo := repo.NewAddressRepo()

	for _, addr := range config.Addresses {
		address := &domain.CreateAddressPayload{
			Address:     addr.Address,
			NetworkCode: addr.NetworkCode,
			WalletID:    addr.WalletID,
		}

		_, err := addressRepo.CreateAddress(ctx, address)
		if err != nil {

			return nil, err
		}
	}

	testEthC, testEthTransferor, _, err := newLocalEvmTransferor(ctx, config, walletIDEvmLocalTestnet, paramEvmLocalTestnetURL)
	if err != nil {
		return nil, err
	}

	testEthClient = testEthC

	localTransferors := map[string]transaction.Transferor{
		domain.TestEth: testEthTransferor,
	}

	transferorMap := make(map[string]transaction.Transferor)
	transferorMap[providerIDLocal] = local.NewTransactionTranferor(localTransferors)

	txmgr := transaction.NewManager(addressRepo, walletRepo, transferorMap)

	return &DemoContext{
		txmgr,
		walletRepo,
		addressRepo,
	}, nil
}
