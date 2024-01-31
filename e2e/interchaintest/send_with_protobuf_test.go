package main

import (
	"context"
	"fmt"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	sdkmath "cosmossdk.io/math"
	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues because the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestIcaContractExecutionTestWithProtobuf() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	caninedUser := s.UserB
	fmt.Println(caninedUser)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestStargate-%s", encoding), func() {
		// Send custom ICA messages through the contract:
		// Let's create a governance proposal on canined and deposit some funds to it.
		testProposal := govtypes.TextProposal{ // WARNING: This is from cosmos-sdk v0.47. If canined rejects it, could be a versioning/protobuf type issue
			Title:       "IBC Gov Proposal",
			Description: "Hey it's Bi sending tokens from the outpost",
		}
		protoAny, err := codectypes.NewAnyWithValue(&testProposal)
		s.Require().NoError(err)

		proposalMsg := &govtypes.MsgSubmitProposal{
			Content:        protoAny,
			InitialDeposit: sdk.NewCoins(sdk.NewCoin(canined.Config().Denom, sdkmath.NewInt(5_000))),
			Proposer:       s.IcaAddress,
		}

		// Create deposit message:
		depositMsg := &govtypes.MsgDeposit{
			ProposalId: 1,
			Depositor:  s.IcaAddress,
			Amount:     sdk.NewCoins(sdk.NewCoin(canined.Config().Denom, sdkmath.NewInt(10_000_000))),
		}

		initialBalance, err := canined.GetBalance(ctx, s.IcaAddress, canined.Config().Denom)
		s.Require().NoError(err)

		logger.LogInfo("initial balance is:", initialBalance)

		logger.LogInfo("Executing custom ICA message now")
		fmt.Println("Executing custom ICA message now")

		// stargate API v0.4.2 works for proposalMsg executing on canined
		// still does not work for filetree at this time.

		sendStargateMsg := testtypes.NewExecuteMsg_SendCosmosMsgs_FromProto(
			[]proto.Message{proposalMsg, depositMsg}, nil, nil,
		)
		err = s.Contract.Execute(ctx, wasmdUser.KeyName(), sendStargateMsg)
		s.Require().NoError(err)

		// err = s.Contract.ExecSendStargateMsgs(ctx, wasmdUser.KeyName(), []proto.Message{proposalMsg, depositMsg}, nil, nil)

		// err = s.Contract.ExecCustomIcaMessages(ctx, wasmdUser.KeyName(), []proto.Message{proposalMsg, depositMsg}, encoding, nil, nil)

		s.Require().NoError(err)
	})

	time.Sleep(time.Duration(1) * time.Hour)

}
