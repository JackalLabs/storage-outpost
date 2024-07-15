package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/migrationtest/testsuite"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/stretchr/testify/suite"
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

3. Perform contract migration.

4. Execute updated contract and query for updated state to confirm new contract works and old state is gone.

TODO: Need more checks to show that old state is definitely gone.
*/

func TestWithMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) TestBasicMigration() {
	ctx := context.Background()
	logger.InitLogger()
	encoding := icatypes.EncodingProtobuf

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupSuite(ctx, chainSpecs)

	s.Run(fmt.Sprintf("TestRunBasicMigration-%s", encoding), func() {
		// Store v1 contract
		codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/basic_migration_v1.wasm")
		s.Require().NoError(err)

		// Instantiate v1 contract
		instantiateMsg := "{}"
		contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
		s.Require().NoError(err)

		// Create msg types for v1 and v2
		type ValueResp struct {
			Value string `json:"value"`
		}
		type Response struct {
			Data ValueResp `json:"data"`
		}
		resp := Response{
			Data: ValueResp{
				"Before Turnover",
			},
		}

		// Query the v1 contract to make sure we get the right value
		err = s.ChainA.QueryContract(ctx, contractAddr, "{\"value\":{}}", &resp)
		s.Require().NoError(err)
		s.Assert().Equal("Data saved in v1!", resp.Data.Value)

		// Store v2 contract
		v2_codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/basic_migration_v2.wasm")
		s.Require().NoError(err)

		// Migrate
		_, err = s.ChainA.MigrateContract(ctx, s.UserA.KeyName(), contractAddr, v2_codeId, "{}")
		s.Require().NoError(err)

		// Check to see if contract points to v2 CodeID
		v2_contract_info_resp, err := testsuite.GetContractInfo(ctx, s.ChainA, contractAddr)
		s.Require().NoError(err)
		int_form, _ := strconv.ParseUint(v2_codeId, 10, 64)
		s.Require().Equal(v2_contract_info_resp.CodeID, int_form)

		// Create a QueryMsg
		type QueryMsg struct {
			Data struct{} `json:"data,omitempty"`
		}

		// Instantiate a QueryMsg to get the outposts channel state
		basicQuery := QueryMsg{
			Data: struct{}{},
		}

		basicQueryBytes, err := json.Marshal(basicQuery)
		s.Require().NoError(err)
		basicQueryStr := string(basicQueryBytes)

		// Check to see if you can still query and if it returns the right value
		err = s.ChainA.QueryContract(ctx, contractAddr, basicQueryStr, &resp)
		s.Require().NoError(err)
		s.Assert().Equal("Data saved in v1!", resp.Data.Value)
	})
}
