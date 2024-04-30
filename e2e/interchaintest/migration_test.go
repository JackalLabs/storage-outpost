package main

import (
	"context"
	"encoding/base64"
	"time"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bech32 "github.com/cosmos/cosmos-sdk/types/bech32"
)

func (s *ContractTestSuite) TestIcaContractExecutionTestWithMigration() {
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

	s.Run("TestMigrateAndUpdateAdmin", func() {

		migrateMsg1 := testtypes.MigrateMsg{
			ContractAddr: "wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d", //same contract address but different code id
			NewCodeID:    "2",                                                               //it should error and say something like "NewCodeID does not exist"
			Msg:          "placeholder",
		}

		// contractByCodeRequest := wasmtypes.QueryContractsByCodeRequest{
		// 	CodeId: uint64(1),
		// }
		// contractByCodeResp, err := e2esuite.GRPCQuery[wasmtypes.QueryContractsByCodeResponse](ctx, wasmd2, &contractByCodeRequest)
		// s.Require().NoError(err)
		// s.Require().Len(contractByCodeResp.Contracts, 1)

		_, bz, _ := bech32.DecodeAndConvert(wasmdUser.FormattedAddress())

		newPrefix := "cosmos"

		newAddress, _ := bech32.ConvertAndEncode(newPrefix, bz)

		callerAddress, err := sdk.AccAddressFromBech32(newAddress)
		logger.LogInfo("error is", err)
		s.Require().NoError(err)

		logger.LogInfo("migration caller", callerAddress)
		logger.LogInfo("migration caller as string", callerAddress)

		// Wasmd is expecting AccAddress type but SL interchaintest package only takes in strings
		// Need to convert it to AccAddress type inside of SL package, that would be easier
		// for debugging than forking wasmd.
		// tricky to debug because we didn't build wasmd image.
		// NOTE: might be best to deploy outpost on canine-chain clone so we can
		// use a wasmd fork which has logging
		error := s.Contract.Migrate(ctx, wasmdUser.FormattedAddress(), "2", migrateMsg1)
		// gonna have to change back to wasmdUser
		s.Require().NoError(error)

	})

	time.Sleep(time.Duration(10) * time.Hour)

}

func toBase64(msg string) string {
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
