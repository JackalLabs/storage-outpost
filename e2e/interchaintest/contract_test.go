package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	mysuite "github.com/JackalLabs/cw-ica-controller/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/cw-ica-controller/e2e/interchaintest/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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

func (s *ContractTestSuite) TestIcaContractChannelHandshake() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, icatypes.EncodingProto3JSON)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	fmt.Println(simd)
	fmt.Println(wasmdUser)
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

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)

		simdChannel := simdChannels[0]
		s.T().Logf("simd channel state: %s", toJSONString(simdChannel.State))
		s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
		s.Require().Equal(s.Contract.Port(), simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

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

		s.Require().Equal(wasmdUser.FormattedAddress(), contractState.Admin)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
	})
}

func (s *ContractTestSuite) TestIcaRelayerInstantiatedChannelHandshake() {
	ctx := context.Background()

	// This starts the
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)
	wasmd, simd := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	var err error
	// Upload and Instantiate the contract on wasmd:
	s.Contract, err = types.StoreAndInstantiateNewContract(ctx, wasmd, wasmdUser.KeyName(), "../../artifacts/cw_ica_controller.wasm")
	s.Require().NoError(err)

	version := fmt.Sprintf(`{"version":"%s","controller_connection_id":"%s","host_connection_id":"%s","address":"","encoding":"%s","tx_type":"%s"}`, icatypes.Version, s.ChainAConnID, s.ChainBConnID, icatypes.EncodingProtobuf, icatypes.TxTypeSDKMultiMsg)
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

		simdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, simd.Config().ChainID)
		s.Require().NoError(err)

		simdChannel := simdChannels[0]
		s.T().Logf("simd channel state: %s", toJSONString(simdChannel.State))
		s.Require().Equal(icatypes.HostPortID, simdChannel.PortID)
		s.Require().Equal(s.Contract.Port(), simdChannel.Counterparty.PortID)
		s.Require().Equal(channeltypes.OPEN.String(), simdChannel.State)

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

		s.Require().Equal(wasmdUser.FormattedAddress(), contractState.Admin)
		s.Require().Equal(wasmdChannel.ChannelID, contractState.IcaInfo.ChannelID)
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
