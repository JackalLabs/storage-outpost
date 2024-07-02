use cosmwasm_std::{entry_point, to_json_binary, Binary, Deps, Empty, Env, MessageInfo, Response};

use cosmwasm_std::{DepsMut, StdResult, WasmMsg};
use cw_storage_plus::Item;
use msg::{ValueResp, ExecuteMsg};
use state::DATA_AFTER_MIGRATION;

pub mod msg;
mod state;

// Instantiate a contract, begin with "Erroneous data before migration!!!" as the value of "DATA_AFTER_MIGRATION" in state
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: Empty,
) -> StdResult<Response> {
    let save_str: String = String::from("Erroneous data before migration!!!");
    DATA_AFTER_MIGRATION.save(deps.storage, &save_str)?;
    Ok(Response::new())
}

// Immediately return the state of "DATA_AFTER_MIGRATION"
#[entry_point]
pub fn query(deps: Deps, _env: Env, _msg: msg::QueryMsg) -> StdResult<Binary> {
    let data_saved: String = DATA_AFTER_MIGRATION.load(deps.storage)?;
    let resp: ValueResp = ValueResp {value : data_saved};
    to_json_binary(&resp)
}

// Immediately return ok response to satisfy ContractWrapper and allow testing
#[entry_point]
pub fn execute(deps: DepsMut, env: Env, info: MessageInfo, msg: ExecuteMsg) -> StdResult<Response> { 
    match msg {
        ExecuteMsg::PingPong(ping_pong_msg) => execute::ping_pong(deps, env, info, ping_pong_msg),
        ExecuteMsg::PongPing(pong_ping_msg) => execute::pong_ping(deps, env, info, pong_ping_msg)
    }
}

/*
    On Chain:
        data_to_migrate : "Data to migrate!"

    In the contract:
        Item<String> = Item::new
        Item.load()
*/
#[entry_point]
pub fn migrate(deps: DepsMut, _env: Env, _msg: Empty) -> StdResult<Response> {
    // Get the old data
    const DATA_TO_MIGRATE: Item<String> = Item::new("migration_key");
    let data_from_v1_state = DATA_TO_MIGRATE.load(deps.storage)?;

    // Save it under the new name
    DATA_AFTER_MIGRATION.save(
        deps.storage,
        &data_from_v1_state,
    )?;

    // Deliver the Ok response
    Ok(Response::new())
}

mod execute {
    use cosmwasm_std::{to_json_binary, Binary, Env, MessageInfo, Response};
    use cosmwasm_std::{DepsMut, StdResult, WasmMsg};
    use crate::msg::ExecuteMsg;
    use crate::msg::options::{PingPongMsg, PongPingMsg};

    pub fn ping_pong(deps: DepsMut, _env: Env, _info: MessageInfo, msg: PingPongMsg) -> StdResult<Response> { 
        let pong_ping_message = ExecuteMsg::PongPing({
            PongPingMsg {}
        });

        let pong_msg = WasmMsg::Execute {
            contract_addr: msg.addr.to_string(),
            msg: to_json_binary(&pong_ping_message).unwrap(),
            funds: vec![],
        };

        Ok(Response::new().add_message(pong_msg))
    }

    pub fn pong_ping(deps: DepsMut, _env: Env, _info: MessageInfo, msg: PongPingMsg) -> StdResult<Response> {
        Ok(Response::new().add_attribute("pong_key", "pong_value"))
    }
}

/* 
    WARNING: Remove this in production!
    Test the v2 contract to make sure we're properly initing the contract and to do migration
*/
mod test {
    use cosmwasm_std::{Addr, Empty};
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};

    use crate::msg::options::PingPongMsg;
    use crate::msg::{ExecuteMsg, QueryMsg, ValueResp};
    use crate::state::DATA_AFTER_MIGRATION;
    use crate::{execute, instantiate, migrate, query};
    

    // Create a version of the v1 contract wrapped in a ContractWrapper so we can run tests on it
    fn v1() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(basic_migration_v1::execute, basic_migration_v1::instantiate, basic_migration_v1::query);
        Box::new(contract)
    }

    // Create a version of the v2 contract wrapped in a ContractWrapper so we can run tests on it
    fn v2() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(execute, instantiate, query).with_migrate(migrate);
        Box::new(contract)
    }

    // Instantiate a test contract, then see if it's storing the right initial value of "DATA_AFTER_MIGRATION"
    #[test]
    fn instantiate_and_check() {
        let mut app: App = App::default();
 
        let contract_id = app.store_code(v2());
     
        let contract_addr = app
            .instantiate_contract(
                contract_id,
                Addr::unchecked("sender"),
                &Empty {},
                &[],
                "v2 contract",
                None,
            )
            .unwrap();

        let resp: ValueResp = app
            .wrap()
            .query_wasm_smart(contract_addr, &QueryMsg::Data {})
            .unwrap();

        assert_eq!(resp, ValueResp { value : String::from("Erroneous data before migration!!!")});
    }
 
    #[test]
    fn migration() {
        let admin = Addr::unchecked("admin");
        //let owner = Addr::unchecked("owner");
        let sender = Addr::unchecked("sender");
    
        let mut app = App::default();
    
        let old_code_id = app.store_code(v1());
        let new_code_id = app.store_code(v2());
    
        let contract = app
        .instantiate_contract(
            old_code_id,
            sender.clone(),
            &Empty {},
            &[],
            "v1 contract",
            Some(admin.to_string())
        ).unwrap();
    
        let data_before_resp: ValueResp = app
            .wrap()
            .query_wasm_smart(contract.clone(), &QueryMsg::Data{})
            .unwrap();
        let data_before: String = data_before_resp.value;

        let _migrate_response = app
            .migrate_contract(
            (&admin).clone(),
            contract.clone(), 
            &Empty {},
            new_code_id
        ).unwrap();

        let data_after = DATA_AFTER_MIGRATION.query(&app.wrap(), contract.clone()).unwrap();
        assert_eq!(data_before, data_after);
    }

    #[test]
    fn inter_contract_call() {
        use crate::msg::{ExecuteMsg, options};

        let admin = Addr::unchecked("admin");
        let sender = Addr::unchecked("sender");
    
        let mut app = App::default();
    
        let v2_code = app.store_code(v2());
    
        let contract_1 = app
        .instantiate_contract(
            v2_code,
            sender.clone(),
            &Empty {},
            &[],
            "contract 1",
            Some(admin.to_string())
        ).unwrap();
    
        let contract_2 = app
        .instantiate_contract(
            v2_code,
            sender.clone(),
            &Empty {},
            &[],
            "contract 2",
            Some(admin.to_string())
        ).unwrap();

        let ping_pong_msg = ExecuteMsg::PingPong (
            options::PingPongMsg {
                addr : contract_2
            });
        
        let con2exe_result = app.execute_contract(
            sender,
            contract_1,
            &ping_pong_msg,
            &[]);

        let con2exe = con2exe_result.unwrap();
        let con2exe_wasm_attrib = &con2exe.events[2].attributes[1];
        
        assert_eq!("pong_key".to_string(), con2exe_wasm_attrib.key);
        assert_eq!("pong_value".to_string(), con2exe_wasm_attrib.value);
    }
}