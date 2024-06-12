use cosmwasm_schema::cw_serde;

// Type you SEND to the query service
#[cw_serde]
pub enum QueryMsg { 
    Value {}
}

/*
    Response containing a string
    In this test contract, only serialized then sent back from the query entrypoint
*/
#[cw_serde]
pub struct ValueResp {
    pub value: String,
}
