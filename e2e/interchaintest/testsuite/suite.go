package testsuite

import (
	"context"
	"encoding/json"
	"os"

	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
)

type TestSuite struct {
	suite.Suite

	ChainA       *cosmos.CosmosChain
	ChainB       *cosmos.CosmosChain
	UserA        ibc.Wallet
	UserA2       ibc.Wallet
	UserB        ibc.Wallet
	ChainAConnID string
	ChainBConnID string
	dockerClient *dockerclient.Client
	Relayer      ibc.Relayer
	network      string
	logger       *zap.Logger
	ExecRep      *testreporter.RelayerExecReporter
	PathName     string
}

// SetupSuite sets up the chains, relayer, user accounts, clients, and connections
func (s *TestSuite) SetupSuite(ctx context.Context, chainSpecs []*interchaintest.ChainSpec) {
	if len(chainSpecs) != 2 {
		panic("ContractTestSuite requires exactly 2 chain specs")
	}

	t := s.T()

	s.logger = zaptest.NewLogger(t)
	s.dockerClient, s.network = interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(s.logger, chainSpecs)

	chains, err := cf.Chains(t.Name())
	s.Require().NoError(err)
	s.ChainA = chains[0].(*cosmos.CosmosChain)
	s.ChainB = chains[1].(*cosmos.CosmosChain)

	// docker run -it --rm --entrypoint echo ghcr.io/cosmos/relayer "$(id -u):$(id -g)"
	customRelayerImage := relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "", "100:1000")

	s.Relayer = interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		customRelayerImage,
	).Build(t, s.dockerClient, s.network)

	s.ExecRep = testreporter.NewNopReporter().RelayerExecReporter(t)

	s.PathName = s.ChainA.Config().Name + "-" + s.ChainB.Config().Name

	ic := interchaintest.NewInterchain().
		AddChain(s.ChainA).
		AddChain(s.ChainB).
		AddRelayer(s.Relayer, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  s.ChainA,
			Chain2:  s.ChainB,
			Relayer: s.Relayer,
			Path:    s.PathName,
		})

	s.Require().NoError(ic.Build(ctx, s.ExecRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           s.dockerClient,
		NetworkID:        s.network,
		SkipPathCreation: true,
	}))
	logger.InitLogger()

	// Map all query request types to their gRPC method paths
	//s.Require().NoError(populateQueryReqToPath(ctx, s.ChainA))
	//s.Require().NoError(populateQueryReqToPath(ctx, s.ChainB))

	// Fund user accounts on ChainA and ChainB
	const userFunds = int64(10_000_000_000)
	// users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, s.ChainA, s.ChainB)
	userASeed := "fork draw talk diagram fragile online style lecture ecology lawn " +
		"dress hat modify member leg pluck leaf depend subway grit trumpet tongue crucial stumble"
	userA, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "wasmd", userASeed, userFunds, s.ChainA)
	s.Require().NoError(err)

	userA2Seed := "cage father indicate hockey rapid wrist symbol apple impulse cradle sock pony foam " +
		"survey squirrel dial drum flavor mansion bicycle master dumb album soccer"
	userA2, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "wasmd", userA2Seed, userFunds, s.ChainA)
	s.Require().NoError(err)

	// this is the seed phrase for the danny user that appears in all of canine-chain's testing scripts
	userBSeed := "brief enhance flee chest rabbit matter chaos clever lady enable luggage arrange hint " +
		"quarter change float embark canoe chalk husband legal dignity music web"
	userB, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "jkl", userBSeed, userFunds, s.ChainB)
	s.Require().NoError(err)

	s.UserA = userA   // the primary wasmd user
	s.UserA2 = userA2 // the secondary wasmd user
	s.UserB = userB   //the jackal user

	// Generate a new IBC path
	err = s.Relayer.GeneratePath(ctx, s.ExecRep, s.ChainA.Config().ChainID, s.ChainB.Config().ChainID, s.PathName)
	s.Require().NoError(err)

	// Wait for blocks
	err = testutil.WaitForBlocks(ctx, 4, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Create new clients
	err = s.Relayer.CreateClients(ctx, s.ExecRep, s.PathName, ibc.CreateClientOptions{TrustingPeriod: "330h"})
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 4, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Wasmd should have a light client that's tracking the state of jackal-1
	lightClients0, lightClients1 := s.Relayer.GetClients(ctx, s.ExecRep, s.ChainA.Config().ChainID)

	// log first wasmd light client
	lc0, err := json.MarshalIndent(lightClients0, "", "  ")

	logger.LogInfo("Wasmd First light client is")

	if err != nil {
		// handle error
		logger.LogError("failed to marshal light client:", err)
	} else {
		logger.LogInfo(string(lc0))
	}

	// log second wasmd light client
	// note: second object being returned seems to be nil
	lc1, err := json.MarshalIndent(lightClients1, "", "  ")

	logger.LogInfo("Wasmd Second light client is")

	if err != nil {
		// handle error
		logger.LogError("failed to marshal light client:", err)
	} else {
		logger.LogInfo(string(lc1))
	}

	// Let's log jackal-1 light clients
	// jackal-1 should have a light client that's tracking the state of wasmd
	jackalLC0, jackalLC1 := s.Relayer.GetClients(ctx, s.ExecRep, s.ChainB.Config().ChainID)

	// log first jackal light client
	jlc0, err := json.MarshalIndent(jackalLC0, "", "  ")

	logger.LogInfo("jackal first light client is")

	if err != nil {
		// handle error
		logger.LogError("failed to marshal light client:", err)
	} else {
		logger.LogInfo(string(jlc0))
	}

	// log second jackal light client
	// note: second object being returned seems to be nil
	jlc1, err := json.MarshalIndent(jackalLC1, "", "  ")

	logger.LogInfo("jackal second light client is")

	if err != nil {
		// handle error
		logger.LogError("failed to marshal light client:", err)
	} else {
		logger.LogInfo(string(jlc1))
	}

	// Create a new connection
	err = s.Relayer.CreateConnections(ctx, s.ExecRep, s.PathName)
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 4, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Query for the newly created connections in wasmd
	wasmdConnections, err := s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainA.Config().ChainID)
	s.Require().NoError(err)

	// log first wasmd connection
	wc0JsonBytes, err := json.MarshalIndent(wasmdConnections[0], "", "  ")

	if err != nil {
		// handle error
		logger.LogError("failed to marshal connection:", err)
	} else {
		logger.LogInfo(string(wc0JsonBytes))
	}

	//log second wasmd connection

	wc1JsonBytes, err := json.MarshalIndent(wasmdConnections[1], "", "  ")

	if err != nil {
		// handle error
		logger.LogError("failed  to marshal connection:", err)
	} else {
		logger.LogInfo(string(wc1JsonBytes))
	}

	// localhost is always a connection since ibc-go v7.1+
	s.Require().Equal(2, len(wasmdConnections))

	// additional note: wasmd has 2 established connections but canined only has 1. Need to log.

	wasmdConnection := wasmdConnections[0]
	s.Require().NotEqual("connection-localhost", wasmdConnection.ID)
	s.ChainAConnID = wasmdConnection.ID

	// Query for the newly created connections in canined
	caninedConnections, err := s.Relayer.GetConnections(ctx, s.ExecRep, s.ChainB.Config().ChainID)
	s.Require().NoError(err)

	// localhost is always a connection since ibc-go v7.1+
	// but canine-chain is running ibc-go v4.4.2, so perhaps there's only 1 connection that isn't localhost?

	s.Require().Equal(1, len(caninedConnections))

	logger.LogInfo("The first canined connections are:")
	// log the first canined connection
	cc1JsonBytes, err := json.MarshalIndent(caninedConnections[0], "", "  ")

	if err != nil {
		// handle error
		logger.LogError("failed  to marshal connection:", err)
	} else {
		logger.LogInfo(string(cc1JsonBytes))
	}

	caninedConnection := caninedConnections[0]

	s.Require().NotEqual("connection-localhost", caninedConnection.ID)
	s.ChainBConnID = caninedConnection.ID

	// Start the relayer and set the cleanup function.
	err = s.Relayer.StartRelayer(ctx, s.ExecRep, s.PathName)
	s.Require().NoError(err)

	t.Cleanup(
		func() {
			if os.Getenv("KEEP_CONTAINERS_RUNNING") != "1" {
				err := s.Relayer.StopRelayer(ctx, s.ExecRep)
				if err != nil {
					t.Logf("an error occurred while stopping the relayer: %s", err)
				}
			} else {
				t.Logf("Skipping relayer stop due to KEEP_CONTAINERS_RUNNING flag")
			}
		},
	)
}
