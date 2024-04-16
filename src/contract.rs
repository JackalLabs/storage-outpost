//! This module handles the execution logic of the contract.

use cosmos_sdk_proto::tendermint::p2p::packet;
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Event, Empty};
use crate::ibc::types::stargate::channel::new_ica_channel_open_init_cosmos_msg;
use crate::types::keys::{CONTRACT_NAME, CONTRACT_VERSION};
use crate::types::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};
use crate::types::state::{
    CallbackCounter, ChannelState, ContractState, CALLBACK_COUNTER, CHANNEL_STATE, STATE,
};
use crate::types::ContractError;
use crate::types::filetree::{MsgPostKey, MsgPostFile};
use crate::helpers::filetree_helpers::{hash_and_hex, merkle_helper};


/// Instantiates the contract.
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let admin = if let Some(admin) = msg.admin {
        deps.api.addr_validate(&admin)?
    } else {
        info.sender.clone()
    };

    // Make an event to log the admin
    let mut event = Event::new("logging admin");
    event = event.add_attribute("admin", admin.clone());
    event = event.add_attribute("sender", info.sender.clone());

    // Save the admin. Ica address is determined during handshake.
    STATE.save(deps.storage, &ContractState::new(admin))?;
    // Initialize the callback counter.
    CALLBACK_COUNTER.save(deps.storage, &CallbackCounter::default())?;

    // If channel open init options are provided, open the channel.
    if let Some(channel_open_init_options) = msg.channel_open_init_options {
        let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            channel_open_init_options.connection_id,
            channel_open_init_options.counterparty_port_id,
            channel_open_init_options.counterparty_connection_id,
            channel_open_init_options.tx_encoding,
        );

        Ok(Response::new().add_message(ica_channel_open_init_msg).add_event(event))
    } else {
        Ok(Response::default())
    }
}

/// Handles the execution of the contract.
#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::CreateChannel(options) => execute::create_channel(deps, env, info, options),
        ExecuteMsg::CreateTransferChannel(options) => execute::create_transfer_channel(deps, env, info, options),
        ExecuteMsg::SendCosmosMsgs {
            messages,
            packet_memo,
            timeout_seconds,
        } => {
            execute::send_cosmos_msgs(deps, env, info, messages, packet_memo, timeout_seconds)
        },
        ExecuteMsg::SendCosmosMsgsCli {
            packet_memo,
            timeout_seconds,
        } => {
            execute::send_cosmos_msgs_cli(deps, env, info, packet_memo, timeout_seconds)
        },
        ExecuteMsg::SendTransferMsg { 
            packet_memo, 
            timeout_seconds,
            recipient,
        } => {
            execute::send_transfer_msg(deps, env, info, packet_memo, timeout_seconds, recipient)
        }
    }
}

/// Handles the query of the contract.
#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetChannel {} => to_json_binary(&query::channel(deps)?),
        QueryMsg::GetCallbackCounter {} => to_json_binary(&query::callback_counter(deps)?),
        // QueryMsg::Ownership {} => to_json_binary(&cw_ownable::get_ownership(deps.storage)?),
    }
}

/// Migrate contract if version is lower than current version
#[entry_point]
pub fn migrate(deps: DepsMut, _env: Env, _msg: MigrateMsg) -> Result<Response, ContractError> {
    migrate::validate_semver(deps.as_ref())?;

    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    // If state structure changed in any contract version in the way migration is needed, it
    // should occur here

    Ok(Response::default())
}

mod execute {
    use cosmwasm_std::{coin, coins, BankMsg, CosmosMsg, IbcMsg, IbcTimeout, IbcTimeoutBlock, StdResult};
    use prost::Message;

    use crate::{
        ibc::types::{metadata::TxEncoding, packet::IcaPacketData, stargate::channel},
        types::msg::options::ChannelOpenInitOptions,
    };

    use cosmos_sdk_proto::cosmos::{bank::v1beta1::MsgSend, base::v1beta1::Coin};
    use cosmos_sdk_proto::Any;

    use super::*;

    /// Submits a stargate MsgChannelOpenInit to the chain.
    pub fn create_channel(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        options: ChannelOpenInitOptions,
    ) -> Result<Response, ContractError> {
        let mut contract_state = STATE.load(deps.storage)?;
        contract_state.verify_admin(info.sender)?;

        contract_state.enable_channel_open_init();
        STATE.save(deps.storage, &contract_state)?;

        let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            options.connection_id,
            options.counterparty_port_id,
            options.counterparty_connection_id,
            options.tx_encoding,
        );

