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
        ExecuteMsg::CreateOutpost {
            channel_open_init_options,
        } => execute::create_outpost(deps, env, info, channel_open_init_options),
        ExecuteMsg::MapUserOutpost { outpost_owner} => execute::map_user_outpost(deps, env, info, outpost_owner),
        ExecuteMsg::MigrateOutpost { outpost_owner, new_outpost_code_id } => execute::migrate_outpost(deps, env, info, outpost_owner, new_outpost_code_id),
    }
}
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetUserOutpostAddress { user_address } => to_json_binary(&query::user_outpost_address(deps, user_address)?),
        QueryMsg::GetAllUserOutpostAddresses {  } => to_json_binary(&query::get_all_user_outpost_addresses(deps)?),
    }
}

mod execute {
    use cosmwasm_std::{Addr, BankMsg, Coin, CosmosMsg, Uint128, Event, to_json_binary};
    use storage_outpost::outpost_helpers::StorageOutpostContract;
    use storage_outpost::types::msg::ExecuteMsg as IcaControllerExecuteMsg;
    use storage_outpost::types::msg::MigrateMsg;
    use storage_outpost::types::state::{CallbackCounter, ChannelState /*ChannelStatus*/};
    use storage_outpost::{
        outpost_helpers::StorageOutpostCode,
        types::msg::options::ChannelOpenInitOptions,
    };
    use storage_outpost::types::callback::Callback;
    use serde_json_wasm::from_str;

    use crate::state::{self, USER_ADDR_TO_OUTPOST_ADDR, LOCK};

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

        // Whoever calls this function will save a lock for themselves, which can only be used once. 
        // 'map_user_outpost' executed via callback from the instantiated outpost, can only run if this lock exists
        let _lock = LOCK.save(deps.storage, &info.sender.to_string(), &true);

        let callback = Callback {
            contract: env.contract.address.to_string(),
            // Only nest a msg if use case calls for it. 
            // Refer to commit 937d8f5ffa506e4d3ba34b8946b865c7da1bb4b8 to see a msg nested in the Callback
            msg: None, 
            // Even though this could be spoofed in 'map_user_outpost', that's ok because we have the lock to block
            outpost_owner: info.sender.to_string(),
        };

        let instantiate_msg = storage_outpost::types::msg::InstantiateMsg {
            // NOTE: The user that executes this function is both the owner and the admin of the outpost they create
            owner: Some(info.sender.to_string()), 
            admin: Some(env.contract.address.to_string()), // Factory address is now admin of outpost
            channel_open_init_options: Some(channel_open_init_options),
            callback: Some(callback),
        };

        let label
         = format!("storage_outpost-owned by: {}", &info.sender.to_string());

        // 'instantiate2' has the ability to pre compute the outpost's address
        // Unsure if 'instantiate2_address' from cosmwasm-std will work on Archway so we're not doing this for now

        let cosmos_msg = storage_outpost_code_id.instantiate(
            instantiate_msg,
            label,
            Some(env.contract.address.to_string()), // Factory address is now admin of outpost
        )?;

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
        let lock = LOCK.may_load(deps.storage, &outpost_owner)?; 

        // Check if the lock exists and is true
        if let Some(true) = lock {
            // If it does, overwrite it with false
            LOCK.save(deps.storage, &outpost_owner, &false)?;
        } else {
            // This function can only get called if the Lock was set in 'create_outpost'
            // If it doesn't exist or is false, return an unauthorized error
            return Err(ContractError::MissingLock {  })
        }

    USER_ADDR_TO_OUTPOST_ADDR.save(deps.storage, &outpost_owner, &info.sender.to_string())?; // again, info.sender is actually the outpost address

    let mut event = Event::new("FACTORY:map_user_outpost");
    event = event.add_attribute("info.sender", &info.sender.to_string());

    Ok(Response::new().add_event(event)) // NOTE: this event is not propagated back up to the tx resp of the 'create_outpost' call
    }

    pub fn migrate_outpost(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        outpost_owner: String,
        new_outpost_code_id: String,
    ) -> Result<Response, ContractError> {
        // TODO: Migration is done via a cross contract call, which means this factory address will make the call
        // Given that the factory is the admin of all outposts, every migration call will succeed
        // We don't want this to be called by any random person though, so let's save an admin 
        // when the factory is instantiated



        // Find the owner's outpost address
        let outpost_address = USER_ADDR_TO_OUTPOST_ADDR.load(deps.storage, &outpost_owner)?;

        let error_msg: String = String::from("Outpost contract address is not a valid bech32 address. Conversion back to addr failed");

        // Call the outpost's helper API 
        let storage_outpost_code = StorageOutpostContract::new(deps.api.addr_validate(&outpost_address).expect(&error_msg));

        // The outpost's migrate entry point is just '{}'
        let migrate_msg = MigrateMsg {};

        let cast_err: String = String::from("Could not cast new outpost code to u64");
        let new_outpost_code_id_u64 = new_outpost_code_id.parse::<u64>().expect(&cast_err);

        let cosmos_msg = storage_outpost_code.migrate(
            migrate_msg,
            new_outpost_code_id_u64,
        )?;

        let mut event = Event::new("Migration: success");

        // Optimistically make sure the factory knows the new code id of the outpost 
        let mut state = STATE.load(deps.storage)?;

        // A new code id will trigger many migrations, so we only need to save it once
        if state.storage_outpost_code_id != new_outpost_code_id_u64 {

            state.storage_outpost_code_id = new_outpost_code_id_u64;
            STATE.save(deps.storage, &state)?;

        }
        
        Ok(Response::new().add_message(cosmos_msg).add_event(event)) 
    }
}

mod query {
    use crate::state::{USER_ADDR_TO_OUTPOST_ADDR};

    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    /// Returns the outpost address this user owns
    pub fn user_outpost_address(deps: Deps, user_address: String) -> StdResult<String> {
        USER_ADDR_TO_OUTPOST_ADDR.load(deps.storage, &user_address)
    }

    // Get every key value pair from the 'USER_ADDR_TO_OUTPOST_ADDR' map
    pub fn get_all_user_outpost_addresses(deps: Deps) -> StdResult<Vec<(String, String)>> {
        // Create a vector to store all entries
        let mut all_entries = Vec::new();
    
        // Use the prefix_range function to iterate over all key-value pairs in the map
        let pairs = USER_ADDR_TO_OUTPOST_ADDR
            .range(deps.storage, None, None, cosmwasm_std::Order::Ascending);
    
        // Collect each key-value pair
        for pair in pairs {
            let (key, value) = pair?;
            all_entries.push((key.to_string(), value));
        }
    
        Ok(all_entries)
    }
}

#[cfg(test)]
mod tests {}




