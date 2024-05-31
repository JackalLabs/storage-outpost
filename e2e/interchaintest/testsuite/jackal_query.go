package testsuite

import (
	"context"
	"fmt"

	treetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func QueryAllKeys(ctx context.Context, chain *cosmos.CosmosChain) { /*
		Previous endpoints:
		- /cosmos.reflection.v1.ReflectionService/FileDescriptors
		- https://api.jackalprotocol.com/jackal-dao/canine-chain/filetree/pubkeys/<Address>
		- /jackal-dao/canine-chain/filetree/pubkeys

		*** /jackal/canine-chain/filetree/pub_keys

		- Pass in PubkeysRequest object
		- Get back PubkeysResponse object

		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/types/query.pb.go
		// QueryServer is the server API for Query service.
		type QueryServer interface {
		Params(context.Context, *QueryParams) (*QueryParamsResponse, error)
		// Queries a File by address and owner_address.
		File(context.Context, *QueryFile) (*QueryFileResponse, error)
		// Queries a list of File items.
		AllFiles(context.Context, *QueryAllFiles) (*QueryAllFilesResponse, error)
		// Queries a PubKey by address.
		PubKey(context.Context, *QueryPubKey) (*QueryPubKeyResponse, error)
		// Queries a list of PubKey items.
		AllPubKeys(context.Context, *QueryAllPubKeys) (*QueryAllPubKeysResponse, error)

		Basically just copy this:
		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/client/cli/query_pubkey.go

		Struct from here "QueryPubkey":
		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/types/query.pb.go
	*/

	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return
	}

	defer grpcConn.Close()

	queryClient := treetypes.NewQueryClient(grpcConn)

	params := &treetypes.QueryAllPubKeys{}

	res, err := queryClient.AllPubKeys(context.Background(), params)
	if err != nil {
		return
	}

	fmt.Println(res)

	return
}

func QueryPubKey(ctx context.Context, chain *cosmos.CosmosChain, address string) { /*
		Previous endpoints:
		- /cosmos.reflection.v1.ReflectionService/FileDescriptors
		- https://api.jackalprotocol.com/jackal-dao/canine-chain/filetree/pubkeys/<Address>
		- /jackal-dao/canine-chain/filetree/pubkeys

		*** /jackal/canine-chain/filetree/pub_keys

		- Pass in PubkeysRequest object
		- Get back PubkeysResponse object

		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/types/query.pb.go
		// QueryServer is the server API for Query service.
		type QueryServer interface {
		Params(context.Context, *QueryParams) (*QueryParamsResponse, error)
		// Queries a File by address and owner_address.
		File(context.Context, *QueryFile) (*QueryFileResponse, error)
		// Queries a list of File items.
		AllFiles(context.Context, *QueryAllFiles) (*QueryAllFilesResponse, error)
		// Queries a PubKey by address.
		PubKey(context.Context, *QueryPubKey) (*QueryPubKeyResponse, error)
		// Queries a list of PubKey items.
		AllPubKeys(context.Context, *QueryAllPubKeys) (*QueryAllPubKeysResponse, error)

		Basically just copy this:
		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/client/cli/query_pubkey.go

		Struct from here "QueryPubkey":
		https://github.com/JackalLabs/canine-chain/blob/master/x/filetree/types/query.pb.go
	*/

	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return
	}

	defer grpcConn.Close()

	queryClient := treetypes.NewQueryClient(grpcConn)

	params := &treetypes.QueryPubKey{
		Address: address,
	}

	res, err := queryClient.PubKey(context.Background(), params)
	if err != nil {
		return
	}

	fmt.Println("PRINT OUT PUBKEY TEST:")
	fmt.Println(res)

	return
}
