package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
)

func (s *ContractTestSuite) TestIcaParams() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupContractTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	// wasmdUser := s.UserA

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestIcaParamsSuccess-%s", encoding), func() {

		hostParams, err := canined.QueryParam(ctx, "icahost", "allow_messages")
		s.Require().NoError(err)

		hp, err := json.MarshalIndent(hostParams, "", "  ")

		logger.LogInfo("host params are")

		if err != nil {
			// handle error
			logger.LogError("failed to marshal host params:", err)
		} else {
			logger.LogInfo(string(hp))
		}

	},
	)
	/*
		   	canined.QueryParam returns this:

		   host params are

		   	INFO: 2023/12/21 13:21:13 {
		   	  "subspace": "icahost",
		   	  "key": "allow_messages",
		   	  "value": ""
		   	}

		   but interestingly

		   'canined q ica host params' returns this:

		   allow_messages:
			- '*'
		host_enabled: true
	*/
}
