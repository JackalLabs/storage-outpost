use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

pub use contract::ContractState;
pub use ica::{IcaContractState, IcaState};

/// The item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");
/// The map used to store the state of the storage-outpost contracts.
pub const ICA_STATES: Map<u64, IcaContractState> = Map::new("ica_states");
/// The item used to store the count of the storage-outpost contracts.
pub const ICA_COUNT: Item<u64> = Item::new("ica_count");
/// The item used to map contract addresses to ICA IDs.
pub const CONTRACT_ADDR_TO_ICA_ID: Map<Addr, u64> = Map::new("contract_addr_to_ica_id");
/// NOTE: this is just temporary usage. If we can confirm the callback works, we'll create
/// a mapping of user_bech32_addresses <> outpost_contract_addresses
/// 
/// The item used to track if the callback was successful
pub const CALLBACK_COUNT: Item<u64> = Item::new("call_back_count");

/// A mapping of the user's address to the outpost address they own
/// NOTE: Should we consider calling this 'OWNER_ADDR...' given that each outpost belongs to one user and is owned by that user?
pub const USER_ADDR_TO_OUTPOST_ADDR: Map<&str, String> = Map::new("user_addr_to_outpost_addr");

/// This behaves like a lock file which ensures that users can only create an outpost for themselves
/// It's a needed work around that's caused by inter-contract executions being signed by the calling contract instead of the user's signature
pub const LOCK: Map<&str, bool> = Map::new("lock");

mod contract {
    use crate::ContractError;

    use super::*;

    /// ContractState is the state of the IBC application.
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

mod ica {
    use storage_outpost::{ibc::types::metadata::TxEncoding, types::state::ChannelState};

    use super::*;

    /// IcaContractState is the state of the storage-outpost contract.
    #[cw_serde]
    pub struct IcaContractState {
        pub contract_addr: Addr,
        pub ica_state: Option<IcaState>,
    }

    /// IcaState is the state of the ICA.
    #[cw_serde]
    pub struct IcaState {
        pub ica_id: u64, // curious what this is used for?
        pub ica_addr: String,
        pub tx_encoding: TxEncoding,
        pub channel_state: ChannelState,
    }

    impl IcaContractState {
        /// Creates a new [`IcaContractState`].
        pub fn new(contract_addr: Addr) -> Self {
            Self {
                contract_addr,
                ica_state: None,
            }
        }
    }

    impl IcaState {
        /// Creates a new [`IcaState`].
        pub fn new(
            ica_id: u64,
            ica_addr: String,
            tx_encoding: TxEncoding,
            channel_state: ChannelState,
        ) -> Self {
            Self {
                ica_id,
                ica_addr,
                tx_encoding,
                channel_state,
            }
        }
    }
}
