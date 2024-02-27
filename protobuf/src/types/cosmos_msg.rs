//! This module contains the helpers to convert [`CosmosMsg`] to [`cosmos_sdk_proto::Any`]
//! or a [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) string.

use cosmos_sdk_proto::{prost::EncodeError, Any};
use cosmwasm_std::{BankMsg, Coin, CosmosMsg, IbcMsg};

/*

`convert_to_proto_any` converts a [`CosmosMsg`] to a [`cosmos_sdk_proto::Any`].

`from_address` is not used in [`CosmosMsg::Stargate`]

# Errors

Returns an error on serialization failure.

# Panics

Panics if the [`CosmosMsg`] is not supported.

*/

pub fn convert_to_proto_any(msg: CosmosMsg, from_address: String) -> Result<Any, EncodeError> {
    match msg {
        CosmosMsg::Stargate { type_url, value } => Ok(Any {
            type_url,
            value: value.to_vec(),
        }),
        CosmosMsg::Bank(bank_msg) => convert_to_any::bank(bank_msg, from_address),
        _ => panic!("Unsupported CosmosMsg"),

    }
}

mod convert_to_any {
    use cosmos_sdk_proto::{
        cosmos::bank::v1beta1::MsgSend,
        ibc::{applications::transfer::v1::MsgTransfer, core::client::v1::Height},
        prost::EncodeError,
        traits::Message,
        Any,
    };

    use cosmwasm_std::BankMsg;

    pub fn bank(msg: BankMsg, from_address: String) -> Result<Any, EncodeError> {
        match msg {
            BankMsg::Send { to_address, amount } => Any::from_msg(&MsgSend {
                from_address,
                to_address,
                amount: amount
                    .into_iter()
                    .map(|coin| ProtoCoin {
                        denom: coin.denom,
                        amount: coin.amount.to_string(),
                    })
                    .collect(),
            }),
            _ => panic!("Unsupported BankMsg"),
        }
    }


}