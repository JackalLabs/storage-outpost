package main

import (
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

var genesisAllowICH = map[string]interface{}{
	"host_genesis_state": map[string]interface{}{
		"active_channels":     []interface{}{},
		"interchain_accounts": []interface{}{},
		"port":                "icahost",
		"params": map[string]interface{}{
			"host_enabled":   true,
			"allow_messages": []interface{}{"*"},
		},
	},
}

var chainSpecs = []*interchaintest.ChainSpec{
	// -- WASMD --
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "wasmd",
			ChainID: "wasmd-1",
			Images: []ibc.DockerImage{
				{
					Repository: "cosmwasm/wasmd", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "v0.45.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:           "wasmd",
			Bech32Prefix:  "wasm",
			Denom:         "stake",
			GasPrices:     "0.00stake",
			GasAdjustment: 1.3,
			// cannot run wasmd commands without wasm encoding
			EncodingConfig: wasm.WasmEncoding(),
			TrustingPeriod: "508h",
			NoHostMount:    false,
		},
	},

	// -- CANINED --
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "canined",
			ChainID: "jackal-1",
			Images: []ibc.DockerImage{
				{
					Repository: "biphan4/canine-chain", // FOR LOCAL IMAGE USE: Docker Image Name
					// issue: we tried both the github link and the module declaration but both cause the image not to be pulled
					Version: "0.0.7", // FOR LOCAL IMAGE USE: Docker Image Tag
					// NOTE: 0.0.7 was built using a devnet script that is hopefully compatible with this
					// entire test suite. Hopefully the genesis.json file will be found in /var and updated properly
				},
			},
			Bin:            "canined",
			Bech32Prefix:   "jkl",
			Denom:          "jkl", // do we have to use ujkl or is jkl ok?
			GasPrices:      "0.00ujkl",
			GasAdjustment:  1.3,
			TrustingPeriod: "508h",
			NoHostMount:    false,
			// ModifyGenesis:  ,
		},
	},
}
