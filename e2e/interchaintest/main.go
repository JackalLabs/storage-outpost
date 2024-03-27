package main

import (
	"encoding/base64"

	"github.com/cosmos/gogoproto/proto"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"

	filetreetypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/filetreetypes"
	e2etypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	sdkcodectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// Our sandbox to help with debugging
func main() {

	logger.InitLogger()

	filetreeMsg := &filetreetypes.MsgPostKey{
		Creator: "alice",
		Key:     "placeholder",
	}

	logger.LogInfo(filetreeMsg)

	// filetree msg is valid proto msg
	protoMsg := []proto.Message{filetreeMsg}

	logger.LogInfo("=============")
	logger.LogInfo(protoMsg)

	bz, err := proto.Marshal(protoMsg[0])
	if err != nil {
		panic(err)
	}

	logger.LogInfo("=============")
	logger.LogInfo(bz)

	protoAny := &sdkcodectypes.Any{
		TypeUrl: "/" + proto.MessageName(protoMsg[0]),
		Value:   bz,
		// Note that cachedValue is not public, but I
		// don't believe we're even using it anyway
		// cachedValue: protoMsg[0],
	}

	protoAny.TypeUrl = "/canine_chain.filetree.MsgPostKey"

	Stargate := &e2etypes.StargateCosmosMsg{
		TypeUrl: protoAny.TypeUrl,
		Value:   base64.StdEncoding.EncodeToString(protoAny.Value),
	}

	logger.LogInfo(Stargate)
}
