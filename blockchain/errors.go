package blockchain

import (
	"fmt"
)

type NetworkNotSupportedError struct {
	NetworkCode string
}

func (e NetworkNotSupportedError) Error() string {
	return fmt.Sprintf("network %s not supported", e.NetworkCode)
}

type TransactionError struct {
	Message string
}

func (e TransactionError) Error() string {
	return e.Message
}
