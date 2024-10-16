use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

pub use contract::ContractState;

/// The item used for storing the outpost's code id 
/// TODO: need a function to update the code ID when we release an updated version of the outpost 
pub const STATE: Item<ContractState> = Item::new("state");

/// A mapping of the user's address to the outpost address they own
/// NOTE: Should we consider calling this 'OWNER_ADDR...' given that each outpost belongs to one user and is owned by that user?
pub const USER_ADDR_TO_OUTPOST_ADDR: Map<&str, String> = Map::new("user_addr_to_outpost_addr");

/// This behaves like a lock file which ensures that users can only create an outpost for themselves
/// It's a needed work around that's caused by inter-contract executions being signed by the calling contract instead of the user's signature
pub const LOCK: Map<&str, bool> = Map::new("lock");

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