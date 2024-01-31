package models

type BaseNotifier struct {
	BoxCount          int
	MaxNum            int
	URL               string
	Source            string
	ChannelID         string
	FsDocID           string
	NumNoticeSelector string
	BoxNoticeSelector string
}
