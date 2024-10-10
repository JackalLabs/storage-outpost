/**
* This file was automatically generated by @cosmwasm/ts-codegen@0.35.7.
* DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
* and run the @cosmwasm/ts-codegen generate command to regenerate this file.
*/

import { selectorFamily } from "recoil";
import { cosmWasmClient } from "./chain";
import { InstantiateMsg, ExecuteMsg, TxEncoding, ChannelOpenInitOptions, QueryMsg, ContractState, String } from "./OutpostFactory.types";
import { OutpostFactoryQueryClient } from "./OutpostFactory.client";
type QueryClientParams = {
  contractAddress: string;
};
export const queryClient = selectorFamily<OutpostFactoryQueryClient, QueryClientParams>({
  key: "outpostFactoryQueryClient",
  get: ({
    contractAddress
  }) => ({
    get
  }) => {
    const client = get(cosmWasmClient);
    return new OutpostFactoryQueryClient(client, contractAddress);
  }
});
export const getContractStateSelector = selectorFamily<ContractState, QueryClientParams & {
  params: Parameters<OutpostFactoryQueryClient["getContractState"]>;
}>({
  key: "outpostFactoryGetContractState",
  get: ({
    params,
    ...queryClientParams
  }) => async ({
    get
  }) => {
    const client = get(queryClient(queryClientParams));
    return await client.getContractState(...params);
  }
});
export const getUserOutpostAddressSelector = selectorFamily<String, QueryClientParams & {
  params: Parameters<OutpostFactoryQueryClient["getUserOutpostAddress"]>;
}>({
  key: "outpostFactoryGetUserOutpostAddress",
  get: ({
    params,
    ...queryClientParams
  }) => async ({
    get
  }) => {
    const client = get(queryClient(queryClientParams));
    return await client.getUserOutpostAddress(...params);
  }
});