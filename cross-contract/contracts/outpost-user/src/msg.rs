use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::CosmosMsg;
use storage_outpost::outpost_helpers::ica_callback_execute;
use storage_outpost::types::msg::options::ChannelOpenInitOptions;
use storage_outpost::types::msg::ExecuteMsg as OutpostExecuteMsg;

#[cw_serde]
pub struct InstantiateMsg {
    pub storage_outpost_address: String,
}

// #[ica_callback_execute] let's implement this later
#[cw_serde]
pub enum ExecuteMsg {

    CallOutpost {
        // no need for outpost address here, it's already saved in state
        msg: OutpostExecuteMsg,
    },

    SaveNote {
        note: String, 
    },

    // Save the outpost's address to state
    // This is useful for contracts that are already on mainnet and need to migrate to enable calling the outpost
    SaveOutpost {
        address: String, 
    },

    // === Wrap the outpost's API directly for easier jjs integration ===

    // NOTE: This was first attempt. The entry point in 'contract.rs' can't immediately resolve that 
    // 'send_cosmos_msgs' is an enum variant of 'OutpostExecuteMsg' 
    // Outpost(OutpostExecuteMsg),

    // We have to just copy and paste the outpost's enum variant exactly. 
    SendCosmosMsgs {
        /// The stargate messages to convert and send to the ICA host.
        messages: Vec<CosmosMsg>,
        /// Optional memo to include in the ibc packet.
        #[serde(skip_serializing_if = "Option::is_none")]
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet. 
        /// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
    },

}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetContractState returns the contact's state.
    #[returns(crate::state::ContractState)]
    GetContractState {},
    /// GetNote returns a single note based on the user's address.
    #[returns(Option<String>)]
    GetNote { address: String },

}
