package logger

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	types1 "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func InitLogger() {
	path := "logs/"

	// Create directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.Create(path + "test.log")
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Exported function for info logging
func LogInfo(v ...interface{}) {
	InfoLogger.Println(v...)
}

// Exported function for err logging
func LogError(v ...interface{}) {
	ErrorLogger.Println(v...)
}

// Exported function to log events
func LogEvents(events []types1.Event) {
	for _, event := range events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			LogError("Failed to marshal event: ", err)
			continue
		}
		LogInfo("Event: ", string(eventJSON))
	}
}

// Retrieves outpost_address from the main 'wasm' event which is emitted by 'map_user_outpost'
// Consider moving this function out of logger.go
func ParseOutpostAddressFromEvent(events []types1.Event) string {
	for _, event := range events {
		if event.GetType() == "wasm" {
			for _, attribute := range event.GetAttributes() {
				if attribute.GetKey() == "outpost_address" {
					return attribute.GetValue()
				}
			}
		}
	}
	return ""
}

// Function to parse data
func parseData(resData string) (string, error) {
	// Decode the hex string
	dataFromHex, fromHexError := hex.DecodeString(resData)
	if fromHexError != nil {
		return "", fmt.Errorf("error decoding hex string: %w", fromHexError)
	}

	// Unmarshal dataFromHex into TxMsgData
	var txMsgData sdktypes.TxMsgData
	unmarshalMsgDataError := txMsgData.Unmarshal(dataFromHex)
	if unmarshalMsgDataError != nil {
		return "", fmt.Errorf("error unmarshalling TxMsgData: %w", unmarshalMsgDataError)
	}

	// Ensure there is at least one MsgResponse
	if len(txMsgData.MsgResponses) == 0 {
		return "", fmt.Errorf("no MsgResponses found in TxMsgData")
	}

	// Get the first MsgResponse and unmarshal it
	MsgResponseAsAny := txMsgData.MsgResponses[0]

	var executeResponseFromAny wasmtypes.MsgExecuteContractResponse
	unmarshalFromAnyError := executeResponseFromAny.Unmarshal(MsgResponseAsAny.Value)
	if unmarshalFromAnyError != nil {
		return "", fmt.Errorf("error unmarshalling MsgExecuteContractResponse: %w", unmarshalFromAnyError)
	}

	// Return the fully decoded data as a string
	fullyDecodedData := string(executeResponseFromAny.Data)

	return fullyDecodedData, nil
}
