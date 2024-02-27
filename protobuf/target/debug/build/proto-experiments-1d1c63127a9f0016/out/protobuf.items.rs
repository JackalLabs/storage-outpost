/// A pair of Hugo Boss shoes
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Shoe {
    #[prost(string, tag="1")]
    pub brand: std::string::String,
}
pub mod shoe {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Size {
        Small = 0,
        Medium = 1,
        Large = 2,
    }
}
