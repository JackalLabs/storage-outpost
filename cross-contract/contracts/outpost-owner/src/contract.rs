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
        } => execute::create_ica_contract(deps, env, info, salt, channel_open_init_options)
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
    }
}

mod execute {
    use cosmwasm_std::{Addr, BankMsg, Coin, CosmosMsg, Uint128, Event};
    use storage_outpost::outpost_helpers::StorageOutpostContract;
    use storage_outpost::types::msg::ExecuteMsg as IcaControllerExecuteMsg;
    use storage_outpost::types::state::{ChannelState, /*ChannelStatus*/};
    use storage_outpost::{
        outpost_helpers::StorageOutpostCode,
        types::msg::options::ChannelOpenInitOptions,
    };

    use crate::state::{self, CONTRACT_ADDR_TO_ICA_ID, ICA_COUNT, ICA_STATES};

    use super::*;

    pub fn create_ica_contract(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        salt: Option<String>,
        channel_open_init_options: ChannelOpenInitOptions,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        if state.admin != info.sender {
            return Err(ContractError::Unauthorized {});
        }

        let ica_code = StorageOutpostCode::new(state.storage_outpost_code_id);

        let instantiate_msg = storage_outpost::types::msg::InstantiateMsg {
            owner: Some(env.contract.address.to_string()),
            admin: Some(info.sender.to_string()),
            channel_open_init_options: Some(channel_open_init_options),
            // send_callbacks_to: Some(env.contract.address.to_string()), not using for now 
        };

        let ica_count = ICA_COUNT.load(deps.storage).unwrap_or(0);

        let salt = salt.unwrap_or(env.block.time.seconds().to_string());
        let label = format!("storage_outpost-{}-{}", env.contract.address, ica_count);

        let cosmos_msg = ica_code.instantiate(
            instantiate_msg,
            label,
            Some(info.sender.to_string()),
        )?;

        // Looks like Serdar used 'instantiate2' which has the ability to pre compute the outpost's address
        // We're not using this feature for now
        // I think we still get the code id of the outpost contract when we instantiate it 
        //let initial_state = state::IcaContractState::new(contract_addr.clone());

        // ICA_STATES.save(deps.storage, ica_count, &initial_state)?;

        // CONTRACT_ADDR_TO_ICA_ID.save(deps.storage, contract_addr, &ica_count)?;

        // Make an event to log the admin
        let mut event = Event::new("cross-contract-logging");
        event = event.add_attribute("creator", info.sender.clone());

        ICA_COUNT.save(deps.storage, &(ica_count + 1))?;

        Ok(Response::new().add_message(cosmos_msg))
    }
}

mod query {
    use crate::state::{IcaContractState, ICA_COUNT, ICA_STATES};

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
}

#[cfg(test)]
mod tests {}
