package models

type NotifierConfig struct {
	DocumentID string `json:"documentID"`
	URL        string `json:"url"`
	Source     string `json:"source"`
	ChannelID  string `json:"channelID"`
	Type       int    `json:"type"`
}
