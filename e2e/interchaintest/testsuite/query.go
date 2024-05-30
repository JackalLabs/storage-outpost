package testsuite

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/protobuf/proto"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

var queryReqToPath = make(map[string]string)

func populateQueryReqToPath(ctx context.Context, chain *cosmos.CosmosChain) error {
	resp, err := queryFileDescriptors(ctx, chain)
	if err != nil {
		return err
	}

	for _, fileDescriptor := range resp.Files {
		for _, service := range fileDescriptor.GetService() {
			// Skip services that are annotated with the "cosmos.msg.v1.service" option.
			if ext := pb.GetExtension(service.GetOptions(), msgv1.E_Service); ext != nil && ext.(bool) {
				continue
			}

			for _, method := range service.GetMethod() {
				// trim the first character from input which is a dot
				queryReqToPath[method.GetInputType()[1:]] = fileDescriptor.GetPackage() + "." + service.GetName() + "/" + method.GetName()
			}
		}
	}

	return nil
}

// TEMPORARY FUNCTION
// For debugging, prints the content of the queryReqToPath variable
func queryPrinter(m map[string]string) {
	fmt.Println("Contents of queryReqToPath:")
	for key, value := range m {
		fmt.Println("Key:", key, ", Value:", value)
	}
}

// Queries the chain with a query request and deserializes the response to T
func GRPCQuery[T any](ctx context.Context, chain *cosmos.CosmosChain, req proto.Message, opts ...grpc.CallOption) (*T, error) {
	path, ok := queryReqToPath[proto.MessageName(req)]
	if !ok {
		fmt.Println("MILO, PRINTING QUERYREQ:")
		queryPrinter(queryReqToPath)
		return nil, fmt.Errorf("no path found for %s", proto.MessageName(req))
	}

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	resp := new(T)
	err = grpcConn.Invoke(ctx, path, req, resp, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func queryFileDescriptors(ctx context.Context, chain *cosmos.CosmosChain) (*reflectionv1.FileDescriptorsResponse, error) {
	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	resp := new(reflectionv1.FileDescriptorsResponse)

	/*
	 Replaced the constant ReflectionService_FileDescriptors_FullMethodName with "/cosmos.reflection.v1.ReflectionService/FileDescriptors".
	 Our cosmossdk version (v0.3.1) lacks this constant. Constant definition here: https://pkg.go.dev/cosmossdk.io/api@v0.7.5/cosmos/reflection/v1
	*/

	/*
		Previous endpoints:
			- /cosmos.reflection.v1.ReflectionService/FileDescriptors
			- https://api.jackalprotocol.com/jackal-dao/canine-chain/filetree/pubkeys/<Address>
			- /jackal-dao/canine-chain/filetree/pubkeys
	*/
	var test_endpoint = "/jackal-dao/canine-chain/filetree/pubkeys"
	err = grpcConn.Invoke(
		ctx, test_endpoint,
		&reflectionv1.FileDescriptorsRequest{}, resp,
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
