//! # Messages
//!
//! This module defines the messages the ICA controller contract receives.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, CosmosMsg};

/// The message to instantiate the ICA controller contract.
#[cw_serde]
pub struct InstantiateMsg {
    /// The address of the admin of the ICA application.
    /// If not specified, the sender is the admin.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub admin: Option<String>,
    /// The options to initialize the IBC channel upon contract instantiation.
    /// If not specified, the IBC channel is not initialized, and the relayer must.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub channel_open_init_options: Option<options::ChannelOpenInitOptions>,
}

/// The messages to execute the ICA controller contract.
#[cw_serde]
pub enum ExecuteMsg {
    /// CreateChannel makes the contract submit a stargate MsgChannelOpenInit to the chain.
    /// This is a wrapper around [`options::ChannelOpenInitOptions`] and thus requires the
    /// same fields.
    CreateChannel(options::ChannelOpenInitOptions),

    /// CreateTransferChannel makes the contract submit a stargate MsgChannelOpenInit to the chain.
    /// This works the same as above but opens a channel for the transfer module specifically.
    CreateTransferChannel(options::ChannelOpenInitOptions),

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
    /// WARNING: This ExecuteMsg is completely experimental and not ready for production.
    /// `SendCosmosMsgsCli` works the same as above, with the addition that canine-chain's filetree msgs can be 
    /// packed into CosmosMsgs completely from the cli
    SendCosmosMsgsCli {
        // NOTE: we can include Vec<CosmosMsg> here if needed, but if it's unused in contract.rs,
        // the chain tx to execute the contract will not parse into this enum variant 

        /// Optional memo to include in the ibc packet.
        #[serde(skip_serializing_if = "Option::is_none")]
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet. 
        /// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
    },
    /// `SendTransferMsg` sends a local token to Jackal using ICS-20 
    SendTransferMsg {
        /// Let's hard code one specific transfer msg for now just to see if it works 
        // messages: Vec<CosmosMsg>,
        /// Optional memo to include in the ibc packet.
        #[serde(skip_serializing_if = "Option::is_none")]
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet. 
        /// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
        /// The receiver of the tokens on the Jackal chain
        recipient: String,
    },
}

/// The messages to query the ICA controller contract.
/// #[cw_ownable::cw_ownable_query] NOTE: enable this macro if we want the ownership feature for the outpost 
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
    
}

/// The message to migrate this contract.
#[cw_serde]
pub struct MigrateMsg {}

/// Option types for other messages.
pub mod options {
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
