package notifiers

import (
	"strings"
	"time"

	. "Notifier/models"
	. "Notifier/src/utils"
	"github.com/PuerkitoBio/goquery"
)

type Type1Notifier struct {
	BaseNotifier
}

func (Type1Notifier) New(baseNotifier *BaseNotifier) *Type1Notifier {
	baseNotifier.BoxNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr[class$=\"b-top-box\"]"
	baseNotifier.NumNoticeSelector = "#cms-content > div > div > div.type01 > table > tbody > tr:not([class$=\"b-top-box\"])"
	baseNotifier.ContentSelector = "#cms-content > div > div > div.bn-view-common01.type01 > div.b-main-box > div.b-content-box p"
	baseNotifier.ImagesSelector = "#cms-content > div > div > div.bn-view-common01.type01 > div.b-main-box > div.b-content-box img"

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
	url = notifier.NoticeUrl + strings.Join(split[0:2], "&")

	department := sel.Find("td:nth-child(5)").Text()
	department = strings.TrimSpace(department)

	date := time.Now().Format(time.RFC3339)
	date = date[:19]

	doc, err := NewDocumentFromPage(url)
	if err != nil {
    	ErrorLogger.Printf("Failed to load notice page: %s, URL: %s", err, url)
    	noticeChan <- Notice{} // 에러 발생 시 빈 Notice 반환 
    	return
	}

	contents := make([]string, 0, sel.Length())
	sel = doc.Find(notifier.ContentSelector)
	sel.Each(func(_ int, s *goquery.Selection) {
		if s.Text() != "" && s.Text() != "\u00a0" {
			str := strings.ReplaceAll(s.Text(), "\u00a0", " ")
			str = strings.ReplaceAll(str, "\n\n", "\\n")
			str = strings.ReplaceAll(str, "\n", "\\n")
			contents = append(contents, strings.TrimSpace(str))
		}
	})
	content := strings.Join(contents, "\\n")

	images := make([]string, 0, sel.Length())
	sel = doc.Find(notifier.ImagesSelector)
	sel.Each(func(_ int, s *goquery.Selection) {
		image, _ := s.Attr("src")
		if strings.Contains(image, "base64,") {
			return
		}
		if strings.Contains(image, "fonts.gstatic.com") {
			return
		}
		if !strings.Contains(image, "http://") && !strings.Contains(image, "https://") {
			image = "https://www.ajou.ac.kr" + image
		}
		images = append(images, image)
	})

	notice := Notice{
		ID:           id,
		Category:     category,
		Title:        title,
		Department:   department,
		Date:         date,
		Url:          url,
		Content:      content,
		Images:       images,
		EnglishTopic: notifier.EnglishTopic,
		KoreanTopic:  notifier.KoreanTopic,
	}

	noticeChan <- notice
}
