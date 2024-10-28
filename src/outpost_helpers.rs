//! This file contains helper functions for working with this contract from
//! external contracts.

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{
    to_json_binary, Addr, Api, CosmosMsg, Env, QuerierWrapper, StdError,
    StdResult, WasmMsg,
};

use crate::types::{msg, state};

// I think we can still use Serdar's derive macro to merge two enum variants
pub use cw_ica_controller_derive::ica_callback_execute; // re-export for use in macros

/// `StorageOutpostContract` is a wrapper around Addr that provides helpers
/// for working with this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct StorageOutpostContract(pub Addr);

/// `StorageOutpostCode` is a wrapper around u64 that provides helpers for
/// initializing this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct StorageOutpostCode(pub u64);

impl StorageOutpostContract {
    /// new creates a new [`StorageOutpostContract`]
    #[must_use]
    pub const fn new(addr: Addr) -> Self {
        Self(addr)
    }

    /// addr returns the address of the contract
    #[must_use]
    pub fn addr(&self) -> Addr {
        self.0.clone()
    }

    /// call creates a [`WasmMsg::Execute`] message targeting this contract,
    ///
    /// # Errors
    ///
    /// This function returns an error if the given message cannot be serialized
    pub fn call(&self, msg: impl Into<msg::ExecuteMsg>) -> StdResult<CosmosMsg> {
        let msg = to_json_binary(&msg.into())?;
        Ok(WasmMsg::Execute {
            contract_addr: self.addr().into(),
            msg,
            funds: vec![],
        }
        .into())
    }

    /// `query_channel` queries the [`state::ChannelState`] of this contract
    ///
    /// # Errors
    ///
    /// This function returns an error if the query fails
    pub fn query_channel(&self, querier: QuerierWrapper) -> StdResult<state::ChannelState> {
        querier.query_wasm_smart(self.addr(), &msg::QueryMsg::GetChannel {})
    }

    /// `query_state` queries the [`state::ContractState`] of this contract
    ///
    /// # Errors
    ///
    /// This function returns an error if the query fails
    pub fn query_state(&self, querier: QuerierWrapper) -> StdResult<state::ContractState> {
        querier.query_wasm_smart(self.addr(), &msg::QueryMsg::GetContractState {})
    }

    /// `update_admin` creates a [`WasmMsg::UpdateAdmin`] message targeting this contract
    pub fn update_admin(&self, admin: impl Into<String>) -> CosmosMsg {
        WasmMsg::UpdateAdmin {
            contract_addr: self.addr().into(),
            admin: admin.into(),
        }
        .into()
    }



    /// `migrate` creates a [`WasmMsg::Migrate`] message targeting this contract
    ///
    /// # Errors
    ///
    /// This function returns an error if the given message cannot be serialized
    pub fn migrate(
        &self,
        msg: impl Into<msg::MigrateMsg>,
        new_code_id: u64,
    ) -> StdResult<CosmosMsg> {
        let msg = to_json_binary(&msg.into())?;
        Ok(WasmMsg::Migrate {
            contract_addr: self.addr().into(),
            new_code_id,
            msg,
        }
        .into())
    }

}


impl StorageOutpostCode {
    /// new creates a new [`CwIcaControllerCode`]
    #[must_use]
    pub const fn new(code_id: u64) -> Self {
        Self(code_id)
    }

    /// `code_id` returns the code id of this code
    #[must_use]
    pub const fn code_id(&self) -> u64 {
        self.0
    }

    /// `instantiate` creates a [`WasmMsg::Instantiate`] message targeting this code
    ///
    /// # Errors
    ///
    /// This function returns an error if the given message cannot be serialized
    pub fn instantiate(
        &self,
        msg: impl Into<msg::InstantiateMsg>,
        label: impl Into<String>,
        admin: Option<impl Into<String>>,
    ) -> StdResult<CosmosMsg> {
        let msg = to_json_binary(&msg.into())?;
        Ok(WasmMsg::Instantiate {
            code_id: self.code_id(),
            msg, // the callback object will be nested here 
            funds: vec![],
            label: label.into(),
            admin: admin.map(Into::into),
        }
        .into())
    }

    // We aren't implementing `instantiate2` for now 
    
}
