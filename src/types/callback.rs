//! # Callback
//!
//! Callback contains the address of the contract to call back 
//! along with the msg that we will ask that contract to execute

use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

/// The message to instantiate the ICA controller contract.
#[cw_serde]
pub struct Callback {
    /// The contract address that we will call back
    pub contract: String,
    /// The msg we will make the above contract execute
    pub msg: Binary,
    
}