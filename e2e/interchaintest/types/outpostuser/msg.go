/* Code generated by github.com/srdtrk/go-codegen, DO NOT EDIT. */
package outpostuser

import testtypes "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"

type InstantiateMsg struct {
	StorageOutpostAddress string `json:"storage_outpost_address"`
}

type ExecuteMsg struct {
	CallOutpost *ExecuteMsg_CallOutpost `json:"call_outpost,omitempty"`
}

type ExecuteMsg_CallOutpost struct {
	Msg *testtypes.ExecuteMsg `json:"msg,omitempty"`
}
