package main

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	outpostfactory "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types/outpostfactory"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/stretchr/testify/suite"
)

// NOTE: This test should be run before running 'yarn outpost_factory' in outpost-client-demo linked below.
// https://github.com/BiPhan4/outpost-client-demo

// SetupContractAndChannel starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *FactoryTestSuite) SetupFactoryClientTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	logger.InitLogger()

	// Upload the outpost's wasm module on Wasmd
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.OutpostContractCodeId, err = strconv.ParseInt(codeId, 10, 64)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_factory.wasm")
	s.Require().NoError(err)

	instantiateMsg := outpostfactory.InstantiateMsg{StorageOutpostCodeId: int(s.OutpostContractCodeId)}
	// this is the outpost factory
	outpostfactoryContractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, toString(instantiateMsg), false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	logger.LogInfo(fmt.Sprintf("factory address: %s", outpostfactoryContractAddr))

}

func TestWithFactoryClientTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) TestOutpostFactoryClient() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupFactoryClientTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json

	time.Sleep(time.Duration(10) * time.Hour)

}
