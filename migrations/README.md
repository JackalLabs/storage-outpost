# Outpost Migrations 

Contracts to showcase migrations with the storage outpost 

## Building the Contracts

Run the following commands in the root directory of this repository:

### `basic-migration`

v1:

```text
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="devcontract_cache_burner",target=/code/contracts/burner/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.15.1 /code/migrations/basic-migration/v1

```
