//! # filetree
//!
//! Contains all the transaction msgs needed to interact with canine-chain's filetree module.
//! TODO: add remaining msgs and storage module's transaction msgs

/// Post your public key to canine-chain filetree 
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostKey {
    /// jkl address of public key owner
    #[prost(string, tag = "1")]
    pub creator: String, 
    /// user public key: hex.encode(ecies.PublicKey)
    
    // WARNING: our prost declaration was very outdated, so using
    // ::prost::alloc::string::String should now resolve. String is universal though so hopefully this won't be an issue
    #[prost(string, tag = "2")]
    pub key: String,
}
/// A successful broadcast guarantees that the key is saved on chain, so we can leave the tx response empty for now
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostKeyResponse {}