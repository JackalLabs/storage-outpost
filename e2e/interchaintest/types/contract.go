package types

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type Contract struct {
	Address string
	CodeID  string
	Chain   *cosmos.CosmosChain
}

// NewContract creates a new Contract instance
func NewContract(address string, codeId string, chain *cosmos.CosmosChain) Contract {
	return Contract{
		Address: address,
		CodeID:  codeId,
		Chain:   chain,
	}
}

func (c *Contract) Port() string {
	return "wasm." + c.Address
}

// ExecAnyMsg executes the contract with the given exec message.
func (c *Contract) ExecAnyMsg(ctx context.Context, callerKeyName string, execMsg string, extraExecTxArgs ...string) error {
	_, err := c.Chain.ExecuteContract(ctx, callerKeyName, c.Address, execMsg, extraExecTxArgs...)
	return err
}
