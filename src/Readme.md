# storage outpost

We can send canine-chain messages across two IBC enabled chains with the below transaction. 

### transaction

`wasmd tx wasm execute <bech32_contract_address> <JsonObject> [flags]  `

Example:

`wasmd tx wasm execute wasm1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqhs9hr8 <JsonObject> --gas 500000 --gas-prices 0.00uwsm --gas-adjustment 1.3 --from alice --keyring-backend test --output json -y --chain-id localwasm-1` 

with JsonObject as: 

```json
{
  "send_cosmos_msgs": {
    "messages": [
      {
        "stargate": {
          "type_url": "/canine_chain.filetree.MsgPostKey",
          "value": "binary data here"
        }
      }
    ]
  }
}
```

A complete golang breakdown of how a canine-chain protobuf msg is packaged for this API is as follows:

```go

package main

import (
	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

    // We declare a filetree post key msg using canine-chain's tx.pb.go
    // NOTE: In our e2e testing, we've copied canine-chain's tx.pb.go file into its own directory, as to avoid cosmos-sdk conflicts
    // between the testing package and canine-chain.
    // You can also import it from github.com/jackalLabs/canine-chain/v4/x/filetree/types 
    filetreeMsg := &filetreetypes.MsgPostKey{
        Creator: s.Contract.IcaAddress,
        Key: "Wow it really works <3",
    }

    // Declare the typeURL of the msg
    // NOTE that Jackal Labs js sdk (jackal.js) has helper functions for obtaining the typeURL of a msg
    // We simply hard code it here for demonstration purposes
    typeURL := "/canine_chain.filetree.MsgPostKey"

    // We package the msg into a protobuf 'Any' type 
    // NOTE that the cosmos ecosystem also provides JS and Rust APIs for working with the protobuf 'Any' type
	protoAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

    // We make use of the CosmosMsg API--specifically the 'Stargate' variant
    // NOTE that in our golang e2e testing, we declare a 'ContractCosmosMsg' object in a helper file which
    // maps directly to the Rust API provided by CosmWasm, which is found here:
    // [cosmos_msg.rs on GitHub](https://github.com/CosmWasm/cosmwasm/blob/main/packages/std/src/results/cosmos_msg.rs)

	cosmosMsg := ContractCosmosMsg{
		Stargate: &StargateCosmosMsg{
			TypeUrl: typeURL,
			Value:   base64.StdEncoding.EncodeToString(protoAny.Value),
		},
	}

    // Finally, we declare the 'SendCosmosMsgs' variant of ExecuteMsg, before executing the contract. 
    executeMsg := ExecuteMsg{
	SendCosmosMsgs: &ExecuteMsg_SendCosmosMsgs{
		Messages:       cosmosMsgs,
		PacketMemo:     memo,
		TimeoutSeconds: timeout,
	},
}
```

The outpost can be called by another contract. See cross-contract/contracts/outpost-user/Readme.md for a walkthrough.






