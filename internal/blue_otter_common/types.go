package blue_otter_common

type ChatMessage struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

type BootstrapInfo struct {
	Addresses []string `json:"addresses"`
}