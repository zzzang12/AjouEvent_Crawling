package models

type Notice struct {
	ID           string   `json:"id"`
	Category     string   `json:"category"`
	Title        string   `json:"title"`
	Department   string   `json:"department"`
	Date         string   `json:"date"`
	Url          string   `json:"url"`
	Content      string   `json:"content"`
	Images       []string `json:"images"`
	EnglishTopic string   `json:"englishTopic"`
	KoreanTopic  string   `json:"koreanTopic"`
}
