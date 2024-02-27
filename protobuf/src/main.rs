use std::env;

fn main() {
    print!("Building all proto files");

    // This is just a sandbox/playground so we don't need to use a build script for now

    let out_dir = "target/debug/build/";
    env::set_var("OUT_DIR", out_dir);

    // Lets build the filetree transaction Rust files from its definition
    prost_build::compile_protos(&["src/proto_definitions/tx.proto"],
                                &["src/"]).unwrap();
}
