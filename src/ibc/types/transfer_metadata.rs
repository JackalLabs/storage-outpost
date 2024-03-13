//! # Metadata
//!
//! This file contains the [`TransferMetadata`] struct and its methods.
//!

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Deps, IbcChannel};

use crate::types::{state::CHANNEL_STATE, ContractError};

use super::keys::ICA_VERSION;

/// TransferMetadata is the metadata of the IBC application communicated during the handshake.
#[cw_serde]
pub struct TransferMetadata {
    /// The version of the IBC application.
    pub version: String,
    /// Controller's connection id.
    pub controller_connection_id: String,
    /// Counterparty's connection id.
    pub host_connection_id: String,
}

impl TransferMetadata {
    /// Creates a new TransferMetadata
    pub fn new(
        version: String,
        controller_connection_id: String,
        host_connection_id: String,
    ) -> Self {
        Self {
            version,
            controller_connection_id,
            host_connection_id,
        }
    }
}

impl ToString for TransferMetadata {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}