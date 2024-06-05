package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	mysuite "github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	outpostfactory "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types/outpostfactory"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"
)

type FactoryTestSuite struct {
	mysuite.TestSuite

	OutpostContractCodeId int64

	Contract              *types.IcaContract
	IcaHostAddress        string
	NumOfOutpostContracts uint32
}

// SetupContractAndChannel starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *FactoryTestSuite) SetupFactoryTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	logger.InitLogger()

	// TODO: how does the factory know the code ID of the outpost?
	// Upload the outpost's wasm module on Wasmd
	codeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.OutpostContractCodeId, err = strconv.ParseInt(codeId, 10, 64)
	s.Require().NoError(err)

	codeId, err = s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_factory.wasm")
	s.Require().NoError(err)

	instantiateMsg := outpostfactory.InstantiateMsg{StorageOutpostCodeId: int(s.OutpostContractCodeId)}
	// this is the outpost factory
	outpostfactoryContractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), codeId, toString(instantiateMsg), false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)

	s.NumOfOutpostContracts = 0

	// TODO: wrapping the encoding with 'TxEncoding' is not needed anymore because 'Proto3Json'
	// is not the recommended encoding type for the ICA channel
	// we should just use an optional string
	proto3Encoding := outpostfactory.TxEncoding(encoding)

	// Create UserA's outpost
	createMsg := outpostfactory.ExecuteMsg{
		CreateOutpost: &outpostfactory.ExecuteMsg_CreateOutpost{
			Salt: nil,
			ChannelOpenInitOptions: outpostfactory.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
				TxEncoding:               &proto3Encoding,
			},
		},
	}

	res, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(createMsg), "--gas", "500000")
	s.Require().NoError(err)
	outpostAddressFromEvent := logger.ParseOutpostAddress(res.Events)
	logger.LogInfo(outpostAddressFromEvent)

	// We know that the outpost we just made emitted an event showing its address
	// We can now query the mapping inside of 'outpost factory' to confirm that we mapped the correct address
	// Query for the relevant addresses to ensure everything exists
	outpostAddressRes, addressErr := testsuite.GetOutpostAddress(ctx, s.ChainA, outpostfactoryContractAddr, s.UserA.FormattedAddress())
	s.Require().NoError(addressErr)
	var mappedOutpostAddress string
	if err := json.Unmarshal(outpostAddressRes.Data, &mappedOutpostAddress); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(outpostAddressFromEvent, mappedOutpostAddress)

	fmt.Printf("User Outpost Address: %s\n", mappedOutpostAddress)

	// TODO: Why do we need this?
	s.NumOfOutpostContracts++

	// TODO: are we getting 'mappedOutpostAddress' correctly to be used in the Equality assertion?
	// is UserA allowed to just create another outpost again? They shouldn't be able to
	_, creationErr := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(createMsg), "--gas", "500000")
	expectedCreationErrorMsg := fmt.Sprintf("error in transaction (code: 5): failed to execute message; message index: 0:"+
		" Outpost already created. Outpost Address: %s: execute wasm contract failed", mappedOutpostAddress)
	s.Require().EqualError(creationErr, expectedCreationErrorMsg)

	// Wait for the channel to get set up
	err = testutil.WaitForBlocks(ctx, 5, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// In the docker session, we can see that the ica channel was created

	mapOutpostMsg := outpostfactory.ExecuteMsg{
		MapUserOutpost: &outpostfactory.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA.FormattedAddress(),
		},
	}
	// This failed because UserA already used their lock when creating the outpost
	res1, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsg), "--gas", "500000")
	expectedErrorMsg := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg)
	fmt.Printf(res1.TxHash)

	// logger.LogEvents(res1.Events)

	// Let's get UserA2 to create a user<>outpost mapping WITHOUT creating an outpost. It will fail because no lock file exists
	mapOutpostMsgForUserA2 := outpostfactory.ExecuteMsg{
		MapUserOutpost: &outpostfactory.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA2.FormattedAddress(),
		},
	}

	res2, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsg1 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg1)
	fmt.Printf(res2.TxHash)

	//logger.LogInfo(res2)

	// UserA2 should be able to make an outpost

	res3, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(createMsg), "--gas", "500000")
	fmt.Printf(res3.TxHash)

	//logger.LogInfo(res3)
	s.Require().NoError(err)

	// If UserA2 tries to map again, lock file doesn't exist because it was consumed during the creation of their outpost
	res4, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsg2 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg2)
	fmt.Printf(res4.TxHash)

	//logger.LogInfo(res4)

	// UserA2 tries to maliciously create a mapping for UserA3
	mapOutpostMsgForUserA3 := outpostfactory.ExecuteMsg{
		MapUserOutpost: &outpostfactory.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA3.FormattedAddress(), // put in UserA3 address
		},
	}
	res5, err := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA3), "--gas", "500000")
	expectedErrorMsg3 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(err, expectedErrorMsg3)
	fmt.Printf(res5.TxHash)

	// logger.LogInfo(res5)

	// To check that the mappings were done correctly.
	// Above, we should parse out the outpost address that's created for userA using the event
	// we should then assert that it's equal to 'outpostAddress', much like how we assert PubKeys are equal
	// s.Require().Equal(pubRes.PubKey.GetKey(), filetreeMsg.GetKey(), "Expected PubKey does not match the returned PubKey")

	// TODO: Make sure that users are admins of their own outposts
	// Auto-gen query client to query for the admin?

}

func TestWithFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) TestFactoryCreateOutpost() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupFactoryTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json
	// wasmd, canined := s.ChainA, s.ChainB

	// We weren't able to precompute the outpost's address at the time of creation, so we need to query for the address
	// right now
	// query by code ID and sender address? The sender being the user that executed the creation
	// The port id of the outpost should be wasm.contractAddress so can't we retrieve the address from that?

	// time.Sleep(time.Duration(10) * time.Hour)

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
