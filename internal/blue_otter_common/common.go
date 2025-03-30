package common

// ChatMessage represents a chat message in the system
type ChatMessage struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

// BootstrapInfo represents information about bootstrap nodes
type BootstrapInfo struct {
	Addresses []string `json:"addresses"`
}

// SystemNotification represents a system notification to be displayed to the user
type SystemNotification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
