package main

import (
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

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
					Version:    "v0.41.0",        // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:           "wasmd",
			Bech32Prefix:  "wasm",
			Denom:         "stake",
			GasPrices:     "0.00stake",
			GasAdjustment: 1.3,
			// cannot run wasmd commands without wasm encoding
			EncodingConfig:         wasm.WasmEncoding(),
			TrustingPeriod:         "508h",
			NoHostMount:            false,
			UsingNewGenesisCommand: true,
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
					Repository: "github.com/JackalLabs/canine-chain", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "v3.0.2",                             // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:                    "canined",
			Bech32Prefix:           "jkl",
			Denom:                  "jkl", // do we have to use ujkl or is jkl ok?
			GasPrices:              "0.00ujkl",
			GasAdjustment:          1.3,
			TrustingPeriod:         "508h",
			NoHostMount:            false,
			UsingNewGenesisCommand: false,
		},
	},
}
