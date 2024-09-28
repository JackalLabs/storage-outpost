use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

pub use contract::ContractState;

/// The item used for storing the outpost's code id 
/// TODO: need a function to update the code ID when we release an updated version of the outpost 
pub const STATE: Item<ContractState> = Item::new("state");

/// A mapping of the user's address to the outpost address they own
/// NOTE: Should we consider calling this 'OWNER_ADDR...' given that each outpost belongs to one user and is owned by that user?
pub const FILE_NOTE: Map<&str, String> = Map::new("file_note");

mod contract {

    use super::*;

    #[cw_serde]
    pub struct ContractState {
        /// The code ID of the storage-outpost contract.
        pub storage_outpost_code_id: u64,
    }

    impl ContractState {
        /// Creates a new ContractState.
        pub fn new(storage_outpost_code_id: u64) -> Self {
            Self {
                storage_outpost_code_id,
            }
        }
    }
}