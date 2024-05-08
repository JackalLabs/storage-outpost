package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	mysuite "github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type ContractTestSuite struct {
	mysuite.TestSuite

	Contract   *types.IcaContract
	IcaAddress string
}

// SetupContractAndChannel starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *ContractTestSuite) SetupContractTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	// Upload and Instantiate the contract on wasmd:
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	admin := s.UserA.FormattedAddress()

	// Instantiate the contract with channel:
	instantiateMsg := types.NewInstantiateMsgWithChannelInitOptions(&admin, s.ChainAConnID, s.ChainBConnID, nil, &encoding)

	contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	// Store storage_outpost_v2.wasm
	_, error := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/v2/storage_outpost_v2.wasm")
	s.Require().NoError(error)

	logger.InitLogger()
	fmt.Println("The sender of instantiate is", s.UserA.KeyName())
	logger.LogInfo("The sender of instantiate is", s.UserA.KeyName())

	s.Contract = types.NewIcaContract(types.NewContract(contractAddr, codeId, s.ChainA))

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := s.Contract.QueryContractState(ctx)
	s.Require().NoError(err)

	// NOTE: note sure if Jackal Outpost needs the ownership feature
	// ownershipResponse, err := s.Contract.QueryOwnership(ctx)
	// s.Require().NoError(err)

	s.IcaAddress = contractState.IcaInfo.IcaAddress
	s.Contract.SetIcaAddress(s.IcaAddress)

	// s.Require().Equal(s.UserA.FormattedAddress(), ownershipResponse.Owner)
	// s.Require().Nil(ownershipResponse.PendingOwner)
	// s.Require().Nil(ownershipResponse.PendingExpiry)
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestIcaContractChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json
	wasmd, canined := s.ChainA, s.ChainB

	s.Run("TestChannelHandshakeSuccess", func() {
		// Test if the handshake was successful
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))

		wasmdChannel := wasmdChannels[0]
		s.T().Logf("wasmd channel: %s", toJSONString(wasmdChannel))
		s.Require().Equal(s.Contract.Port(), wasmdChannel.PortID)
		s.Require().Equal(icatypes.HostPortID, wasmdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), wasmdChannel.State)

		// It looks like canined takes some time to open the channel, let's check for its state now and then check again at the end
		// EDIT: we previously tried to query for a TRYOPEN state here, but the handshake is sometimes fast
		// and also sometimes slow. We opted to wait until after the sleep time to query.

		// Check contract's channel state
		contractChannelState, err := s.Contract.QueryChannelState(ctx)
		s.Require().NoError(err)

		s.T().Logf("contract's channel store after handshake: %s", toJSONString(contractChannelState))

		s.Require().Equal(wasmdChannel.State, contractChannelState.ChannelStatus)
		s.Require().Equal(wasmdChannel.Version, contractChannelState.Channel.Version)
		s.Require().Equal(wasmdChannel.ConnectionHops[0], contractChannelState.Channel.ConnectionID)
		s.Require().Equal(wasmdChannel.ChannelID, contractChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(wasmdChannel.PortID, contractChannelState.Channel.Endpoint.PortID)
		s.Require().Equal(wasmdChannel.Counterparty.ChannelID, contractChannelState.Channel.CounterpartyEndpoint.ChannelID)
		s.Require().Equal(wasmdChannel.Counterparty.PortID, contractChannelState.Channel.CounterpartyEndpoint.PortID)
		s.Require().Equal(wasmdChannel.Ordering, contractChannelState.Channel.Order)

		// Check contract state
		contractState, err := s.Contract.QueryContractState(ctx)
		s.Require().NoError(err)

		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
		s.Require().Equal(false, contractState.AllowChannelOpenInit)

		// Give canined some time to finish the handshake

		time.Sleep(time.Duration(30) * time.Second)

		// It should be open by now

		updatedCaninedChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, canined.Config().ChainID)
		s.Require().NoError(err)

		updatedCaninedChannel := updatedCaninedChannels[0]
		s.T().Logf("canined channel state: %s", toJSONString(updatedCaninedChannel.State))
		s.Require().Equal(icatypes.HostPortID, updatedCaninedChannel.PortID)
		s.Require().Equal(s.Contract.Port(), updatedCaninedChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), updatedCaninedChannel.State)

	})
}

// toJSONString returns a string representation of the given value
// by marshaling it to JSON. It panics if marshaling fails.
func toJSONString(v any) string {
	bz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bz)
}
