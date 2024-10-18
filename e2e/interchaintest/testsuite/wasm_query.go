package testsuite

import (
	"context"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	outposttypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: If we have these functions, consider deleting e2e/interchaintest/types/outpostfactory/query.go?
func GetOutpostAddressFromFactoryMap(ctx context.Context, chain *cosmos.CosmosChain, factoryContractAddress string, userAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	// TODO: replace with query msg type in types/outpostfactory/msg.go
	queryData := map[string]interface{}{
		"get_user_outpost_address": map[string]string{
			"user_address": userAddress,
		},
	}

	queryDataBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}

	params := &wasmtypes.QuerySmartContractStateRequest{
		Address:   factoryContractAddress,
		QueryData: queryDataBytes,
	}
	return queryClient.SmartContractState(ctx, params)
}

// Get the entire factory map
func GetFactoryMap(ctx context.Context, chain *cosmos.CosmosChain, factoryContractAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	queryData := map[string]interface{}{
		"get_all_user_outpost_addresses": struct{}{},
	}

	queryDataBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}

	params := &wasmtypes.QuerySmartContractStateRequest{
		Address:   factoryContractAddress,
		QueryData: queryDataBytes,
	}
	return queryClient.SmartContractState(ctx, params)
}

func GetContractInfo(ctx context.Context, chain *cosmos.CosmosChain, contractAddress string) (*wasmtypes.QueryContractInfoResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	params := &wasmtypes.QueryContractInfoRequest{
		Address: contractAddress,
	}
	return queryClient.ContractInfo(ctx, params)
}

func GetOutpostOwner(ctx context.Context, chain *cosmos.CosmosChain, factoryContractAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	// // TODO: replace with query msg type in types/outpostfactory/msg.go
	// queryData := map[string]interface{}{
	// 	"ownership": map[string]string{},
	// }

	queryData := outposttypes.QueryMsg{
		Ownership: &struct{}{},
	}

	queryDataBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}

	params := &wasmtypes.QuerySmartContractStateRequest{
		Address:   factoryContractAddress,
		QueryData: queryDataBytes,
	}
	return queryClient.SmartContractState(ctx, params)
}

func GetNote(ctx context.Context, chain *cosmos.CosmosChain, userAddress, outpostUserAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	// // TODO: replace with query msg type in types/outpostfactory/msg.go
	// queryData := map[string]interface{}{
	// 	"ownership": map[string]string{},
	// }

	queryData := outposttypes.QueryMsg{
		GetNote: &outposttypes.GetNoteRequest{Address: userAddress},
	}

	queryDataBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}

	params := &wasmtypes.QuerySmartContractStateRequest{
		Address:   outpostUserAddress,
		QueryData: queryDataBytes,
	}
	return queryClient.SmartContractState(ctx, params)
}

func GetMigrationData(ctx context.Context, chain *cosmos.CosmosChain, outpostAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	queryData := outposttypes.QueryMsg{
		GetMigrationData: &struct{}{},
	}

	queryDataBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}

	params := &wasmtypes.QuerySmartContractStateRequest{
		Address:   outpostAddress,
		QueryData: queryDataBytes,
	}
	return queryClient.SmartContractState(ctx, params)
}
