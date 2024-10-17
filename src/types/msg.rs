//! # Messages
//!
//! This module defines the messages the ICA controller contract receives.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, CosmosMsg};

use super::callback::Callback;

/// The message to instantiate the ICA controller contract.
#[cw_serde]
pub struct InstantiateMsg {
    /// The address of the owner of the outpost.
    /// If not specified, the sender is the owner.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub owner: Option<String>,
    /// This inner admin really has no authority
    /// The address of the admin of the outpost.
    /// If not specified, the sender is the admin.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub admin: Option<String>,
    /// The options to initialize the IBC channel upon contract instantiation.
    /// If not specified, the IBC channel is not initialized, and the relayer must create the channel
    #[serde(skip_serializing_if = "Option::is_none")]
    pub channel_open_init_options: Option<options::ChannelOpenInitOptions>,
    /// The callback information to be used
    #[serde(skip_serializing_if = "Option::is_none")]
    pub callback: Option<Callback>

}

/// The messages to execute the ICA controller contract.
/// #[cw_ownable::cw_ownable_execute] - might need this?
#[cw_serde]
pub enum ExecuteMsg {
    /// `CreateChannel` makes the contract submit a stargate MsgChannelOpenInit to the chain.
    /// This is a wrapper around [`options::ChannelOpenInitOptions`] and thus requires the
    /// same fields. If not specified, then the options specified in the contract instantiation
    /// are used.
    CreateChannel {
        /// The options to initialize the IBC channel.
        /// If not specified, the options specified in the last channel creation are used.
        /// Must be `None` if the sender is not the owner.
        #[serde(skip_serializing_if = "Option::is_none")]
        channel_open_init_options: Option<options::ChannelOpenInitOptions>,
    },

    /// `SendCosmosMsgs` converts the provided array of [`CosmosMsg`] to an ICA tx and sends them to the ICA host.
    /// [`CosmosMsg::Stargate`] and [`CosmosMsg::Wasm`] are only supported if the [`TxEncoding`](crate::ibc::types::metadata::TxEncoding) is 
    /// [`TxEncoding::Protobuf`](crate::ibc::types::metadata::TxEncoding).
    /// 
    /// **This is the recommended way to send messages to the ICA host.**
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

    /// Save some arbitrary data to confirm migration success
    SetDataAfterMigration {
        /// Arbitary string
        data: String, 
    },
}

/// The outpost factory depends on the outpost, which causes a cyclic dependency if the outpost called
/// The outpost factory's ExecuteMsg enum.
/// We can get around this by creating the below enum which has the variant 'MapUserOutpost' from outpost-factory.
/// This is not elegant, but is simple, readable and fits our needs for now.
/// If the topology of cross contract interactions gets too complicated, creating a shared library of ExecuteMsg enums
/// or using a macro to merge two enum variants is a more elegant solution
/// 
// #[ica_callback_execute] This is Serdar's macro to merge two enum variants, we can use it later if needed.
#[cw_serde]
pub enum OutpostFactoryExecuteMsg {
    /// When the outpost is created for a user, the created outpost contract will call back the factory contract
    /// to execute the below function and map the user's address to their owned outpost
    MapUserOutpost {
        /// The user's address who will own the outpost
        outpost_owner: String, // this function is called for a specific purpose of updating a map so nothing is optional
    }
}

/// The messages to query the ICA controller contract.
#[cw_ownable::cw_ownable_query]
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetChannel returns the IBC channel info.
    #[returns(crate::types::state::ChannelState)]
    GetChannel {},
    /// GetContractState returns the contact's state.
    #[returns(crate::types::state::ContractState)]
    GetContractState {},
    /// GetCallbackCounter returns the callback counter.
    #[returns(crate::types::state::CallbackCounter)]
    GetCallbackCounter {},
    /// Return the migration data
    #[returns(String)]
    GetMigrationData {},
}

/// The message to migrate this contract.
#[cw_serde]
pub struct MigrateMsg {}

/// Option types for other messages.
pub mod options {
    use cosmwasm_std::IbcOrder;
    use super::*;
    use crate::ibc::types::{keys::HOST_PORT_ID, metadata::TxEncoding};

    /// The message used to provide the MsgChannelOpenInit with the required data.
    #[cw_serde]
    pub struct ChannelOpenInitOptions {
        /// The connection id on this chain.
        pub connection_id: String,
        /// The counterparty connection id on the counterparty chain.
        pub counterparty_connection_id: String,
        /// The counterparty port id. If not specified, [crate::ibc::types::keys::HOST_PORT_ID] is used.
        /// Currently, this contract only supports the host port.
        pub counterparty_port_id: Option<String>,
        /// TxEncoding is the encoding used for the ICA txs. If not specified, [TxEncoding::Protobuf] is used.
        pub tx_encoding: Option<TxEncoding>,
        /// The order of the channel. If not specified, [`IbcOrder::Ordered`] is used.
        /// [`IbcOrder::Unordered`] is only supported if the counterparty chain is using `ibc-go`
        /// v8.1.0 or later.
        pub channel_ordering: Option<IbcOrder>,
    }

    impl ChannelOpenInitOptions {
        /// Returns the counterparty port id.
        pub fn counterparty_port_id(&self) -> String {
            self.counterparty_port_id
                .clone()
                .unwrap_or(HOST_PORT_ID.to_string())
        }

        /// Returns the tx encoding.
        pub fn tx_encoding(&self) -> TxEncoding {
            self.tx_encoding.clone().unwrap_or(TxEncoding::Protobuf)
        }
    }
}
