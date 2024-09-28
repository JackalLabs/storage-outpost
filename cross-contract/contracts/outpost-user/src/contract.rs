#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{ContractState, STATE};

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
        &ContractState::new(msg.storage_outpost_code_id),
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
    }
}
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
    }
}

mod execute {
    use cosmwasm_std::{Addr, BankMsg, Coin, CosmosMsg, Uint128, Event, to_json_binary};
    use storage_outpost::outpost_helpers::StorageOutpostContract;
    use storage_outpost::types::msg::ExecuteMsg as IcaControllerExecuteMsg;
    use storage_outpost::types::state::{CallbackCounter, ChannelState /*ChannelStatus*/};
    use storage_outpost::{
        outpost_helpers::StorageOutpostCode,
        types::msg::options::ChannelOpenInitOptions,
    };
    use storage_outpost::types::callback::Callback;

    use crate::state::{self};

    use super::*;

    pub fn save_note(
        deps: DepsMut,
        env: Env,
        info: MessageInfo, //info.sender will be the outpost's address 
        outpost_owner: String, 
    ) -> Result<Response, ContractError> {

    // TODO: Save the note after posting the file 

    Ok(Response::new()) 
}
}

mod query {
    use crate::state;

    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }
}


