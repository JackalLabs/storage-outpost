# Cross Contract

This directory contains smart contracts that will call the outpost. The aim is to develop an API that works seamlessly with jackal.js, enabling current
and future Dapps to have Jackal hot storage.

A full suite of e2e tests will be developed for these contracts, which includes contract migration to enable Jackal hot storage.

## Contracts

### `outpost-owner`

This contract is used to test how the `storage-outpost` could be controlled by another smart contract.

## Building the Contracts

Run the following commands in the root directory of this repository:

### `outpost-owner`

```text
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="devcontract_cache_burner",target=/code/contracts/burner/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.15.1 /code/cross-contract/contracts/outpost-owner

```
