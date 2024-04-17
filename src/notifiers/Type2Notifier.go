package notifiers

import (
	"strings"
	"time"

	. "Notifier/models"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type Type2Notifier struct {
	BaseNotifier
}

func (Type2Notifier) New(baseNotifier *BaseNotifier) *Type2Notifier {
	baseNotifier.BoxNoticeSelector = "#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):has(td:first-child > img)"
	baseNotifier.NumNoticeSelector = "#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):not(:has(td:first-child > img))"

	return &Type2Notifier{
		BaseNotifier: *baseNotifier,
	}
}

func (notifier *Type2Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel := doc.Find(notifier.NumNoticeSelector)
	if sel.Nodes == nil ||
		sel.Find("td:nth-child(1)").Nodes == nil ||
		sel.Find("td:nth-child(3) > a").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type2Notifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	var id string
	if sel.Find("td:nth-child(1):has(img)").Nodes != nil {
		id = "공지"
	} else {
		id = sel.Find("td:nth-child(1)").Text()
		id = strings.TrimSpace(id)
	}

	title := sel.Find("td:nth-child(3) > a").Text()
	title, _, _ = transform.String(korean.EUCKR.NewDecoder(), title)
	title = strings.TrimSpace(title)

	url, _ := sel.Find("td:nth-child(3) > a").Attr("href")
	split := strings.FieldsFunc(url, func(c rune) bool {
		return c == '&'
	})
	url = notifier.BaseUrl + "&" + strings.Join(split[1:3], "&")

	date := time.Now().Format(time.RFC3339)
	date = date[:19]

	notice := Notice{
		ID:           id,
		Title:        title,
		Date:         date,
		Url:          url,
		EnglishTopic: notifier.EnglishTopic,
		KoreanTopic:  notifier.KoreanTopic,
	}

	noticeChan <- notice
}