        Ok(Response::new().add_message(ica_channel_open_init_msg))
    }

    /// Submits a stargate MsgChannelOpenInit to the chain for the transfer module
    pub fn create_transfer_channel(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        options: ChannelOpenInitOptions,
    ) -> Result<Response, ContractError> {
        let mut contract_state = STATE.load(deps.storage)?;
        contract_state.verify_admin(info.sender)?;

        contract_state.enable_channel_open_init();
        STATE.save(deps.storage, &contract_state)?;

        let transfer_channel_open_init_msg = channel::new_transfer_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            options.connection_id,
            options.counterparty_port_id,
            options.counterparty_connection_id,
        );

        Ok(Response::new().add_message(transfer_channel_open_init_msg))
    }

    /// Sends an array of [`CosmosMsg`] to the ICA host.
    #[allow(clippy::needless_pass_by_value)]
    pub fn send_cosmos_msgs(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        messages: Vec<CosmosMsg>,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<Response, ContractError> {
        // TODO: We can assert ownership of the contract later, but does this really make a difference if the
        // ica module ensures the controller 'owns' or 'is paired with' the host?
        // Ownership of the root Files{} object for filetree is also checked in canine-chain

        // cw_ownable::assert_owner(deps.storage, &info.sender)?;
        let contract_state = STATE.load(deps.storage)?;
        let ica_info = contract_state.get_ica_info()?;

        let ica_packet = IcaPacketData::from_cosmos_msgs(
            messages,
            &ica_info.encoding,
            packet_memo,
            &ica_info.ica_address,
        )?;
        let send_packet_msg = ica_packet.to_ibc_msg(&env, ica_info.channel_id, timeout_seconds)?;

        // Make a logging event 
        let mut event = Event::new("logging");

        // Add some placeholder logs
        event = event.add_attribute("log", "placeholder logs");

        Ok(Response::default().add_message(send_packet_msg).add_event(event))

    }

    // TODO: add explanation for why this function is useful to us

    /// Sends an array of [`CosmosMsg`] to the ICA host.
    #[allow(clippy::needless_pass_by_value)]
    pub fn send_cosmos_msgs_cli(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<Response, ContractError> {

        let contract_state = STATE.load(deps.storage)?;
        let ica_info = contract_state.get_ica_info()?;

        // TODO: Create Vec<CosmosMsg>
        // This isn't the final implementation of the function, we're just prototyping to see what works 

        // TODO: port this type  into src/types 
        // and pack it into a CosmosMsg

        let (parent_hash, child_hash) = merkle_helper("s/home/");

        // Declare an instance of msg_post_file
        let msg_post_file = MsgPostFile {
            // TODO: implement proper borrowing and don't use clone. Poor memory manamgement leads to high transaction gas cost
            creator: ica_info.ica_address.clone(), 
            account: hash_and_hex(&ica_info.ica_address),
            hash_parent: parent_hash,
            hash_child: child_hash,
            contents: "placeholder".to_owned(),
            viewers: "placeholder".to_owned(),
            editors: "placeholder".to_owned(),
            tracking_number: "placeholder".to_owned(),
        };

        // Let's marshal post key to bytes and pack it into stargate API 
        let encoded = msg_post_file.encode_to_vec();

        // WARNING: This is first attempt, there's a good chance we did something wrong when converting post key to bytes
        let cosmos_msg: CosmosMsg<Empty> = CosmosMsg::Stargate { 
            type_url: String::from("/canine_chain.filetree.MsgPostFile"), 
            value: cosmwasm_std::Binary(encoded.to_vec()) 
        };

        let mut messages = Vec::<CosmosMsg>::new();
        messages.insert(0, cosmos_msg);

        let ica_packet = IcaPacketData::from_cosmos_msgs(
            messages,
            &ica_info.encoding,
            packet_memo,
            &ica_info.ica_address,
        )?;
        let send_packet_msg = ica_packet.to_ibc_msg(&env, ica_info.channel_id, timeout_seconds)?;

        // Make a logging event 
        let mut event = Event::new("logging");

        // Add some placeholder logs
        event = event.add_attribute("file account", msg_post_file.account);

        Ok(Response::default().add_message(send_packet_msg).add_event(event))
    }

    /// Sends an array of [`CosmosMsg`] to the ICA host.
    #[allow(clippy::needless_pass_by_value)]
    pub fn send_transfer_msg(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
        recipient: String
    ) -> Result<Response, ContractError> {

        let contract_state = STATE.load(deps.storage)?;
        let ica_info = contract_state.get_ica_info()?;

        // let jackakl_host_address = "jkl1jvnz5jcymt3357k63vemme6vfmagvc07clvmwu0csapvvelvsm8q40cwxz".to_string();

        let timeout_block = IbcTimeoutBlock {
            revision: 10,
            height: 1000000,
        };

        // WARNING: This moves tokens that the CONTRACT owns over to the jkl address on canine-chain, NOT tokens
        // that the admin owns

        // TODO: Need to fund the contract address with tokens before calling this.
        // UX is: One Click to fund contract address, and another click to send the transfer msg?
        // Or can you bundle both messages in the web client's 'signAndBroadcast'? 

        // let cosmos_bank_msg: CosmosMsg<Empty> = CosmosMsg::Bank(BankMsg::Send { to_address: (), amount: () })
        let cosmos_msg: CosmosMsg<Empty> = CosmosMsg::Ibc(IbcMsg::Transfer { 
            channel_id: "channel-1".to_string(),
            to_address: recipient,
            amount: coin(6000, "stake"),
            timeout: IbcTimeout::with_block(timeout_block) });

        Ok(Response::default().add_message(cosmos_msg))
    }
}



mod query {
    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    /// Returns the saved channel state if it exists.
    pub fn channel(deps: Deps) -> StdResult<ChannelState> {
        CHANNEL_STATE.load(deps.storage)
    }

    /// Returns the saved callback counter.
    pub fn callback_counter(deps: Deps) -> StdResult<CallbackCounter> {
        CALLBACK_COUNTER.load(deps.storage)
    }
}

mod migrate {
    use super::*;

    pub fn validate_semver(deps: Deps) -> Result<(), ContractError> {
        let prev_cw2_version = cw2::get_contract_version(deps.storage)?;
        if prev_cw2_version.contract != CONTRACT_NAME {
            return Err(ContractError::InvalidMigrationVersion {
                expected: CONTRACT_NAME.to_string(),
                actual: prev_cw2_version.contract,
            });
        }

        let version: semver::Version = CONTRACT_VERSION.parse()?;
        let prev_version: semver::Version = prev_cw2_version.version.parse()?;
        if prev_version >= version {
            return Err(ContractError::InvalidMigrationVersion {
                expected: format!("> {}", prev_version),
                actual: CONTRACT_VERSION.to_string(),
            });
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::ibc::types::{metadata::TxEncoding, packet::IcaPacketData};

    use super::*;
    use cosmos_sdk_proto::Any;
    use cosmos_sdk_proto::cosmos::bank::v1beta1::MsgSend;
    use cosmos_sdk_proto::cosmos::base::v1beta1::Coin;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{Api, SubMsg, to_binary};

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);

        let msg = InstantiateMsg {
            admin: None,
            channel_open_init_options: None,
        };

        // Ensure the contract is instantiated successfully
        let res = instantiate(deps.as_mut(), env, info.clone(), msg).unwrap();
        assert_eq!(0, res.messages.len());

        // Ensure the admin is saved correctly
        let state = STATE.load(&deps.storage).unwrap();
        assert_eq!(state.admin, info.sender);

        // Ensure the callback counter is initialized correctly
        let counter = CALLBACK_COUNTER.load(&deps.storage).unwrap();
        assert_eq!(counter.success, 0);
        assert_eq!(counter.error, 0);
        assert_eq!(counter.timeout, 0);

        // Ensure that the contract name and version are saved correctly
        let contract_version = cw2::get_contract_version(&deps.storage).unwrap();
        assert_eq!(contract_version.contract, CONTRACT_NAME);
        assert_eq!(contract_version.version, CONTRACT_VERSION);
    }

    #[test]
    fn test_execute_send_custom_proto_ica_messages() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg {
                admin: None,
                channel_open_init_options: None
            },
        )
        .unwrap();

        // NOTE: when is the ica info set automatically? 
        STATE
        .update(&mut deps.storage, |mut state| -> StdResult<ContractState> {
            state.set_ica_info("ica_address", "channel-0", TxEncoding::Protobuf);
            Ok(state)
        })
        .unwrap();

        let contract_state = STATE.load(&deps.storage).unwrap();
        contract_state.verify_admin(info.sender.clone()).unwrap();
        let ica_info = contract_state.get_ica_info().unwrap();


        let proto_message = MsgSend {
            from_address: ica_info.ica_address,
            to_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            amount: vec![Coin {
                denom: "stake".to_string(),
                amount: "100".to_string(),
            }],
        };

        let ica_packet = IcaPacketData::from_proto_anys(
            vec![Any::from_msg(&proto_message).unwrap()],
            None,
        );
    }

}









