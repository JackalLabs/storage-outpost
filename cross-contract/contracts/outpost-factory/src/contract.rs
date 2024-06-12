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
    // TODO: admin should be set in the wasm.Instanstiate protobuf msg
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
        ExecuteMsg::CreateOutpost {
            channel_open_init_options,
        } => execute::create_outpost(deps, env, info, channel_open_init_options),
        ExecuteMsg::MapUserOutpost { outpost_owner} => execute::map_user_outpost(deps, env, info, outpost_owner),
    }
}
// TODO: figure out which queries and states are useless
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
    pub fn create_outpost(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        channel_open_init_options: ChannelOpenInitOptions,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        // WARNING: This function is called by the user, so we cannot error:unauthorized if info.sender != admin 

        let storage_outpost_code_id = StorageOutpostCode::new(state.storage_outpost_code_id);

        // Check if key already exists and disallow multiple outpost creations 
        // If key exists, we don't care what the address is, just the mere existence of the key means an outpost was 
        // already created
            
        if let Some(value) = USER_ADDR_TO_OUTPOST_ADDR.may_load(deps.storage, &info.sender.to_string())? {
            return Err(ContractError::AlreadyCreated(value))
        }

        let _lock = LOCK.save(deps.storage, &info.sender.to_string(), &true);

        let callback = Callback {
            contract: env.contract.address.to_string(),
            // Only nest a msg if use case calls for it. 
            // Refer to commit 937d8f5ffa506e4d3ba34b8946b865c7da1bb4b8 to see a msg nested in the Callback
            msg: None, 
            // Even though this could be spoofed in 'map_user_outpost', that's ok because we have the lock to block
            outpost_owner: info.sender.to_string(),
        };

        // TODO: Admin should be outpost_owner
        let instantiate_msg = storage_outpost::types::msg::InstantiateMsg {
            // right now the owner of every outpost is the address of the outpost factory
            owner: Some(info.sender.to_string()), // WARNING: The owner should also be the info.sender, this param will be deleted soon
            admin: Some(info.sender.to_string()), // WARNING: I think the owner and admin is the user? query in E2E to double check 
            channel_open_init_options: Some(channel_open_init_options),
            // nest the call back object here 
            callback: Some(callback),
        };

        let label
         = format!("storage_outpost-owned by: {}", &info.sender.to_string());

        // 'instantiate2' which has the ability to pre compute the outpost's address
        // Unsure if 'instantiate2_address' from cosmwasm-std will work on Archway so we're not doing this for now

        let cosmos_msg = storage_outpost_code_id.instantiate(
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
            // This function can only get called if the Lock was set in 'create_outpost'
            // If it doesn't exist or is false, return an unauthorized error
            return Err(ContractError::MissingLock {  })
        }

    // When the factory created an outpost, the factory's address was set as the outpost owner
    // TODO: give ownership of the outpost to the outpost_owner

    USER_ADDR_TO_OUTPOST_ADDR.save(deps.storage, &outpost_owner, &info.sender.to_string())?; // again, info.sender is actually the outpost address

    // TODO: put the event back in
    let mut event = Event::new("FACTORY: map_user_outpost");
        event = event.add_attribute("info.sender", &info.sender.to_string());
        event = event.add_attribute("outpost_owner", &outpost_owner);
        event = event.add_attribute("outpost_address", &info.sender.to_string());

    // TODO: add an attribute to show the outpost address 
    // outpost address is info.sender because the outpost called this function 
    // DOCUMENT: note in README that a successful outpost creation shall return the address in the tx.res.attribute 
    // and a failure will throw 'AlreadyCreated' contractError

    // calling '.add_attribute' just adds a key value pair to the main wasm attribute 
    // WARNING: is it possible at all that these bytes are non-deterministic?
    // This can't be because we take from 'info.sender' which only exists if this function is called in the first place
    // This function is called only if the outpost executes the callback, otherwise the Tx was abandoned while sitting in the 
    // mem pool

    // TODO: is it more meaningful to put this attribute inside of the outpost's instantiation response?
    Ok(Response::new().add_attribute("outpost_address", &info.sender.to_string())) // this data is not propagated back up to the tx resp of the 'create_outpost' call
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




