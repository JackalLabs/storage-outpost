use cosmwasm_schema::{cw_serde, QueryResponses};
use storage_outpost::outpost_helpers::ica_callback_execute;
use storage_outpost::types::msg::options::ChannelOpenInitOptions;

#[cw_serde]
pub struct InstantiateMsg {
    pub storage_outpost_code_id: u64,
}

// #[ica_callback_execute] let's implement this later
#[cw_serde]
pub enum ExecuteMsg {

    SaveNote {
        note: String, 
    }
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetContractState returns the contact's state.
    #[returns(crate::state::ContractState)]
    GetContractState {},
}
