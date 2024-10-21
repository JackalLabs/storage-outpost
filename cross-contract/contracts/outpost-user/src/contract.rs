#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{ContractState, STATE, FILE_NOTE};

/*
// version info for migration info
const CONTRACT_NAME: &str = "crates.io:outpost-factory"; // just a placeholder, not yet published
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
*/

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // NOTE: admin should be set in the wasm.Instanstiate protobuf msg
    // Setting it into contract state is actually useless when wasmd checks for migration permissions
    
    // This contract cannot have an owner because it needs to be called by all users to map their outpost
    // We have a check below which ensures that users cannot call 'map' twice 

    STATE.save(
        deps.storage,
        &ContractState::new(msg.storage_outpost_address),
    )?;
    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SaveNote { note} => execute::save_note(deps, env, info, note),
        ExecuteMsg::CallOutpost { msg } => execute::call_outpost(deps, env, info, msg),
        ExecuteMsg::SaveOutpost { address } => execute::save_outpost(deps, env, info, address),

    }
}
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetNote { address } => to_json_binary(&query::query_note_by_address(deps, address)?),


    }
}

mod execute {
    use cosmwasm_std::{Addr, BankMsg, Coin, CosmosMsg, Uint128, Event, to_json_binary};
    use storage_outpost::outpost_helpers::StorageOutpostContract;
    use storage_outpost::types::msg::ExecuteMsg as OutpostExecuteMsg;
    use storage_outpost::types::state::{CallbackCounter, ChannelState /*ChannelStatus*/};
    use storage_outpost::{
        outpost_helpers::StorageOutpostCode,
        types::msg::options::ChannelOpenInitOptions,
    };
    use storage_outpost::types::callback::Callback;

    use crate::state::{self, FILE_NOTE};

    use super::*;

    pub fn save_note(
        deps: DepsMut,
        env: Env,
        info: MessageInfo, 
        note: String, 
    ) -> Result<Response, ContractError> {

    // Use the sender's address as the key to store the note
    let caller_addr = info.sender.as_str();
    
    // Update the FILE_NOTE map with the note
    FILE_NOTE.save(deps.storage, caller_addr, &note)?;

    // Emit an event to confirm the note has been saved
    let response = Response::new()
        .add_attribute("action", "save_note")
        .add_attribute("note", note)
        .add_attribute("sender", caller_addr);
    Ok(response)
    }

    pub fn call_outpost(
        deps: DepsMut,
        env: Env,
        info: MessageInfo, //info.sender will be the outpost's address 
        outpost_msg: OutpostExecuteMsg, 
    ) -> Result<Response, ContractError> {

        let state = STATE.load(deps.storage)?;
        // WARNING: This function is called by the user, so we cannot error:unauthorized if info.sender != admin 

        let storage_outpost_address = state.storage_outpost_address;

        // Convert the bech32 string back to 'Addr' type before passing to the canine_bindings helper API
        let error_msg: String = String::from("Bindings contract address is not a valid bech32 address. Conversion back to addr failed");
        let outpost_contract = StorageOutpostContract::new(deps.api
            .addr_validate(&storage_outpost_address)
            .expect(&error_msg));

        let outpost_msg = outpost_contract.call(outpost_msg)?;

    // TODO: Save the note after posting the file 
    Ok(Response::new().add_message(outpost_msg)) 
    }

    pub fn save_outpost(
        deps: DepsMut,
        env: Env,
        info: MessageInfo, //info.sender will be the outpost's address 
        address: String, 
    ) -> Result<Response, ContractError> {

        let mut state = STATE.load(deps.storage)?;
        // WARNING: This function is called by the user, so we cannot error:unauthorized if info.sender != admin 

        state.storage_outpost_address = address;

        // Save the updated state back to storage
        STATE.save(deps.storage, &state)?;

        // Return a success response
        Ok(Response::new()
        .add_attribute("action", "update_outpost_address")
        .add_attribute("new_outpost_address", state.storage_outpost_address))
    }
}

mod query {
    use cosmwasm_std::to_binary;

    use crate::state;

    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    pub fn query_note_by_address(deps: Deps, address: String) -> StdResult<String> {
        FILE_NOTE.load(deps.storage, &address)
    }
}


