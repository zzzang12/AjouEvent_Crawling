package models

type BaseNotifier struct {
	BoxCount          int
	MaxNum            int
	URL               string
	Source            string
	ChannelID         string
	DocumentID        string
	NumNoticeSelector string
	BoxNoticeSelector string
}
