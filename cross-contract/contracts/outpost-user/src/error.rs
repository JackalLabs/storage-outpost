use cosmwasm_std::{Instantiate2AddressError, StdError};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("ica information is not set")]
    IcaInfoNotSet {},

    #[error("Outpost already created. Outpost Address: {0}")]
    AlreadyCreated(String),
}
