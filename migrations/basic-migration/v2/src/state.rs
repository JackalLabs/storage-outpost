use cw_storage_plus::Item;

// Store a test value to migrate on the V2 contract
pub const DATA_AFTER_MIGRATION: Item<String> = Item::new("after_migration_key");