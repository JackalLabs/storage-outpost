# Outpost User

A demonstration of how any contract can call the outpost's API in the same transaction. 

To call the outpost: 

### transaction

`wasmd tx wasm execute <bech32_contract_address> <JsonObject> [flags]  `

Example:

`wasmd tx wasm execute wasm1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqhs9hr8 <JsonObject> --gas 500000 --gas-prices 0.00uwsm --gas-adjustment 1.3 --from alice --keyring-backend test --output json -y --chain-id localwasm-1` 

with JsonObject as: 

```json
{
    "call_outpost": {
        "msg": {
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
    }
}

```

A well annotated e2e test can be found at e2e/interchaintest/

The testing command is:

```

go test -v . -run TestWithContractTestSuite -testify.m TestOutpostUser -timeout 12h

```




