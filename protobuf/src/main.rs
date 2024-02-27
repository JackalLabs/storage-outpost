
fn main() {
    print!("Hello World!");
    prost_build::compile_protos(&["src/items.proto"],
                                &["src/"]).unwrap();
}