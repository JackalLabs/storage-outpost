package main

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	mysuite "github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	outpostowner "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types/outpostowner"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"
)

type OwnerTestSuite struct {
	mysuite.TestSuite

	OutpostContractCodeId int64

	Contract              *types.IcaContract
	IcaHostAddress        string
	NumOfOutpostContracts uint32
}

// SetupContractAndChannel starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *OwnerTestSuite) SetupOwnerTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	// Upload and Instantiate the contract on wasmd:
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	// admin := s.UserA.FormattedAddress()

	// codeId is string and needs to be converted to uint64
	s.OutpostContractCodeId, err = strconv.ParseInt(codeId, 10, 64)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_owner.wasm")
	s.Require().NoError(err)

	instantiateMsg := outpostowner.InstantiateMsg{StorageOutpostCodeId: int(s.OutpostContractCodeId)}

	outpostOwnerContractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, toString(instantiateMsg), false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	s.NumOfOutpostContracts = 0

	// Create the Outpost Contract
	createMsg := outpostowner.ExecuteMsg{
		CreateIcaContract: &outpostowner.ExecuteMsg_CreateIcaContract{
			Salt: nil,
			ChannelOpenInitOptions: outpostowner.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
				TxEncoding:               outpostowner.TxEncoding(encoding),
			},
		},
	}

	res, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostOwnerContractAddr, toString(createMsg), "--gas", "500000")
	s.Require().NoError(err)
	logger.LogInfo(res)

	s.NumOfOutpostContracts++

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// In the docker session, we can see that the ica channel was created

	mapOutpostMsg := outpostowner.ExecuteMsg{
		MapUserOutpost: &outpostowner.ExecuteMsg_MapUserOutpost{
			OutpostAddress: "wasm1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrss5maay",
			OutpostOwner:   s.UserA.FormattedAddress(),
		},
	}
	// This should fail but the events should still emit no? or is it just the error that we get back?
	updatedResponse, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostOwnerContractAddr, toString(mapOutpostMsg), "--gas", "500000")
	s.Require().NoError(err)

	logger.InitLogger()
	logger.LogEvents(updatedResponse.Events)
}

func TestWithOwnerTestSuite(t *testing.T) {
	suite.Run(t, new(OwnerTestSuite))
}

func (s *OwnerTestSuite) TestOwnerCreateIcaContract() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupOwnerTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json
	// wasmd, canined := s.ChainA, s.ChainB

	// We weren't able to precompute the outpost's address at the time of creation, so we need to query for the address
	// right now
	// query by code ID and sender address? The sender being the user that executed the creation
	// The port id of the outpost should be wasm.contractAddress so can't we retrieve the address from that?

	time.Sleep(time.Duration(10) * time.Hour)

}

// // toJSONString returns a string representation of the given value
// // by marshaling it to JSON. It panics if marshaling fails.
// func toJSONString(v any) string {
// 	bz, err := json.Marshal(v)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(bz)
// }

// toString converts the message to a string using json
func toString(msg any) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
