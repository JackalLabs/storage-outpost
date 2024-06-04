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
    // TODO: admin of the outpost factory should really just be info.sender but that can be passed into the outer Instantiate msg
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
        //TODO: Change this to 'CreateOutpost'
        ExecuteMsg::CreateOutpost {
            salt,
            channel_open_init_options,
        } => execute::create_outpost(deps, env, info, salt, channel_open_init_options),
        ExecuteMsg::MapUserOutpost { outpost_owner} => execute::map_user_outpost(deps, env, info, outpost_owner),
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
        QueryMsg::GetUserOutpostAddress { user_address } => to_json_binary(&query::user_outpost_address(deps, user_address)?),
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

    use crate::state::{self, CONTRACT_ADDR_TO_ICA_ID, ICA_COUNT, ICA_STATES, CALLBACK_COUNT, USER_ADDR_TO_OUTPOST_ADDR, LOCK};

    use super::*;
// WARNING: check if kv pair for user exists before creating an outpost, to prevent users from spamming this function
// A bad actor could spam this function by creating new addresses, but gas requirement means they'd be paying real $$$ 
    pub fn create_outpost(
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

        // Ensures only info.sender can create address<>outpost_address mapping when map_user_outpost is called below

        // WARNING: MUST DO: If users accidentally spam outpost creations for themselves, it becomes difficult to keep track
        // of their outpost address
        // TODO: can we use the LOCK to ensure a user can only call this function once ever?

        // We can put a check here: If the user<>outpost mapping already exists, they can't call this function

        let lock = LOCK.save(deps.storage, &info.sender.to_string(), &true);
                

        let callback = Callback {
            contract: env.contract.address.to_string(),
            // TODO: make this comment more professional
            // refer to commit 937d8f5ffa506e4d3ba34b8946b865c7da1bb4b8 to see a msg nested in the Callback
            msg: None, 
            // Even though this could be spoofed in 'map_user_outpost', that's ok because we have the lock to block
            outpost_owner: info.sender.to_string(),
        };

        let instantiate_msg = storage_outpost::types::msg::InstantiateMsg {
            // right now the owner of every outpost is the address of the outpost factory
            owner: Some(info.sender.to_string()), // WARNING: The owner should also be the info.sender, this param will be deleted soon
            admin: Some(info.sender.to_string()),
            channel_open_init_options: Some(channel_open_init_options),
            // nest the call back object here 
            callback: Some(callback),
        };

        let label = format!("storage_outpost-owned by: {}", &info.sender.to_string());

        // 'instantiate2' which has the ability to pre compute the outpost's address
        // Unsure if 'instantiate2_address' from cosmwasm-std will work on Archway so we're not doing this for now

        let cosmos_msg = ica_code.instantiate(
            instantiate_msg,
            label,
            Some(info.sender.to_string()),
        )?;

        // Idea: this owner contract can instantiate multiple outpost (ica) contracts. The CONTRACT_ADDR_TO_ICA_ID mapping
        // simply maps the contract address of the instantiated outpost to the ica_id--the ica_id being just a number that indicates
        // how many outposts have been deployed before them
        // It depends on what this owner contract is doing, but each user only needs 1 outpost to be instantiated for them
        // Why not have the mapping be 'sender address : outpost contract address'? The sender address being the user that executes this function

        let mut event = Event::new("FACTORY: create_ica_contract");
        event = event.add_attribute("info.sender", &info.sender.to_string());

        Ok(Response::new().add_message(cosmos_msg).add_event(event))
    }

    pub fn map_user_outpost(
        deps: DepsMut,
        env: Env,
        info: MessageInfo, //info.sender will be the outpost's address 
        outpost_owner: String, 
    ) -> Result<Response, ContractError> {
        // this contract can't have an owner because it needs to be called back by every outpost it instantiates 

        // Load the lock state for the outpost owner
        let lock = LOCK.may_load(deps.storage, &outpost_owner)?; // WARNING-just hardcoding for testing 

        // Check if the lock exists and is true
        if let Some(true) = lock {
            // If it does, overwrite it with false
            LOCK.save(deps.storage, &outpost_owner, &false)?;
        } else {
            // If it doesn't exist or is false, return an unauthorized error
            return Err(ContractError::MissingLock {  })
        }

    // When the factory created an outpost, the factory's address was set as the outpost owner
    // TODO: give ownership of the outpost to the outpost_owner

    USER_ADDR_TO_OUTPOST_ADDR.save(deps.storage, &outpost_owner, &info.sender.to_string())?; // again, info.sender is actually the outpost address

    let mut event = Event::new("FACTORY: map_user_outpost");
        event = event.add_attribute("info.sender", &info.sender.to_string());
        event = event.add_attribute("outpost_owner", &outpost_owner);
        event = event.add_attribute("outpost_address", &info.sender.to_string());

    Ok(Response::new().add_event(event))
    }
}

mod query {
    use crate::state::{IcaContractState, ICA_COUNT, ICA_STATES, CALLBACK_COUNT, USER_ADDR_TO_OUTPOST_ADDR};

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

    /// Returns the outpost address this user owns
    pub fn user_outpost_address(deps: Deps, user_address: String) -> StdResult<String> {
        USER_ADDR_TO_OUTPOST_ADDR.load(deps.storage, &user_address)
    }
}

#[cfg(test)]
mod tests {}
