package models

type NotifierConfig struct {
	FsDocID   string `json:"fsDocID"`
	URL       string `json:"url"`
	Source    string `json:"source"`
	ChannelID string `json:"channelId"`
	Type      int    `json:"type"`
}
