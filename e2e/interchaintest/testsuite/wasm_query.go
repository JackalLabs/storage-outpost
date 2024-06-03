package testsuite

import (
	"context"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func OutpostAddress(ctx context.Context, chain *cosmos.CosmosChain, factoryContractAddress string, userAddress string) (*wasmtypes.QuerySmartContractStateResponse, error) {
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
