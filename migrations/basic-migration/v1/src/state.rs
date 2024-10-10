use cw_storage_plus::Item;

// Store a test value to migrate on the V2 contract
pub const DATA_TO_MIGRATE: Item<String> = Item::new("data_to_migrate");