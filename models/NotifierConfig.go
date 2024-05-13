package models

type NotifierConfig struct {
	Type         int    `json:"type"`
	EnglishTopic string `json:"englishTopic"`
	KoreanTopic  string `json:"koreanTopic"`
	NoticeUrl    string `json:"noticeUrl"`
}
