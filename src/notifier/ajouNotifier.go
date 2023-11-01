package notifier

type Notifier interface {
	Notify()
}

type Notice struct {
	ID         string
	Category   string
	Title      string
	Department string
	Date       string
	Link       string
}

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
