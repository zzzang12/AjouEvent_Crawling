package notifiers

import (
	"strings"
	"time"

	. "Notifier/models"
	"github.com/PuerkitoBio/goquery"
)

type Type5Notifier struct {
	BaseNotifier
}

func (Type5Notifier) New(baseNotifier *BaseNotifier) *Type5Notifier {
	baseNotifier.BoxNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr[class$=\"b-top-box\"]"
	baseNotifier.NumNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr:not([class$=\"b-top-box\"])"

	return &Type5Notifier{
		BaseNotifier: *baseNotifier,
	}
}

func (notifier *Type5Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel := doc.Find(notifier.NumNoticeSelector)
	if sel.Nodes == nil ||
		sel.Find("td:nth-child(1)").Nodes == nil ||
		sel.Find("td:nth-child(2) > div > a").Nodes == nil ||
		sel.Find("td:nth-child(4)").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type5Notifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	title, _ := sel.Find("td:nth-child(2) > div > a").Attr("title")
	title = title[:len(title)-17]
	title = strings.TrimSpace(title)

	url, _ := sel.Find("td:nth-child(2) > div > a").Attr("href")
	split := strings.FieldsFunc(url, func(c rune) bool {
		return c == '&'
	})
	url = notifier.BaseUrl + strings.Join(split[0:2], "&")

	department := sel.Find("td:nth-child(4)").Text()
	department = strings.TrimSpace(department)

	date := time.Now().Format(time.RFC3339)
	date = date[:len(date)-6]

	notice := Notice{
		ID:           id,
		Title:        title,
		Department:   department,
		Date:         date,
		Url:          url,
		EnglishTopic: notifier.EnglishTopic,
		KoreanTopic:  notifier.KoreanTopic,
	}

	noticeChan <- notice
}
