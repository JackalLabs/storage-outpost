#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{ContractState, STATE};

/*
// version info for migration info
const CONTRACT_NAME: &str = "crates.io:cw-ica-owner";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
*/

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    let admin = if let Some(admin) = msg.admin {
        deps.api.addr_validate(&admin)?
    } else {
        info.sender
    };

    STATE.save(
        deps.storage,
        &ContractState::new(admin, msg.storage_outpost_code_id),
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
        ExecuteMsg::CreateIcaContract {
            salt,
            channel_open_init_options,
        } => execute::create_ica_contract(deps, env, info, salt, channel_open_init_options),
        ExecuteMsg::UpdateCallbackCount {} => execute::update_callback_count(deps, env, info),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetIcaContractState { ica_id } => {
            to_json_binary(&query::ica_state(deps, ica_id)?)
        }
        QueryMsg::GetIcaCount {} => to_json_binary(&query::ica_count(deps)?),
        QueryMsg::GetCallbackCount {} => to_json_binary(&query::callback_count(deps)?),
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

    use crate::state::{self, CONTRACT_ADDR_TO_ICA_ID, ICA_COUNT, ICA_STATES, CALLBACK_COUNT};

    use super::*;

    pub fn create_ica_contract(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        salt: Option<String>,
        channel_open_init_options: ChannelOpenInitOptions,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        // TODO: determine who is best to be admin
        // if state.admin != info.sender {
        //     return Err(ContractError::Unauthorized {});
        // }

        let ica_code = StorageOutpostCode::new(state.storage_outpost_code_id);

        let callback = Callback {
            contract: env.contract.address.to_string(),
            // WARNING: do not use unwrap()
            msg: to_json_binary(&ExecuteMsg::UpdateCallbackCount {}).ok().unwrap(), 
        };

        let instantiate_msg = storage_outpost::types::msg::InstantiateMsg {
            owner: Some(env.contract.address.to_string()),
            admin: Some(info.sender.to_string()),
            channel_open_init_options: Some(channel_open_init_options),
            // nest the call back object here 
            callback: Some(callback),
        };

        let ica_count = ICA_COUNT.load(deps.storage).unwrap_or(0);

        let callback_count = CALLBACK_COUNT.load(deps.storage).unwrap_or(0);

        let label = format!("storage_outpost-{}-{}", env.contract.address, ica_count);

        // 'instantiate2' which has the ability to pre compute the outpost's address
        // Unsure if 'instantiate2_address' from cosmwasm-std will work on Archway so we're not doing this for now

        let cosmos_msg = ica_code.instantiate(
            instantiate_msg,
            label,
            Some(info.sender.to_string()),
        )?;

        let mut sender = info.sender;

        // Idea: this owner contract can instantiate multiple outpost (ica) contracts. The CONTRACT_ADDR_TO_ICA_ID mapping
        // simply maps the contract address of the instantiated outpost to the ica_id--the ica_id being just a number that indicates
        // how many outposts have been deployed before them
        // It depends on what this owner contract is doing, but each user only needs 1 outpost to be instantiated for them
        // Why not have the mapping be 'sender address : outpost contract address'? The sender address being the user that executes this function

        // Let's just put the sender's address for now as a place holder until we figure out an alternative
        let initial_state = state::IcaContractState::new(sender.clone());

        ICA_STATES.save(deps.storage, ica_count, &initial_state)?;

        CONTRACT_ADDR_TO_ICA_ID.save(deps.storage, sender.clone(), &ica_count)?;

        ICA_COUNT.save(deps.storage, &(ica_count + 1))?;

        CALLBACK_COUNT.save(deps.storage, &(callback_count + 97))?;

        // Make an event to log the admin
        let mut event = Event::new("cross-contract-logging");
        event = event.add_attribute("creator", sender.clone());

        Ok(Response::new().add_message(cosmos_msg).add_event(event))
    }

    pub fn update_callback_count(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        // TODO: the callback count is really just a placeholder that shows our callback pattern works
        // we may delete this function once we've used the callback pattern extensively
        // if state.admin != info.sender {
        //     return Err(ContractError::Unauthorized {});
        // }

        let callback_count = CALLBACK_COUNT.load(deps.storage).unwrap_or(0);

        CALLBACK_COUNT.save(deps.storage, &(callback_count + 1))?;

        let mut event = Event::new("cross-contract-logging");
        event = event.add_attribute("creator", info.sender.clone());

        Ok(Response::new().add_event(event))
    }
}

mod query {
    use crate::state::{IcaContractState, ICA_COUNT, ICA_STATES, CALLBACK_COUNT};

    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    /// Returns the saved ICA state for the given ICA ID.
    pub fn ica_state(deps: Deps, ica_id: u64) -> StdResult<IcaContractState> {
        ICA_STATES.load(deps.storage, ica_id)
    }

    /// Returns the saved ICA count.
    pub fn ica_count(deps: Deps) -> StdResult<u64> {
        ICA_COUNT.load(deps.storage)
    }

    /// Returns the callback count
    pub fn callback_count(deps: Deps) -> StdResult<u64> {
        CALLBACK_COUNT.load(deps.storage)
    }
}

#[cfg(test)]
mod tests {}
