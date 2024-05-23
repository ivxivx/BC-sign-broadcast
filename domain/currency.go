package domain

import "fmt"

const (
	TestEth string = "TestEth"

	ETH  string = "ETH"

	TestETH  string = "TEST_ETH"
)

type Currency struct {
	Code string
}

type Network struct {
	Code        string
	NativeToken string
}

type NetworkCurrency struct {
	ID       string
	Network  Network
	Currency Currency
	Scale    int
	Address  string
}

func GetNetworkCurrencies(networkCode string) ([]*NetworkCurrency, error) {
	switch networkCode {
	case "TestEth":
		return []*NetworkCurrency{
			{
				ID:      TestETH,
				Scale:   18,
				Address: "",
				Network: Network{
					Code:        TestEth,
					NativeToken: TestETH,
				},
				Currency: Currency{
					Code: ETH,
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("invalid network: %s", networkCode)
}

func NewNetworkCurrency(networkCurrencyID string) (*NetworkCurrency, error) {
	switch networkCurrencyID {
	case "TEST_ETH":
		return &NetworkCurrency{
			ID:      TestETH,
			Scale:   18,
			Address: "",
			Network: Network{
				Code:        TestEth,
				NativeToken: TestETH,
			},
			Currency: Currency{
				Code: ETH,
			},
		}, nil
	}
	return nil, fmt.Errorf("invalid currency: %s", networkCurrencyID)
}
