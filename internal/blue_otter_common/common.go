package common

// common.go contains all custom struct types for the application

// ChatMessage represents a chat message in the system
type ChatMessage struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

// BootstrapInfo represents information about bootstrap nodes
type BootstrapInfo struct {
	BootStrapNodeAddresses []string `json:"bootstrap_node_addresses"`
	Addresses  []string `json:"addresses"`
	PrivateKey string   `json:"private_key,omitempty"`
	PeerID     string   `json:"peer_id,omitempty"`
}

// SystemNotification represents a system notification to be displayed to the user
type SystemNotification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
