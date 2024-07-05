use cosmwasm_schema::cw_serde;

// Type you SEND to the query entrypoint
#[cw_serde]
pub enum QueryMsg { 
    Data {}
}

#[cw_serde]
pub enum ExecuteMsg {
    SetOutpost(options::SetOutpostMsg),
}

/*
    Response containing a string
    In this test contract, only serialized then sent back from the query entrypoint
*/
#[cw_serde]
pub struct ValueResp {
    pub value: String,
}

pub mod options {
    use cosmwasm_schema::cw_serde;
    use cosmwasm_std::Addr;

    #[cw_serde]
    pub struct SetOutpostMsg {
        pub addr: Addr
    }
}