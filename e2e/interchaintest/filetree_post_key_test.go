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
		filetreeMsg := &filetreetypes.MsgPostKey{
			Creator: s.Contract.IcaAddress,
			// we're just hard coding this temporarily for debugging purposes
			// It's the correct jkl ICA address

			// This will soon be the contract address
			// This has to be the jkl address that's created by the controller (this contract)
			// When the channel is opened. If it's not this address, the transaction should error
			// Because the controller account should only be allowed to execute msgs for its host pair
			Key: "Wow it really works <3",
		}

		// func NewAnyWithValue(v proto.Message) (*Any, error) {} inside ica_msg.go is not returning the type URL of the filetree msg

		referencedTypeUrl := sdk.MsgTypeURL(filetreeMsg)

		fmt.Println("filetree msg satisfy sdk Msg interface?:", referencedTypeUrl)
		logger.LogInfo(referencedTypeUrl)

		fmt.Println("filetree msg as string is", filetreeMsg.String())

		// Filetree msg sent!
		// FOR TEAM: start a shell session within canined's container and run:
		// canined q filetree list-pubkeys
		// to see the posted public key

		// TO DO: Call backs to confirm success

		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{filetreeMsg}, nil, nil,
		)
		error := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(error)

	},
	)

	time.Sleep(time.Duration(10) * time.Hour)

}
