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

	// Confirm that UserA is the admin of the outpost factory
	// Jackal Labs account will be the admin of the outpost factory
	factoryContractInfoRes, infoErr := testsuite.GetContractInfo(ctx, s.ChainA, outpostfactoryContractAddr)
	s.Require().NoError(infoErr)
	s.Require().Equal(factoryContractInfoRes.Admin, s.UserA.FormattedAddress())
	logger.LogInfo(fmt.Sprintf("contract Info is: %s", factoryContractInfoRes))

	// TODO: wrapping the encoding with 'TxEncoding' is not needed anymore because 'Proto3Json'
	// is not the recommended encoding type for the ICA channel
	// we should just use an optional string
	proto3Encoding := outpostfactory.TxEncoding(encoding)

	// Create UserA's outpost
	createOutpostMsg := outpostfactory.ExecuteMsg{
		CreateOutpost: &outpostfactory.ExecuteMsg_CreateOutpost{
			Salt: nil,
			ChannelOpenInitOptions: outpostfactory.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
				TxEncoding:               &proto3Encoding,
			},
		},
	}

	res, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(createOutpostMsg), "--gas", "500000")
	s.Require().NoError(err)
	// logger.LogEvents(res.Events)
	// Confirm that UserA's outpost is administered by the factory
	outpostAddressFromEvent := logger.ParseOutpostAddressFromEvent(res.Events)
	outpostContractInfoRes, outpostInfoErr := testsuite.GetContractInfo(ctx, s.ChainA, outpostAddressFromEvent)
	s.Require().NoError(outpostInfoErr)
	s.Require().Equal(outpostContractInfoRes.Admin, outpostfactoryContractAddr)
	logger.LogInfo(fmt.Sprintf("outpostContractInfo is: %s", outpostContractInfoRes))

	// Confirm UserA is the owner of the outpost they just made
	ownerQueryRes, ownerError := testsuite.GetOutpostOwner(ctx, s.ChainA, outpostAddressFromEvent)
	s.Require().NoError(ownerError)
	var outpostOwner string
	if err := json.Unmarshal(ownerQueryRes.Data, &outpostOwner); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(s.UserA.FormattedAddress(), outpostOwner)

	// We know that the outpost we just made emitted an event showing its address
	// We can now query the mapping inside of 'outpost factory' to confirm that we mapped the correct address
	// Query for the relevant addresses to ensure everything exists
	outpostAddressFromMap, addressErr := testsuite.GetOutpostAddressFromFactoryMap(ctx, s.ChainA, outpostfactoryContractAddr, s.UserA.FormattedAddress())
	s.Require().NoError(addressErr)
	var mappedOutpostAddress string
	if err := json.Unmarshal(outpostAddressFromMap.Data, &mappedOutpostAddress); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(outpostAddressFromEvent, mappedOutpostAddress)

	// is UserA allowed to just create another outpost again? They shouldn't be able to
	_, creationErr := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(createOutpostMsg), "--gas", "500000")
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
	_, mapOutpostError := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsg), "--gas", "500000")
	expectedErrorMsg := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(mapOutpostError, expectedErrorMsg)

	// Let's get UserA2 to create a user<>outpost mapping WITHOUT creating an outpost. It will fail because no lock file exists
	// The lock file is made available to be used only when 'create_outpost' is called
	mapOutpostMsgForUserA2 := outpostfactory.ExecuteMsg{
		MapUserOutpost: &outpostfactory.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA2.FormattedAddress(),
		},
	}

	_, mapOutpostA2Error := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsgA2 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(mapOutpostA2Error, expectedErrorMsgA2)

	// UserA2 should be able to make an outpost
	makeOutpostA2Res, makeOutpostA2Error := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(createOutpostMsg), "--gas", "500000")
	s.Require().NoError(makeOutpostA2Error)

	// Confirm that A2's outpost is administered by the factory
	outpostAddressA2FromEvent := logger.ParseOutpostAddressFromEvent(makeOutpostA2Res.Events)
	outpostContractInfoA2Res, outpostInfoA2Err := testsuite.GetContractInfo(ctx, s.ChainA, outpostAddressA2FromEvent)
	s.Require().NoError(outpostInfoA2Err)
	s.Require().Equal(outpostContractInfoA2Res.Admin, outpostfactoryContractAddr)

	// Confirm that A2's address<>outpostAddress mapping was done correctly
	outpostAddressA2FromMap, addressA2Err := testsuite.GetOutpostAddressFromFactoryMap(ctx, s.ChainA, outpostfactoryContractAddr, s.UserA2.FormattedAddress())
	s.Require().NoError(addressA2Err)
	var mappedOutpostAddressA2 string
	if err := json.Unmarshal(outpostAddressA2FromMap.Data, &mappedOutpostAddressA2); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(outpostAddressA2FromEvent, mappedOutpostAddressA2)

	// If UserA2 tries to map again, lock file doesn't exist because it was consumed during the creation of their outpost
	_, Error := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA2), "--gas", "500000")
	expectedErrorMsg2 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(Error, expectedErrorMsg2)

	// UserA2 tries to maliciously create a mapping for UserA3
	mapOutpostMsgForUserA3 := outpostfactory.ExecuteMsg{
		MapUserOutpost: &outpostfactory.ExecuteMsg_MapUserOutpost{
			OutpostOwner: s.UserA3.FormattedAddress(),
		},
	}
	_, maliciousErr := s.ChainA.ExecuteContract(ctx, s.UserA2.KeyName(), outpostfactoryContractAddr, toString(mapOutpostMsgForUserA3), "--gas", "500000")
	expectedErrorMsg3 := "error in transaction (code: 5): failed to execute message; message index: 0: lock file does not exist: execute wasm contract failed"
	s.Require().EqualError(maliciousErr, expectedErrorMsg3)

	// TODO: Confirm that outpost made actually works to store pubkey

}

func TestWithFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) TestFactoryCreateOutpost() {
	ctx := context.Background()

	//TODO: Accidentally put the entire testing logic inside a function called 'SetupFactoryTestSuite'
	//rename it accordingly
	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupFactoryTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json

	time.Sleep(time.Duration(10) * time.Hour)

}

// toString converts the message to a string using json
func toString(msg any) string {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(bz)
}
