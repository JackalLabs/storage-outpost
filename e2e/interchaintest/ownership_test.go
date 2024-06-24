package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestIcaContractExecutionTestWithOwnership() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	wasmdUserA := s.UserA
	wasmdUserA2 := s.UserA2

	logger.LogInfo(canined.FullNodes)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestSendCustomIcaMesssagesSuccess-%s", encoding), func() {
		postKeyMsg0 := &filetreetypes.MsgPostKey{
			Creator: s.Contract.IcaAddress,
			Key:     "Wow it really works <3",
		}

		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{postKeyMsg0}, nil, nil, "/canine_chain.filetree.MsgPostKey",
		)
		error := s.Contract.Execute(ctx, wasmdUserA.KeyName(), sendStargateMsg)
		s.Require().NoError(error)

		logger.LogInfo("wasmd primary user:", wasmdUserA.FormattedAddress())
		logger.LogInfo("wasmd secondary user:", wasmdUserA2.FormattedAddress())

		// Let's have wasmdUserA2 attempt to overwrite wasmduserA's public key
		postKeyMsg1 := &filetreetypes.MsgPostKey{
			Creator: s.Contract.IcaAddress, // The ica host address which UserA created by instantiating the outpost
			Key:     "This is the take over >:)",
		}

		sendStargateMsg2 := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{postKeyMsg1}, nil, nil, "/canine_chain.filetree.MsgPostKey",
		)
		err := s.Contract.Execute(ctx, wasmdUserA2.KeyName(), sendStargateMsg2)
		expectedErrorMsg := "error in transaction (code: 5): failed to execute message; message index: 0: Caller is not the contract's current owner: execute wasm contract failed"
		s.Require().EqualError(err, expectedErrorMsg)

	},
	)

	// time.Sleep(time.Duration(10) * time.Hour)

}
