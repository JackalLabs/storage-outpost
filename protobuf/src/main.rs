use std::{env, string::String};

fn main() {
    print!("Building all proto files");

    // This is just a sandbox/playground so we don't need to use a build script for now

    let out_dir = "target/debug/build/";
    env::set_var("OUT_DIR", out_dir);

    // Lets build the filetree transaction Rust files from its definition
    prost_build::compile_protos(&["src/proto_definitions/tx.proto"],
                                &["src/"]).unwrap();

    // Declare an instance of MsgPostKey
    let msg_post_key = MsgPostKey {
        creator: String::from("Alice"),
        key: String::from("Alice's Public Key"),
    };
}

/*
from: 
cosmos_sdk_proto::traits::Message,

use this:

    fn encode_length_delimited_to_vec(&self) -> Vec<u8>

*/

#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostKey {
    #[prost(string, tag = "1")]
    pub creator: String,
    #[prost(string, tag = "2")]
    pub key: String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostKeyResponse {}