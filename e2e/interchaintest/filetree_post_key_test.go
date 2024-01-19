package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestIcaContractExecutionTestWithFiletree() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	wasmdUser := s.UserA

	logger.LogInfo(canined.FullNodes)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestSendCustomIcaMesssagesSuccess-%s", encoding), func() {
		filetreeMsg := filetreetypes.MsgPostkey{
			// NOTE: filetree is calling a testing helpers function, "MakePrivateKey"
			// Not sure why this happens nor where it's happening in the call stack
			// Perhaps because the wasmdUser address is not a jkl bech32 address so this function was called
			// to create an interim correct Creator address

			// Update: The above error is likely arising from the fact that canined uses cosmos-sdk 0.45 but
			// This test suite uses cosmos-sdk 0.47
			Creator: wasmdUser.KeyName(), // This will soon be the contract address
			Key:     "Hey it's Bi from the outpost on another chain. We reached filetree!!! <3",
		}

		// Execute the contract:
		err := s.Contract.ExecCustomIcaMessages(ctx, wasmdUser.KeyName(), []proto.Message{&filetreeMsg}, encoding, nil, nil)
		s.Require().NoError(err)

		// We haven't implemented call backs so at this point we could just start a shell session in the container to
		// view the filetree entry

	},
	)

	time.Sleep(time.Duration(1) * time.Hour)

}
