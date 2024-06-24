# Cross Contract

This directory contains smart contracts that will call the outpost. The aim is to develop an API that works seamlessly with jackal.js, enabling current
and future Dapps to have Jackal hot storage.

A full suite of e2e tests will be developed for these contracts, which includes contract migration to enable Jackal hot storage.

## Contracts

### `outpost-factory`

This contract is used to create a unique instance of `storage-outpost` for the user. The created outpost will then call this contract back
and map its own address as a value keyed by the user's address. 

## Building the Contracts

Run the following command in the root directory of this repository:

### `outpost-factory`

```text
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="devcontract_cache_burner",target=/code/contracts/burner/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.15.1 /code/cross-contract/contracts/outpost-factory

```
