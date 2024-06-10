use cosmwasm_schema::cw_serde;
use cosmwasm_std::Coin;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecMsg {
    Withdraw {},
    WithdrawTo {
        receiver: String,
        #[serde(default)]
        funds: Vec<Coin>,
    },
}

#[cw_serde]
pub struct ValueResp {
    pub value: u64,
}
