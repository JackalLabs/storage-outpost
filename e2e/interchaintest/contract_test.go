package main

import (
	"context"
	"fmt"
	"testing"

	mysuite "github.com/JackalLabs/cw-ica-controller/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/cw-ica-controller/e2e/interchaintest/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"
)

type ContractTestSuite struct {
	mysuite.TestSuite

	Contract   *types.Contract
	IcaAddress string
}

// SetupContractAndChannel starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *ContractTestSuite) SetupContractTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	var err error
	// Upload and Instantiate the contract on wasmd:
	s.Contract, err = types.StoreAndInstantiateNewContract(ctx, s.ChainA, s.UserA.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	version := fmt.Sprintf(
		`{"version":"%s",`+
			`"controller_connection_id":"%s",`+
			`"host_connection_id":"%s",`+
			`"address":"",`+ // NOTE: why is the address initially empty?
			`"encoding":"%s",`+
			`"tx_type":"%s"}`,
		icatypes.Version, s.ChainAConnID, s.ChainBConnID,
		encoding, icatypes.TxTypeSDKMultiMsg,
	)
	err = s.Relayer.CreateChannel(ctx, s.ExecRep, s.PathName, ibc.CreateChannelOptions{
		SourcePortName: s.Contract.Port(),
		DestPortName:   icatypes.HostPortID,
		Order:          ibc.Ordered,
		// cannot use an empty version here, see README
		Version: version,
	})
	s.Require().NoError(err)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := s.Contract.QueryContractState(ctx)
	s.Require().NoError(err)
	s.IcaAddress = contractState.IcaInfo.IcaAddress
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}
