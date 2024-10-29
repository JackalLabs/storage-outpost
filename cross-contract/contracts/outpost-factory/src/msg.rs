use cosmwasm_schema::{cw_serde, QueryResponses};
use storage_outpost::types::msg::options::ChannelOpenInitOptions;

#[cw_serde]
pub struct InstantiateMsg {
    pub storage_outpost_code_id: u64,
}

#[cw_serde]
pub enum ExecuteMsg {
    CreateOutpost {
        channel_open_init_options: ChannelOpenInitOptions,
    },
    // When the outpost is created for a user, the created outpost contract will call back this factory contract
    // to execute the below function and map the user's address to their owned outpost
    MapUserOutpost {
        outpost_owner: String, // This function is called for a specific purpose of updating a map so we don't make the params optional 
    },
    // Migrations thoroughly tested
    MigrateOutpost {
        outpost_owner: String, 
        new_outpost_code_id: String,
    }
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetContractState returns the contact's state.
    #[returns(crate::state::ContractState)]
    GetContractState {},
    /// GetUserOutpostAddress returns the outpost address owned by the given user address
    #[returns(String)]
    GetUserOutpostAddress { user_address: String},
    /// GetAllUserOutpostAddresses returns all user-to-outpost mappings.
    #[returns(Vec<(String, String)>)]
    GetAllUserOutpostAddresses {},
}
