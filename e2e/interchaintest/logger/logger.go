package logger

import (
	"encoding/json"
	"log"
	"os"

	types1 "github.com/cometbft/cometbft/abci/types"
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
