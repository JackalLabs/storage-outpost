package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestReOpenOrderedChannel() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	wasmd, canined := s.ChainA, s.ChainB
	fmt.Println(wasmd)
	wasmdUser := s.UserA
	fmt.Println(wasmdUser)

	logger.LogInfo(canined.FullNodes)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	contractState, err := s.Contract.QueryContractState(ctx)
	s.Require().NoError(err)

	s.Run("TestCloseChannelOnTimeout", func() {

		// We will send a message to the host that will timeout after 3 seconds.
		// You cannot use 0 seconds because block timestamp will be greater than the timeout timestamp which is not allowed.
		// Host will not be able to respond to this message in time.

		// Stop the relayer so that the host cannot respond to the message:

		error := s.Relayer.StopRelayer(ctx, s.ExecRep)
		s.Require().NoError(error)

		time.Sleep(5 * time.Second)

		timeout := uint64(3)

		filetreeMsg := &filetreetypes.MsgPostKey{
			Creator: s.Contract.IcaAddress,
			Key:     "Wow it really works <3",
		}
		typeURL := "/canine_chain.filetree.MsgPostKey"
		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{filetreeMsg}, nil, &timeout, typeURL,
		)
		executeErr := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(executeErr)

		// Wait until timeout:
		err = testutil.WaitForBlocks(ctx, 5, wasmd, canined)
		s.Require().NoError(err)

		err = s.Relayer.StartRelayer(ctx, s.ExecRep)
		s.Require().NoError(err)

		// Wait until timeout packet is received:
		err = testutil.WaitForBlocks(ctx, 2, wasmd, canined)
		s.Require().NoError(err)

		// Flush to make sure the channel is closed in simd:
		err = s.Relayer.Flush(ctx, s.ExecRep, s.PathName, contractState.IcaInfo.ChannelID)
		s.Require().NoError(err)

		err = testutil.WaitForBlocks(ctx, 2, wasmd, canined)
		s.Require().NoError(err)

		// Check if channel was closed:
		wasmdChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, wasmd.Config().ChainID)
		s.Require().NoError(err)
		s.Require().Equal(1, len(wasmdChannels))
		s.Require().Equal(channeltypes.CLOSED.String(), wasmdChannels[0].State)

		// Query to make sure pub key was not saved

		// To confirm that the package did not make it to the ica host, Query a PubKey to make sure it wasn't saved
		pubRes, pubErr := testsuite.PubKey(ctx, s.ChainB, s.Contract.IcaAddress)
		logger.LogInfo(pubRes)
		s.Require().EqualError(pubErr, "rpc error: code = NotFound desc = not found")

		// s.Require().Equal(pubRes.PubKey.GetKey(), filetreeMsg.GetKey(), "Expected PubKey does not match the returned PubKey")

		// Make sure the outpost itself knows that the channel is closed
		outpostChannelState, err := s.Contract.QueryChannelState(ctx)
		s.Require().NoError(err)
		s.Require().Equal(testtypes.ChannelStatus_StateClosed, outpostChannelState.ChannelStatus)

	})

	s.Run("TestReOpenOrderedChannelAndPostKey", func() {
		// Reopen the channel:

		createChannelMsg := testtypes.ExecuteMsg{
			CreateChannel: &testtypes.ExecuteMsg_CreateChannel{
				ChannelOpenInitOptions: nil, // consequence of not putting open init options?...
			},
		}
		//	contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())

		executeErr := s.Contract.Execute(ctx, wasmdUser.KeyName(), createChannelMsg, "--gas", "500000")
		s.Require().NoError(executeErr)

		// Wait for the channel to get set up
		err = testutil.WaitForBlocks(ctx, 10, s.ChainA, s.ChainB)
		s.Require().NoError(err)

		// Check if a new channel was opened in canined
		caninedChannels, err := s.Relayer.GetChannels(ctx, s.ExecRep, canined.Config().ChainID)
		logger.LogInfo(caninedChannels)
		logger.LogInfo(err)
		s.Require().NoError(err)
		s.Require().Equal(channeltypes.OPEN.String(), caninedChannels[1].State) // First channel closed, second channel identical to 1st but is open

		// Ensure outpost recognizes new channel
		outpostChannelState, err := s.Contract.QueryChannelState(ctx)
		s.Require().NoError(err)
		logger.LogInfo(outpostChannelState.Channel.Endpoint.ChannelID)
		s.Require().Equal(outpostChannelState.Channel.Endpoint.ChannelID, "channel-1")

		// send filetree msg without timeout and confirm it was received
		filetreeMsg := &filetreetypes.MsgPostKey{
			Creator: s.Contract.IcaAddress,
			Key:     "Wow it really works <3",
		}
		typeURL := "/canine_chain.filetree.MsgPostKey"
		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{filetreeMsg}, nil, nil, typeURL,
		)
		executeError := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(executeError)

		err = testutil.WaitForBlocks(ctx, 10, wasmd, canined)
		s.Require().NoError(err)

		// Query the PubKey
		pubRes, pubErr := testsuite.PubKey(ctx, s.ChainB, s.Contract.IcaAddress)
		s.Require().NoError(pubErr)
		s.Require().Equal(pubRes.PubKey.GetKey(), filetreeMsg.GetKey(), "Expected PubKey does not match the returned PubKey")

		fmt.Printf("*****TEST DONE*****")
		// time.Sleep(time.Duration(10) * time.Hour)

		// TODO: If it works, ensure that the channel is actually opened

	})

}
