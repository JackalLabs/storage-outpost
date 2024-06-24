# outpost factory


The outpost factory can mint instances of 'storage-outpost' and will automatically map the user's address to their outpost address using a callback mechanism.
Client developers need not understand the callback mechanism, as it's triggered by 'create_outpost'.

### transactions

The below command will create the outpost and the user_address<>contract_address mapping via callback. 

`wasmd tx wasm execute <bech32_contract_address> <JsonObject> [flags]  `


Example:

`wasmd tx wasm execute wasm1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqhs9hr8 <JsonObject> --gas 500000 --gas-prices 0.00uwsm --gas-adjustment 1.3 --from alice --keyring-backend test --output json -y --chain-id localwasm-1` 

with JsonObject as: 

```json
{
  "create_outpost": {
    "channel_open_init_options": {
      "connection_id": "connection-0",
      "counterparty_connection_id": "connection-0",
      "tx_encoding": "proto3"
    }
  }
}
```

### queries

We can query for the user's outpost address with the below command.


`wasmd query wasm contract-state smart [bech32_contract_address] <JsonObject> [flags]`

Example:

`wasmd query wasm contract-state smart wasm1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqhs9hr8 <JsonObject> --output json` 

with JsonObject as: 

```json
{
  "get_user_outpost_address": {
    "user_address": "wasm13w0fse6k9tvrq6zn68smdl6ln4s7kmh9fvq8ag"
  }
}
```

will return query response:

```json
{ "data" : "wasm1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrss5maay" }
```

The data string is the user's outpost address
