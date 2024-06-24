package main

import (
	"context"
	"testing"

	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/stretchr/testify/suite"
	// interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	//"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	//"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

/*
Here are two different testing commands in e2e/interchaintest:

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithFiletree -timeout 12h

go test -v . -run TestWithFactoryTestSuite -testify.m TestFactoryCreateOutpost -timeout 12h

Your command in e2e/migrationtest, will look similar to this:

Look something like:

go test -v . -run TestWithMigrationTestSuite -testify.m TestBasicMigration -timeout 12h

Your migration object to act upon:

	type MigrationTestSuite struct {
		mysuite.TestSuite

		Contract              *types.IcaContract
		IcaAddress            string
		FaucetOutpostContract *types.IcaContract
		FaucetJKLHostAddress  string

}

OBJECTIVE:

1. Spin up the two chains. For now, you don't need to interact with canined, but you will later. You're only
interacting with wasmd at the moment.

2. Upload v1 of basic migration. Confirm that v1 works by executing contract and querying contract for state to confirm that execution was
successful.

3. Perform congration migration.

4. Execute updated contract and query for updated state to confirm new contract works and old state is gone.

TODO: Need more checks to show that old state is definitely gone.
*/

func TestWithMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) TestBasicMigration() {
	ctx := context.Background()

	logger.InitLogger()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupSuite(ctx, chainSpecs)
	_, canined := s.ChainA, s.ChainB

	logger.LogInfo(canined.FullNodes)
}
