use cosmwasm_std::{Instantiate2AddressError, StdError};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("error when computing the instantiate2 address: {0}")]
    Instantiate2AddressError(#[from] Instantiate2AddressError),

    #[error("unauthorized: key exists but only the outpost address can override its user's kv pair. expected outpost address: {expected}, but got user address: {actual}")]
    Unauthorized {
        expected: String,
        actual: String,
    },

    #[error("ica information is not set")]
    IcaInfoNotSet {},

    #[error("lock file does not exist")]
    MissingLock {},

    #[error("Outpost already created. Outpost Address: {0}")]
    AlreadyCreated(String),

    #[error("Only the factory admin can perform outpost migrations")]
    NotAdmin {},
}
