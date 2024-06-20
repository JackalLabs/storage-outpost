use cosmwasm_std::{entry_point, to_json_binary, Binary, Deps, Empty, Env, MessageInfo, Response};

use cosmwasm_std::{DepsMut, StdResult};
use cw_storage_plus::Item;
use msg::ValueResp;
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
pub fn execute(_deps: DepsMut, _env: Env, _info: MessageInfo, _msg: Empty) -> StdResult<Response> { Ok(Response::new() )}

#[entry_point]
pub fn migrate(deps: DepsMut, _env: Env, _msg: Empty) -> StdResult<Response> {
    // Get the old data
    const DATA_TO_MIGRATE: Item<String> = Item::new("data_to_migrate");
    let migrating_data = DATA_TO_MIGRATE.load(deps.storage)?;

    // Save it under the new name
    DATA_AFTER_MIGRATION.save(
        deps.storage,
        &migrating_data,
    )?;

    // Deliver the Ok response
    Ok(Response::new())
}

/* 
    WARNING: Remove this in production!
    Test the v2 contract to make sure we're properly initing the contract and to do migration
*/
/*
mod test {
    use cosmwasm_std::{Addr, Empty};
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};

    use crate::msg::{QueryMsg, ValueResp};
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
            .query_wasm_smart(contract_addr, &QueryMsg::Value {})
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
            .query_wasm_smart(contract.clone(), &QueryMsg::Value{})
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
}*/