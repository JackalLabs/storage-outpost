package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"
)

/*
Here are two different testing commands in e2e/interchaintest:

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithFiletree -timeout 12h

go test -v . -run TestWithFactoryTestSuite -testify.m TestFactoryCreateOutpost -timeout 12h

Your command in e2e/migrationtest, will look similar to this:

Look something like:

go test -v . -run TestWithMigrationTestSuite -testify.m TestBasicMigration -timeout 12h
*/

func TestWithOutpostTestSuite(t *testing.T) {
	suite.Run(t, new(OutpostTestSuite))
}

/*
SetupOutpostTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
sets up the storage outpost contract and does the channel handshake for the outpost test suite.
*/
func (s *OutpostTestSuite) SetupOutpostTestSuite(ctx context.Context, encoding string) string {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	// Upload and Instantiate the storage outpost contract on wasmd:
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	admin := s.UserA.FormattedAddress()

	// Instantiate the storage outpost contract with channel:
	instantiateMsg := types.NewInstantiateMsgWithChannelInitOptions(&admin, s.ChainAConnID, s.ChainBConnID, nil, &encoding)
	contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	// Store storage_outpost_v2.wasm
	_, error := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/v2/storage_outpost_v2.wasm")
	s.Require().NoError(error)

	logger.InitLogger()
	fmt.Println("The sender of instantiate is", s.UserA.KeyName())
	logger.LogInfo("The sender of instantiate is", s.UserA.KeyName())

	s.Contract = types.NewIcaContract(types.NewContract(contractAddr, codeId, s.ChainA))

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	contractState, err := s.Contract.QueryContractState(ctx)
	s.Require().NoError(err)

	s.IcaAddress = contractState.IcaInfo.IcaAddress
	s.Contract.SetIcaAddress(s.IcaAddress)

	return contractAddr
}

func (s *OutpostTestSuite) TestOutpostCall() {
	ctx := context.Background()

	logger.InitLogger()

	encoding := icatypes.EncodingProtobuf
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	outpostAddr := s.SetupOutpostTestSuite(ctx, encoding)
	_, canined := s.ChainA, s.ChainB
	// wasmdUser := s.UserA

	logger.LogInfo(canined.FullNodes)

	// Fund the ICA address:
	s.FundAddressChainB(ctx, s.IcaAddress)

	// Give canined some time to complete the handshake
	time.Sleep(time.Duration(30) * time.Second)

	s.Run(fmt.Sprintf("TestRunHelpersCall-%s", encoding), func() {
		// Store v2 contract
		codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/basic_migration_v2.wasm")
		s.Require().NoError(err)

		// Instantiate v2 contract
		instantiateMsg := "{}"
		contractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, instantiateMsg, false, "--gas", "500000", "--admin", s.UserA.KeyName())
		s.Require().NoError(err)

		// Create the SetOutpostMsg type for interacting with the v2 contracts storage outpost wrapper
		// SetOutpostMsg corresponds to the Rust struct SetOutpostMsg
		type SetOutpostMsg struct {
			Addr string `json:"addr"`
		}

		// QueryChannelMsg corresponds to the Rust struct QueryChannelMsg
		type QueryChannelMsg struct{}

		// ExecuteMsg is a wrapper for the different execute messages
		type ExecuteMsg struct {
			SetOutpost *SetOutpostMsg `json:"set_outpost,omitempty"`
		}

		// Instantiate a SetOutpostMsg on the outpost address
		setOutpostMsg := ExecuteMsg{
			SetOutpost: &SetOutpostMsg{
				Addr: outpostAddr,
			},
		}

		// Serialize the setOutpostMsg to JSON bytes, then a string
		setOutpostBytes, err := json.Marshal(setOutpostMsg)
		s.Require().NoError(err)
		setOutpostStr := string(setOutpostBytes)

		// Set the v2 basic-migration contract to store the address of the storage outpost
		_, err = s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), contractAddr, setOutpostStr)
		s.Require().NoError(err)

		// Create a QueryMsg
		type QueryMsg struct {
			QueryOutpost *QueryChannelMsg `json:"query_outpost_channel,omitempty"`
		}

		// Instantiate a QueryMsg to get the outposts channel state
		outpostChannelMsg := QueryMsg{
			QueryOutpost: &QueryChannelMsg{},
		}

		// Serialize the query for the channel state
		queryChannelBytes, err := json.Marshal(outpostChannelMsg)
		s.Require().NoError(err)
		queryChannelStr := string(queryChannelBytes)

		type Endpoint struct {
			PortID    string `json:"port_id"`
			ChannelID string `json:"channel_id"`
		}

		type CounterPartyEndpoint struct {
			PortID    string `json:"port_id"`
			ChannelID string `json:"channel_id"`
		}

		type Channel struct {
			Endpoint             Endpoint             `json:"endpoint"`
			CounterpartyEndpoint CounterPartyEndpoint `json:"counterparty_endpoint"`
			Order                string               `json:"order"`
			Version              string               `json:"version"`
		}

		type Data struct {
			Channel       Channel `json:"channel"`
			ConnectionID  string  `json:"connection_id"`
			ChannelStatus string  `json:"channel_status"`
		}

		type Response struct {
			Data Data `json:"data"`
		}

		resp := Response{
			Data: Data{
				Channel: Channel{
					Endpoint: Endpoint{
						PortID:    "Before",
						ChannelID: "Before",
					},
					CounterpartyEndpoint: CounterPartyEndpoint{
						PortID:    "Before",
						ChannelID: "Before",
					},
					Order:   "Before",
					Version: "Before",
				},
				ConnectionID:  "Before",
				ChannelStatus: "Before",
			},
		}

		// Use the v2 basic-migration contract to query the storage_outpost for its channel state
		err = s.ChainA.QueryContract(ctx, contractAddr, queryChannelStr, &resp)
		s.Require().NoError(err)
		s.Require().Equal(resp.Data.ChannelStatus, "STATE_OPEN")
	})
}
