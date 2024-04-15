package notifiers

import (
	"strings"
	"time"

	. "Notifier/models"
	"github.com/PuerkitoBio/goquery"
)

type Type1Notifier struct {
	BaseNotifier
}

func (Type1Notifier) New(baseNotifier *BaseNotifier) *Type1Notifier {
	baseNotifier.BoxNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr[class$=\"b-top-box\"]"
	baseNotifier.NumNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr:not([class$=\"b-top-box\"])"

	return &Type1Notifier{
		BaseNotifier: *baseNotifier,
	}
}

func (notifier *Type1Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel := doc.Find(notifier.NumNoticeSelector)
	if sel.Nodes == nil ||
		sel.Find("td:nth-child(1)").Nodes == nil ||
		sel.Find("td:nth-child(2)").Nodes == nil ||
		sel.Find("td:nth-child(3) > div > a").Nodes == nil ||
		sel.Find("td:nth-child(5)").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type1Notifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	category := sel.Find("td:nth-child(2)").Text()
	category = strings.TrimSpace(category)

	title, _ := sel.Find("td:nth-child(3) > div > a").Attr("title")
	title = title[:len(title)-17]
	title = strings.TrimSpace(title)

	url, _ := sel.Find("td:nth-child(3) > div > a").Attr("href")
	split := strings.FieldsFunc(url, func(c rune) bool {
		return c == '&'
	})
	url = notifier.BaseUrl + strings.Join(split[0:2], "&")

	department := sel.Find("td:nth-child(5)").Text()
	department = strings.TrimSpace(department)

	date := time.Now().Format(time.RFC3339)
	date = date[:len(date)-6]

	notice := Notice{
		ID:           id,
		Category:     category,
		Title:        title,
		Department:   department,
		Date:         date,
		Url:          url,
		EnglishTopic: notifier.EnglishTopic,
		KoreanTopic:  notifier.KoreanTopic,
	}

	noticeChan <- notice
}
