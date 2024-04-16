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

/// TODO: update struct documentation comments for v4 chain upgrade 
/// The below are just placeholder comments.
/// documentation for the filetree module can be found here:
/// https://github.com/JackalLabs/canine-chain/tree/master/x/filetree

/// Post a Files struct to chain 
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostFile {
    /// The creator and broadcaster of this message. Pass in alice's Bech32 address
    #[prost(string, tag = "1")]
    pub creator: ::prost::alloc::string::String,

    /// Hex[ hash( alice's Bech32 address )]
    #[prost(string, tag = "2")]
    pub account: ::prost::alloc::string::String,

    /// MerklePath("s")
    #[prost(string, tag = "3")]
    pub hash_parent: ::prost::alloc::string::String,

    /// Hex[ hash("home") ]
    #[prost(string, tag = "4")]
    pub hash_child: ::prost::alloc::string::String,

    /// FID
    #[prost(string, tag = "5")]
    pub contents: ::prost::alloc::string::String,

    /// string(json encoded map) with: 
    /// let c = concatenate( "v", trackingNumber, Bech32 address )
    /// map_key: hex[ hash("c") ]
    /// map_value: ECIES.encrypt( aesIV + aesKey )
    /// Note that map_key and map_value must be strings or else unmarshalling in the keeper will fail.

    #[prost(string, tag = "6")]
    pub viewers: ::prost::alloc::string::String,

    /// same as above but with c = concatenate( "e", trackingNumber, Bech32 address )
    #[prost(string, tag = "7")]
    pub editors: ::prost::alloc::string::String,

    /// UUID. This trackingNumber is one and the same as what is used in editors AND viewers map
    #[prost(string, tag = "8")]
    pub tracking_number: ::prost::alloc::string::String,
}

/// The response object 
/// let fullMerklePath = MerklePath("s/home")
/// 
/// {
///     "path": "fullMerklePath"
/// }

#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostFileResponse {
    /// fullMerklePath
    #[prost(string, tag = "1")]
    pub path: ::prost::alloc::string::String,
}