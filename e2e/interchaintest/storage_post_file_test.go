package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	storagetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/storagetypes"

	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestPostFile() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	wasmd, canined := s.ChainA, s.ChainB
	fmt.Println(wasmd)
	wasmdUser := s.UserA

	logger.LogInfo(canined.FullNodes)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestSendCustomIcaMesssagesSuccess-%s", encoding), func() {

		merkleBytes := []byte{0x01, 0x02, 0x03, 0x04}
		postFileMsg := &storagetypes.MsgPostFile{
			Creator:       s.Contract.IcaAddress,
			Merkle:        merkleBytes,
			FileSize:      100000000,
			ProofInterval: 60,
			ProofType:     1,
			MaxProofs:     100,
			Expires:       100 + ((100 * 365 * 24 * 60 * 60) / 6),
			Note:          `{"description": "alice note", "additional_info": "placeholder"}`,
		}

		// NOTE: func NewAnyWithValue(v proto.Message) (*Any, error) {} inside ica_msg.go is not returning the type URL of the filetree msg

		// referencedTypeUrl := sdk.MsgTypeURL(postFileMsg)

		// fmt.Println("filetree msg satisfy sdk Msg interface?:", referencedTypeUrl)
		// logger.LogInfo(referencedTypeUrl)

		// fmt.Println("filetree msg as string is", filetreeMsg.String())

		// Filetree msg sent!
		// FOR TEAM: start a shell session within canined's container and run:
		// canined q filetree list-pubkeys
		// to see the posted public key

		// TO DO: Call backs to confirm success

		typeURL := "/canine_chain.storage.MsgPostFile"

		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{postFileMsg}, nil, nil, typeURL,
		)
		// TODO: Confirm owner and admin
		error := s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(error)

		// // Query a PubKey
		// pubRes, pubErr := testsuite.PubKey(ctx, s.ChainB, s.Contract.IcaAddress)
		// s.Require().NoError(pubErr)
		// s.Require().Equal(pubRes.PubKey.GetKey(), filetreeMsg.GetKey(), "Expected PubKey does not match the returned PubKey")

	},
	)
	// implement mock query server
	time.Sleep(time.Duration(10) * time.Hour)
}
