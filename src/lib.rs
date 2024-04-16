#![doc = include_str!("../README.md")]
#![deny(missing_docs)]

#[cfg(not(feature = "library"))]
pub mod contract;
pub mod ibc;
pub mod types;
pub mod helpers;