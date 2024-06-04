package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
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

	logger.InitLogger()

	// Upload and Instantiate the contract on wasmd:
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.OutpostContractCodeId, err = strconv.ParseInt(codeId, 10, 64)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_owner.wasm")
	s.Require().NoError(err)

	instantiateMsg := outpostowner.InstantiateMsg{StorageOutpostCodeId: int(s.OutpostContractCodeId)}
	// this is the outpost owner
	outpostOwnerContractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, toString(instantiateMsg), false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	s.NumOfOutpostContracts = 0

	// Create UserA's outpost
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
	outpostAddressFromEvent := logger.ParseOutpostAddress(res.Events)
	logger.LogInfo(outpostAddressFromEvent)

	s.NumOfOutpostContracts++

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// In the docker session, we can see that the ica channel was created

	mapOutpostMsg := outpostowner.ExecuteMsg{
		MapUserOutpost: &outpostowner.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA.FormattedAddress(),
		},
	}
	// This failed because UserA already used their lock when creating the outpost
	res1, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostOwnerContractAddr, toString(mapOutpostMsg), "--gas", "500000")
	expectedErrorMsg := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg)
	fmt.Printf(res1.TxHash)

	// logger.LogEvents(res1.Events)

	// Let's get UserA2 to create a user<>outpost mapping WITHOUT creating an outpost. It will fail because no lock file exists
	mapOutpostMsgForUserA2 := outpostowner.ExecuteMsg{
		MapUserOutpost: &outpostowner.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA2.FormattedAddress(),
		},
	}

	res2, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostOwnerContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsg1 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg1)
	fmt.Printf(res2.TxHash)

	//logger.LogInfo(res2)

	// UserA2 should be able to make an outpost

	res3, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostOwnerContractAddr, toString(createMsg), "--gas", "500000")
	fmt.Printf(res3.TxHash)

	//logger.LogInfo(res3)
	s.Require().NoError(err)

	// If UserA2 tries to map again, lock file doesn't exist because it was consumed during the creation of their outpost
	res4, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostOwnerContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsg2 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg2)
	fmt.Printf(res4.TxHash)

	//logger.LogInfo(res4)

	// UserA2 tries to maliciously create a mapping for UserA3
	mapOutpostMsgForUserA3 := outpostowner.ExecuteMsg{
		MapUserOutpost: &outpostowner.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA3.FormattedAddress(), // put in UserA3 address
		},
	}
	res5, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostOwnerContractAddr, toString(mapOutpostMsgForUserA3), "--gas", "500000")
	expectedErrorMsg3 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg3)
	fmt.Printf(res5.TxHash)

	// logger.LogInfo(res5)

	// Query for the relevant addresses to ensure everything exists
	outpostAddressRes, outpostAddressErr := testsuite.OutpostAddress(ctx, s.ChainA, outpostOwnerContractAddr, s.UserA.FormattedAddress())
	s.Require().NoError(outpostAddressErr)
	var outpostAddress string
	if err := json.Unmarshal(outpostAddressRes.Data, &outpostAddress); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}

	fmt.Printf("User Outpost Address: %s\n", outpostAddress)

	// To check that the mappings were done correctly.
	// Above, we should parse out the outpost address that's created for userA using the event
	// we should then assert that it's equal to 'outpostAddress', much like how we assert PubKeys are equal
	// s.Require().Equal(pubRes.PubKey.GetKey(), filetreeMsg.GetKey(), "Expected PubKey does not match the returned PubKey")

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
