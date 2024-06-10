use cosmwasm_std::{DepsMut, MessageInfo, Response, StdResult};

use crate::state::OWNER;

pub fn instantiate(
    deps: DepsMut,
    info: MessageInfo,
) -> StdResult<Response> {
    OWNER.save(deps.storage, &info.sender)?;
    Ok(Response::new())
}

pub mod exec {
    use cosmwasm_std::{BankMsg, Coin, DepsMut, Env, MessageInfo, Response, Uint128};

    use crate::error::ContractError;
    use crate::state::OWNER;

    pub fn withdraw(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response, ContractError> {
        let owner = OWNER.load(deps.storage)?;
        if info.sender != owner {
            return Err(ContractError::Unauthorized {
                owner: owner.to_string(),
            });
        }

        let balance = deps.querier.query_all_balances(&env.contract.address)?;
        let bank_msg = BankMsg::Send {
            to_address: info.sender.to_string(),
            amount: balance,
        };

        let resp = Response::new()
            .add_message(bank_msg)
            .add_attribute("action", "withdraw")
            .add_attribute("sender", info.sender.as_str());

        Ok(resp)
    }

    pub fn withdraw_to(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        receiver: String,
        funds: Vec<Coin>,
    ) -> Result<Response, ContractError> {
        let owner = OWNER.load(deps.storage)?;
        if info.sender != owner {
            return Err(ContractError::Unauthorized {
                owner: owner.to_string(),
            });
        }

        let mut balance = deps.querier.query_all_balances(&env.contract.address)?;

        if !funds.is_empty() {
            for coin in &mut balance {
                let limit = funds
                    .iter()
                    .find(|c| c.denom == coin.denom)
                    .map(|c| c.amount)
                    .unwrap_or(Uint128::zero());

                coin.amount = std::cmp::min(coin.amount, limit);
            }
        }

        let bank_msg = BankMsg::Send {
            to_address: receiver,
            amount: balance,
        };

        let resp = Response::new()
            .add_message(bank_msg)
            .add_attribute("action", "withdraw")
            .add_attribute("sender", info.sender.as_str());

        Ok(resp)
    }
}
