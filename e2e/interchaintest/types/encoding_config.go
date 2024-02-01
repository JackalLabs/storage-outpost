package types

import (
	// upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	// icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	feetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	v7migrations "github.com/cosmos/ibc-go/v7/modules/core/02-client/migrations/v7"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	solomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	localhost "github.com/cosmos/ibc-go/v7/modules/light-clients/09-localhost"
	simappparams "github.com/cosmos/ibc-go/v7/testing/simapp/params"

	// canine-chain types
	filetreetypes "github.com/jackalLabs/canine-chain/v3/x/filetree"
)

// SDKEncodingConfig returns the global E2E encoding config.
func SDKEncodingConfig() *sdktestutil.TestEncodingConfig {
	_, cfg := codecAndEncodingConfig()
	return &sdktestutil.TestEncodingConfig{
		InterfaceRegistry: cfg.InterfaceRegistry,
		Codec:             cfg.Marshaler, // used to be: cfg.codec,
		TxConfig:          cfg.TxConfig,
		Amino:             cfg.Amino,
	}
}

func codecAndEncodingConfig() (*codec.ProtoCodec, simappparams.EncodingConfig) {
	cfg := simappparams.MakeTestEncodingConfig()

	// ibc types
	icacontrollertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	// icahosttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	solomachine.RegisterInterfaces(cfg.InterfaceRegistry)
	v7migrations.RegisterInterfaces(cfg.InterfaceRegistry)
	transfertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	channeltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	connectiontypes.RegisterInterfaces(cfg.InterfaceRegistry)
	ibctmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	localhost.RegisterInterfaces(cfg.InterfaceRegistry)

	// all other types
	// upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1beta1.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1.RegisterInterfaces(cfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(cfg.InterfaceRegistry)
	grouptypes.RegisterInterfaces(cfg.InterfaceRegistry)
	proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	authz.RegisterInterfaces(cfg.InterfaceRegistry)

	// canine-chain types
	filetreetypes.RegisterInterfaces(cfg.InterfaceRegistry)

	cdc := codec.NewProtoCodec(cfg.InterfaceRegistry)
	return cdc, cfg
}
