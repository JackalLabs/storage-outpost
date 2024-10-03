package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"

	storagetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/storagetypes"
	outpostuser "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types/outpostuser"

	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestOutpostUser() {
	ctx := context.Background()

	logger.InitLogger()

	// Need to instantiate it with the address of the outpost user as the owner

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

	// Let's go ahead and instantiate the outpost user, giving it the address of the outpost
	// Upload and Instantiate the contract on wasmd:
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_user.wasm")
	s.Require().NoError(err)

	// TODO: Can't init outpost user with outpost address - chicken and egg situation
	instantiateMsg := testtypes.NewInstantiateMsgWithOutpostAddress(&s.Contract.Address)

	outpostUserContract, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
	logger.LogInfo(fmt.Sprintf("outpost user address is: %s\n", outpostUserContract))
	s.Require().NoError(err)

	// The 'Setup Function above' sets up relays, channels, and inits an outpost contract.
	// We don't want to change the above function because other tests rely on it
	// For this purposes of this test, we will init a brand new outpost with the address of 'outpost user' as the owner
	admin := s.UserA.FormattedAddress()

	// Instantiate the contract with channel:
	outpostInstantiateMsg := types.InitOutpostWithOwner(&admin, s.ChainAConnID, s.ChainBConnID, nil, &encoding, &outpostUserContract)

	// We know that wasm module of outpost has code ID 1
	outpostAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), "1", outpostInstantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	// Update test suite with new outpost address
	s.Contract.Address = outpostAddr

	logger.LogInfo(fmt.Sprintf("The outpost address is: %s\n", outpostAddr))

	// Query the outpost owner
	outpostOwner, ownerErr := testsuite.GetOutpostOwner(ctx, s.ChainA, outpostAddr)
	s.Require().NoError(ownerErr)
	logger.LogInfo(fmt.Sprintf("The outpost owner is: %s\n", outpostOwner))

	s.Run(fmt.Sprintf("TestOutpostUserSuccess-%s", encoding), func() {

		merkleBytes := []byte{0x01, 0x02, 0x03, 0x04}
		postFileMsg := &storagetypes.MsgPostFile{
			Creator:       s.Contract.IcaAddress,
			Merkle:        merkleBytes,
			FileSize:      100000000,
			ProofInterval: 60,
			ProofType:     1,
			MaxProofs:     100,
			Expires:       100 + ((100 * 365 * 24 * 60 * 60) / 6),
			Note:          `{"description": "outpost user note", "additional_info": "placeholder"}`,
		}

		typeURL := "/canine_chain.storage.MsgPostFile"

		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{postFileMsg}, nil, nil, typeURL,
		)

		// NOTE: Double check this before calling it
		innerOutpostMsg := outpostuser.ExecuteMsg_CallOutpost{
			Msg: &sendStargateMsg,
		}

		outpostUserMsg := outpostuser.ExecuteMsg{
			CallOutpost: &innerOutpostMsg,
		}

		// WARNING NOTE: Only the owner of the outpost can call it.
		// The below execution doesn't work because cross contract calls are made with the calling contract's address as the sender
		// Unfortunately, UserA is set as the outpost owner because UserA instantiated it
		// Seems there's no way around this but to have the outpost user contract also instantiate the outpost

		// We know 'instantiate2' works on canine-chain, so perhaps we can use that and avoid having to use a callback
		badRes, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostUserContract, outpostUserMsg.ToString(), "--gas", "500000")
		s.Require().NoError(err)
		fmt.Println(badRes)

	},
	)
	// implement mock query server
	time.Sleep(time.Duration(10) * time.Hour)
}
