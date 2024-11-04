use cosmwasm_schema::cw_serde;
use cw_storage_plus::{Item, Map};

pub use contract::ContractState;

/// The item used for storing the outpost's code id 
pub const STATE: Item<ContractState> = Item::new("state");

/// A mapping of the user's address to the outpost address they own
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
        pub admin: String,
    }

    impl ContractState {
        /// Creates a new ContractState.
        pub fn new(storage_outpost_code_id: u64, admin: String) -> Self {
            Self {
                storage_outpost_code_id,
                admin
            }
        }
    }
}