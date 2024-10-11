package testsuite

// Struct for channel endpoint
type ChannelEndpoint struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

// Struct for channel details
type Channel struct {
	Endpoint             ChannelEndpoint `json:"endpoint"`
	CounterpartyEndpoint ChannelEndpoint `json:"counterparty_endpoint"`
	Order                string          `json:"order"`
	Version              string          `json:"version"`
	ConnectionID         string          `json:"connection_id"`
}

// Struct for the full response
type ChannelStatusResponse struct {
	Channel       Channel `json:"channel"`
	ChannelStatus string  `json:"channel_status"`
}
