package testsuite

import (
	"context"

	treetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
	Note:
		While writing these functions, I considered creating a helper to eliminate this repeated code:
			grpcConn, err := grpc.Dial(
				chain.GetHostGRPCAddress(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return nil, err
			}
			defer grpcConn.Close()
			queryClient := treetypes.NewQueryClient(grpcConn)

		However, the helper function:
			1. Still required passing the chain.
			2. Still required an error check.
			3. Still required deferring the connection closure outside the function.
			4. Required creating the queryClient afterwards due to the defer statement.

		As a result, the helper function only slightly reduced the code for the grpc.Dial call.
*/

func AllPubKeys(ctx context.Context, chain *cosmos.CosmosChain) (*treetypes.QueryAllPubKeysResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := treetypes.NewQueryClient(grpcConn)
	params := &treetypes.QueryAllPubKeys{}
	return queryClient.AllPubKeys(ctx, params)
}

func PubKey(ctx context.Context, chain *cosmos.CosmosChain, address string) (*treetypes.QueryPubKeyResponse, error) {
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()
	queryClient := treetypes.NewQueryClient(grpcConn)
	params := &treetypes.QueryPubKey{Address: address}
	return queryClient.PubKey(ctx, params)
}
