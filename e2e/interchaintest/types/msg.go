package types

import "encoding/json"

// InstantiateMsg is the message to instantiate cw-ica-controller
type InstantiateMsg struct {
	// The admin address. If not specified, the sender is the admin.
	Owner *string `json:"owner,omitempty"`
	// The options to initialize the IBC channel upon contract instantiation.
	// If not specified, the IBC channel is not initialized, and the relayer must.
	ChannelOpenInitOptions ChannelOpenInitOptions `json:"channel_open_init_options"`
	// The contract address that the channel and packet lifecycle callbacks are sent to.
	// If not specified, then no callbacks are sent.
	SendCallbacksTo *string `json:"send_callbacks_to,omitempty"`
	// The callback information to be used
	Callback *Callback `json:"callback,omitempty"`
}

type Callback struct {
	// The address of the contract that we will call back
	Contract string `json:"contract,omitempty"`
	// The msg we will make the above contract execute
	Msg Binary `json:"msg,omitempty"`
	/// The owner of the outpost
	OutpostOwner string `json:"outpost_owner,omitempty"`
}

// ExecuteMsg is the message to execute cw-ica-controller
type ExecuteMsg struct {
	CreateChannel         *ExecuteMsg_CreateChannel         `json:"create_channel,omitempty"`
	CreateTransferChannel *ExecuteMsg_CreateTransferChannel `json:"create_transfer_channel,omitempty"`
	SendCosmosMsgs        *ExecuteMsg_SendCosmosMsgs        `json:"send_cosmos_msgs,omitempty"`
	SendCustomIcaMessages *ExecuteMsg_SendCustomIcaMessages `json:"send_custom_ica_messages,omitempty"`
	UpdateCallbackAddress *ExecuteMsg_UpdateCallbackAddress `json:"update_callback_address,omitempty"`
}

// QueryMsg is the message to query cw-ica-controller
type QueryMsg struct {
	GetChannel       *struct{}       `json:"get_channel,omitempty"`
	GetContractState *struct{}       `json:"get_contract_state,omitempty"`
	Ownership        *struct{}       `json:"ownership,omitempty"`
	GetNote          *GetNoteRequest `json:"get_note,omitempty"`
}

// GetNoteRequest is the struct for the GetNote query
type GetNoteRequest struct {
	Address string `json:"address"`
}

// MigrateMsg is the message to migrate cw-ica-controller
type MigrateMsg struct {
	// ContractAddr string `json:"contract_addr"`
	// NewCodeID    string `json:"new_code_id"`
	// // base64 encoded bytes
	// Msg string `json:"msg"`
}

// `CreateChannel` makes the contract submit a stargate MsgChannelOpenInit to the chain.
// This is a wrapper around [`options::ChannelOpenInitOptions`] and thus requires the
// same fields. If not specified, then the options specified in the contract instantiation
// are used.
type ExecuteMsg_CreateChannel struct {
	// The options to initialize the IBC channel.
	// If not specified, the options specified in the contract instantiation are used.
	ChannelOpenInitOptions *ChannelOpenInitOptions `json:"channel_open_init_options,omitempty"`
}

// `CreateTransferChannel` is opening a transfer channel
// for development purposees only. Not using ChannelOpenInitOptions
type ExecuteMsg_CreateTransferChannel struct {
	// The connection id on this chain.
	ConnectionId string `json:"connection_id"`
	// The counterparty connection id on the counterparty chain.
	CounterpartyConnectionId string `json:"counterparty_connection_id"`
	// The optional counterparty port id.
	CounterpartyPortId *string `json:"counterparty_port_id,omitempty"`
	// The optional tx encoding.
	TxEncoding *string `json:"tx_encoding,omitempty"`
}

// `SendCosmosMsgs` converts the provided array of [`CosmosMsg`] to an ICA tx and sends them to the ICA host.
// [`CosmosMsg::Stargate`] and [`CosmosMsg::Wasm`] are only supported if the [`TxEncoding`](crate::ibc::types::metadata::TxEncoding) is [`TxEncoding::Protobuf`](crate::ibc::types::metadata::TxEncoding).
//
// **This is the recommended way to send messages to the ICA host.**
type ExecuteMsg_SendCosmosMsgs struct {
	// The stargate messages to convert and send to the ICA host.
	Messages []ContractCosmosMsg `json:"messages"`
	// Optional memo to include in the ibc packet.
	PacketMemo *string `json:"packet_memo,omitempty"`
	// Optional timeout in seconds to include with the ibc packet.
	// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
	TimeoutSeconds *uint64 `json:"timeout_seconds,omitempty"`
}

// `SendCustomIcaMessages` sends custom messages from the ICA controller to the ICA host.
type ExecuteMsg_SendCustomIcaMessages struct {
	Messages string `json:"messages"`
	// Optional memo to include in the ibc packet.
	PacketMemo *string `json:"packet_memo,omitempty"`
	// Optional timeout in seconds to include with the ibc packet.
	// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
	TimeoutSeconds *uint64 `json:"timeout_seconds,omitempty"`
}

// `UpdateCallbackAddress` updates the contract callback address.
type ExecuteMsg_UpdateCallbackAddress struct {
	/// The new callback address. If not specified, then no callbacks are sent.
	CallbackAddress *string `json:"callback_address,omitempty"`
}

// ToString returns a string representation of the message
func (m *InstantiateMsg) ToString() string {
	return toString(m)
}

// ToString returns a string representation of the message
func (m *ExecuteMsg) ToString() string {
	return toString(m)
}

// ToString returns a string representation of the message
func (m *QueryMsg) ToString() string {
	return toString(m)
}

// ToString returns a string representation of the message
func (m *MigrateMsg) ToString() string {
	return toString(m)
}

func toString(v any) string {
	jsonBz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(jsonBz)
}

/*
Binary is a wrapper around Vec<u8> to add base64 de/serialization with serde. It also adds some helper methods to help encode inline.

This is only needed as serde-json-{core,wasm} has a horrible encoding for Vec<u8>. See also <https://github.com/CosmWasm/cosmwasm/blob/main/docs/MESSAGE_TYPES.md>.
*/
type Binary string

// Status is the status of an IBC channel.
type ChannelStatus string

const (
	// Uninitialized is the default state of the channel.
	ChannelStatus_StateUninitializedUnspecified ChannelStatus = "STATE_UNINITIALIZED_UNSPECIFIED"
	// Init is the state of the channel when it is created.
	ChannelStatus_StateInit ChannelStatus = "STATE_INIT"
	// TryOpen is the state of the channel when it is trying to open.
	ChannelStatus_StateTryopen ChannelStatus = "STATE_TRYOPEN"
	// Open is the state of the channel when it is open.
	ChannelStatus_StateOpen ChannelStatus = "STATE_OPEN"
	// Closed is the state of the channel when it is closed.
	ChannelStatus_StateClosed ChannelStatus = "STATE_CLOSED"
	// The channel has just accepted the upgrade handshake attempt and is flushing in-flight packets. Added in `ibc-go` v8.1.0.
	ChannelStatus_StateFlushing ChannelStatus = "STATE_FLUSHING"
	// The channel has just completed flushing any in-flight packets. Added in `ibc-go` v8.1.0.
	ChannelStatus_StateFlushcomplete ChannelStatus = "STATE_FLUSHCOMPLETE"
)
