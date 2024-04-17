package notifiers

import (
	"strings"
	"time"

	. "Notifier/models"
	"github.com/PuerkitoBio/goquery"
)

type Type3Notifier struct {
	BaseNotifier
}

func (Type3Notifier) New(baseNotifier *BaseNotifier) *Type3Notifier {
	baseNotifier.BoxNoticeSelector = "#nil"
	baseNotifier.NumNoticeSelector = "#contents > article > section > div > div:nth-child(3) > div.tb_w > table > tbody > tr"

	return &Type3Notifier{
		BaseNotifier: *baseNotifier,
	}
}

func (notifier *Type3Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel := doc.Find(notifier.NumNoticeSelector)
	if sel.Nodes == nil ||
		sel.Find("td:nth-child(1)").Nodes == nil ||
		sel.Find("td:nth-child(2)").Nodes == nil ||
		sel.Find("td:nth-child(3) > a").Nodes == nil ||
		sel.Find("td:nth-child(3) > a > span").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type3Notifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	category := sel.Find("td:nth-child(2)").Text()
	category = strings.TrimSpace(category)

	url, _ := sel.Find("td:nth-child(3) > a").Attr("href")
	split := strings.FieldsFunc(url, func(c rune) bool {
		return c == ' '
	})
	url = notifier.BaseUrl[:len(notifier.BaseUrl)-7] + "View.do?no=" + split[5]

	title := sel.Find("td:nth-child(3) > a > span").Text()
	title = strings.TrimSpace(title)

	date := time.Now().Format(time.RFC3339)
	date = date[:19]

	notice := Notice{
		ID:           id,
		Category:     category,
		Title:        title,
		Date:         date,
		Url:          url,
		EnglishTopic: notifier.EnglishTopic,
		KoreanTopic:  notifier.KoreanTopic,
	}

	noticeChan <- notice
}
