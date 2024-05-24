use cosmwasm_schema::{cw_serde, QueryResponses};
use storage_outpost::outpost_helpers::ica_callback_execute;
use storage_outpost::types::msg::options::ChannelOpenInitOptions;

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
    pub storage_outpost_code_id: u64,
}

// #[ica_callback_execute] let's implement this later
#[cw_serde]
pub enum ExecuteMsg {
    CreateIcaContract {
        salt: Option<String>,
        channel_open_init_options: ChannelOpenInitOptions,
    },
    UpdateCallbackCount {

    }
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetContractState returns the contact's state.
    #[returns(crate::state::ContractState)]
    GetContractState {},
    /// GetIcaState returns the ICA state for the given ICA ID.
    #[returns(crate::state::IcaContractState)]
    GetIcaContractState { ica_id: u64 },
    /// GetIcaCount returns the number of ICAs.
    #[returns(u64)]
    GetIcaCount {},
    /// GetCallBackCount returns the count in the callback object.
    #[returns(u64)]
    GetCallbackCount {},
}
