use cw_storage_plus::Item;
use cosmwasm_std::Addr;
use storage_outpost::outpost_helpers;

// Store a test value to migrate on the V2 contract
pub const DATA_AFTER_MIGRATION: Item<String> = Item::new("after_migration_key");

// Store a "StorageOutpostContract" wrapper, essentially an Address, on chain
pub const STORAGE_OUTPOST_CONTRACT: Item<outpost_helpers::StorageOutpostContract> = Item::new("storage_outpost_contract");