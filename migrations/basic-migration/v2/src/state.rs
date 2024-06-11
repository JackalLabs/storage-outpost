use cosmwasm_std::Addr;
use cw_storage_plus::Item;

pub const OWNER: Item<Addr> = Item::new("owner");
pub const NEW_AFTER_MIGRATE: Item<String> = Item::new("new_after_migrate");