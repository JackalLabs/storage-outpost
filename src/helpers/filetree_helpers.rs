//! # filetree_helpers
//!
//! helper functions to prepare a filetree msg for broadcasting
//! full documentation for filetree module here https://github.com/JackalLabs/canine-chain/tree/master/x/filetree
//! TODO: update documentation comments for v4 chain upgrade

use sha2::{Sha256, Digest};
use hex::encode;

/// hash a string input and encode to hex string
pub fn hash_and_hex(input: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(input.as_bytes());
    let hash_result = hasher.finalize();
    encode(hash_result)
}

/// full markle path
pub fn merkle_path(path: &str) -> String {
    let trim_path = path.trim_end_matches('/');
    let chunks: Vec<&str> = trim_path.split('/').collect();
    let mut total = String::new();

    for chunk in chunks {
        let mut hasher = Sha256::new();
        hasher.update(chunk.as_bytes());
        let b = encode(hasher.finalize());
        let k = format!("{}{}", total, b);

        let mut hasher1 = Sha256::new();
        hasher1.update(k.as_bytes());
        total = encode(hasher1.finalize());
    }

    total
}

/// return the merkle path of the parent and the hash_and_hex of the child
pub fn merkle_helper(arg_hashpath: &str) -> (String, String) {
    let trim_path = arg_hashpath.trim_end_matches('/');
    let chunks: Vec<&str> = trim_path.split('/').collect();

    let parent_string = chunks[..chunks.len() - 1].join("/");
    let child_string = chunks[chunks.len() - 1].to_string();
    
    let parent_hash = merkle_path(&parent_string);

    let mut hasher = Sha256::new();
    hasher.update(child_string.as_bytes());
    let child_hash = encode(hasher.finalize());

    (parent_hash, child_hash)
}

