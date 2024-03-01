use std::{any::Any, env, fs::File, string::String};
use cosmos_sdk_proto::traits::Message;
use log::{info, LevelFilter};
use simplelog::*;
use storage_outpost::types::cosmos_msg;
use cosmwasm_std::{CosmosMsg, Empty};


fn main() {
    print!("Building all proto files");

    // This is just a sandbox/playground so we don't need to use a build script for now

    let out_dir = "target/debug/build/";
    env::set_var("OUT_DIR", out_dir);

    // Lets build the filetree transaction Rust files from its definition
    prost_build::compile_protos(&["src/proto_definitions/tx.proto"],
                                &["src/"]).unwrap();

    // Init logger
    let log_file = File::create("app.log").unwrap();
    WriteLogger::init(LevelFilter::Info, Config::default(), log_file).unwrap();

    info!("preparing post key for tx");

    // Declare an instance of MsgPostKey
    let msg_post_key = MsgPostKey {
        creator: String::from("Alice"), // TODO: replace with placeholder jkl address 
        key: String::from("Alice's Public Key"),
    };

    // Let's marshal post key to bytes and pack it into stargate API 
    let encoded = msg_post_key.encode_length_delimited_to_vec();

    // WARNING: This is first attempt, there's a good chance we did something wrong when converting post key to bytes
    let cosmos_msg: CosmosMsg<Empty> = CosmosMsg::Stargate { 
        type_url: String::from("/canine_chain.filetree.MsgPostKey"), 
        value: cosmwasm_std::Binary(encoded.to_vec()) };
    
    // Call handle_cosmos_msg and log its output
    let logged_stargate_msg = handle_cosmos_msg(cosmos_msg);
    info!("{}", logged_stargate_msg);

    // This will be helpful for debugging why msgs aren't consumed by the ica host keeper
    info!("Encoded MsgPostKey length: {} bytes, starts with: {:?}",
      encoded.len(),
      &encoded[..std::cmp::min(10, encoded.len())]); // Show up to the first 10 bytes

}

fn handle_cosmos_msg<T>(cosmos_msg: CosmosMsg<T>) -> String {
    match cosmos_msg {
        CosmosMsg::Stargate { type_url, value } => {
            // Create a log message string
            let log_message = format!("Stargate type_url: {}, value: {:?}", type_url, value);
            log_message // Return the log message
        },
        _ => String::from("Encountered a non-Stargate CosmosMsg variant."),
    }
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
    // WARNING: our prost declaration was very outdated, so using
    // ::prost::alloc::string::String should now resolve. String is universal though so hopefully this won't be an issue
    #[prost(string, tag = "2")]
    pub key: String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgPostKeyResponse {}

/*
This the Go code we're trying to re create:

// NewAnyWithValue constructs a new Any packed with the value provided or
// returns an error if that value couldn't be packed. This also caches
// the packed value so that it can be retrieved from GetCachedValue without
// unmarshaling
func NewAnyWithValue(v proto.Message) (*Any, error) {
	if v == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, "Expecting non nil value to create a new Any")
	}

	bz, err := proto.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &Any{
		TypeUrl:     "/" + proto.MessageName(v),
		Value:       bz,
		cachedValue: v,
	}, nil
}
*/