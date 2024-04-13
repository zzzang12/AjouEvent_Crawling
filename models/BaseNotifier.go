package models

type BaseNotifier struct {
	BaseUrl           string
	EnglishTopic      string
	KoreanTopic       string
	BoxCount          int
	MaxNum            int
	BoxNoticeSelector string
	NumNoticeSelector string
}
