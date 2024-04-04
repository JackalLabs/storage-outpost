package main

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
)

// WARNING: strangelove's test package builds chains running ibc-go/v7
// Hopefully this won't cause issues even though the canined image we use is running ibc-go/v4
// and packets should be consumed by the ica host no matter what version of ibc-go the controller chain is running

func (s *ContractTestSuite) TestIcaContractExecutionTestWithBuyStorage() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	wasmdUser := s.UserA
	// caninedUser := s.UserB

	logger.LogInfo(canined.FullNodes)
	logger.LogInfo("The wasmd user is:", wasmdUser.FormattedAddress())

	// NOTE: we're commenting out this code so that the IcaAddress won't have any funds until the user
	// story of buying storage is complete
	// Fund the ICA address:
	// s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestSendCustomIcaMesssagesSuccess-%s", encoding), func() {

		// let's open the transfer channel

		CounterpartyPortId := "transfer"

		createTransferChannelMsg := testtypes.ExecuteMsg{
			CreateTransferChannel: &testtypes.ExecuteMsg_CreateTransferChannel{
				// NOTE: in contract.rs, the order of these params is: connection_id, counterpart_port_id, counterparty_connection_id
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
				CounterpartyPortId:       &CounterpartyPortId,
			},
		}
		err := s.Contract.Execute(ctx, wasmdUser.KeyName(), createTransferChannelMsg)
		s.Require().NoError(err)

		// Give the transfer channel some time to be in the OPEN state
		time.Sleep(time.Duration(60) * time.Second)

		var walletAmount = ibc.WalletAmount{
			Address: wasmdUser.FormattedAddress(),
			Denom:   "ujkl",
			Amount:  math.NewInt(689000),
		}

		denomTrace := types.ParseDenomTrace("ujkl")
		logger.LogInfo("ujkl denomTrace is:", denomTrace)
		ibcDenom := denomTrace.IBCDenom()
		logger.LogInfo("ujkl converted to ibc denom is:", ibcDenom)

		var transferOptions = ibc.TransferOptions{
			Timeout: &ibc.IBCTimeout{
				// does it use a default if these values not set?
			},
			Memo: "none",
		}
		// We know the transfer channel will consistently have a channel id of 'channel-1'
		tx, err := canined.SendIBCTransfer(ctx, "channel-1", s.UserB.KeyName(), walletAmount, transferOptions)
		s.Require().NoError(err)

		logger.LogInfo("The IBC tx hash is:", tx.TxHash)

		/* NOTE: The transfer was successful but we're having trouble printing out the tx hash and getting this error

						=== NAME  TestWithContractTestSuite/TestIcaContractExecutionTestWithBuyStorage/TestSendCustomIcaMesssagesSuccess-proto3
				    buy_storage_test.go:80:
				                Error Trace:    /Users/biphan/jackal/storage-outpost/e2e/interchaintest/buy_storage_test.go:80
				                                                        /Users/biphan/go/pkg/mod/github.com/stretchr/testify@v1.8.4/suite/suite.go:112
				                Error:          Received unexpected error:
				                                failed to get transaction 3F19EB167CB07AA598ED00306770055AB19FB6926F5FC0873A5EDA9F4268C5DA: unable to resolve type URL /ibc.applications.transfer.v1.MsgTransfer: tx parse error [cosmos/cosmos-sdk@v0.47.5/x/auth/tx/decoder.go:42]
				                Test:           TestWithContractTestSuite/TestIcaContractExecutionTestWithBuyStorage/TestSendCustomIcaMesssagesSuccess-proto3

				But the amount came through though:

		/opt # wasmd q bank balances wasm13w0fse6k9tvrq6zn68smdl6ln4s7kmh9fvq8ag
		balances:
		- amount: "689000"
		  denom: ibc/08D1E6BD9CB813AE1E5FF4C0EBC9F4B96B1F3D23DE75077EE6BE79127C497145
		- amount: "10000000000"
		  denom: stake
		pagination:
		  next_key: null
		  total: "0"
		/opt #

		Likely caused by canine-chain being on v0.45 and not v0.47

		*/

	},
	)

	time.Sleep(time.Duration(10) * time.Hour)

}

/*

Buying Storage on Archway:

1. 'mint' jkl on wasmd. Not sure how right now so we're going to make jkl user do ibc transfer of jkl tokens to wasmd user
so we can simulate jkl existing as an ibc token on wasmd. For Archway, jkl would exist as an ibc
token in the Astrovault DEX. For now, this is our best efforts in simulating
jkl on Archway.

2. Wasmd user will broadcast an ibc transfer of their jkl tokens over to ica host on jackal.
This is best efforts in simulating the user funding their ica host address with jkl tokens.
When the wasmd user sends the jkl tokens (on wasmd) back to canine-chain, the token should reclaim its ujkl denom
instead of having the 'ibc/hash()' denom, and the jkl host should have enough funds to buy storage

3. Have the contract use the SendCosmosMsg (Stargate) API to send a 'Buy Storage' to the ica host
to buy Jackal Storage!

SendIBCTransfer from SL interchaintest package is the function we want.

*/
