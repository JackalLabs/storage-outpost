package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		filetreeMsg := filetreetypes.MsgPostKey{
			Creator: wasmdUser.FormattedAddress(), // This will soon be the contract address
			Key:     "Hey it's Bi from the outpost on another chain. We reached filetree!!! <3",
		}

		// func NewAnyWithValue(v proto.Message) (*Any, error) {} inside ica_msg.go is not returning the type URL of the filetree msg

		referencedMsg := &filetreeMsg
		referencedTypeUrl := sdk.MsgTypeURL(referencedMsg)

		fmt.Println("filetree msg satisfy sdk Msg interface?:", referencedTypeUrl)
		logger.LogInfo(referencedTypeUrl)

		// Execute the contract:
		err := s.Contract.ExecSendStargateMsgs(ctx, wasmdUser.KeyName(), []proto.Message{&filetreeMsg}, nil, nil)
		s.Require().NoError(err)

		// We haven't implemented call backs so at this point we could just start a shell session in the container to
		// view the filetree entry

	},
	)

	time.Sleep(time.Duration(1) * time.Hour)

}
