use cosmwasm_std::{entry_point, to_json_binary, Binary, Deps, Empty, Env, MessageInfo, Response};

use cosmwasm_std::{DepsMut, StdResult};
use msg::ValueResp;
use state::DATA_TO_MIGRATE;

pub mod msg;
mod state;

// Instantiate a contract, always with "Data to migrate!" as the value of "DATA_TO_MIGRATE" in state
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: Empty,
) -> StdResult<Response> {
    let save_str: String = String::from("Data to migrate!");
    DATA_TO_MIGRATE.save(deps.storage, &save_str)?;
    Ok(Response::new())
}

// Immediately return the state of "DATA_TO_MIGRATE"
#[entry_point]
pub fn query(deps: Deps, _env: Env, _msg: msg::QueryMsg) -> StdResult<Binary> {
    let data_saved: String = DATA_TO_MIGRATE.load(deps.storage)?;
    let resp: ValueResp = ValueResp {value : data_saved};
    to_json_binary(&resp)
}

// Immediately return ok response to satisfy ContractWrapper and allow testing
#[entry_point]
pub fn execute(_deps: DepsMut, _env: Env, _info: MessageInfo, _msg: Empty) -> StdResult<Response> { Ok(Response::new() )}

/* 
    WARNING: Remove this in production!
    Test the v1 contract to make sure we're properly initing the contract
*/
mod test {
    use cosmwasm_std::{Addr, Empty};
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};

    use crate::msg::{QueryMsg, ValueResp};
    use crate::{execute, instantiate, query};

    // Create a version of the v1 contract wrapped in a ContractWrapper so we can run tests on it
    fn v1() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(execute, instantiate, query);
        Box::new(contract)
    }

    // Instantiate a test contract, then see if it's storing the right initial value of "DATA_TO_MIGRATE"
    #[test]
    fn instantiate_and_check() {
        let mut app: App = App::default();
 
        let contract_id = app.store_code(v1());
     
        let contract_addr = app
            .instantiate_contract(
                contract_id,
                Addr::unchecked("sender"),
                &Empty {},
                &[],
                "v1 contract",
                None,
            )
            .unwrap();


        let resp: ValueResp = app
            .wrap()
            .query_wasm_smart(contract_addr, &QueryMsg::Value {})
            .unwrap();

        assert_eq!(resp, ValueResp { value : String::from("Data to migrate!")});
    }
}