package notifiers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "Notifier/models"
	. "Notifier/src/utils"
	"cloud.google.com/go/firestore"
	"github.com/PuerkitoBio/goquery"
)

type Type4Notifier BaseNotifier

func (Type4Notifier) New(config NotifierConfig) *Type4Notifier {
	dbData := LoadDbData(config.EnglishTopic)

	return &Type4Notifier{
		BaseUrl:      config.BaseUrl,
		EnglishTopic: config.EnglishTopic,
		KoreanTopic:  config.KoreanTopic,
		BoxCount:     int(dbData["box"].(int64)),
		MaxNum:       int(dbData["num"].(int64)),
		//BoxNoticeSelector: "#contents > article > section > div > div:nth-child(3) > div.cnt_in.x_divi_w > ul > li > div",
		BoxNoticeSelector: "",
		NumNoticeSelector: "#contents > article > section > div > div:nth-child(3) > div.tb_w > table > tbody > tr",
	}
}

func (notifier *Type4Notifier) Notify() {
	defer func() {
		recover()
	}()

	notices := notifier.scrapeNotice()
	for _, notice := range notices {
		SendCrawlingWebhook("", notice)
	}
}

func (notifier *Type4Notifier) scrapeNotice() []Notice {
	resp, err := http.Get(notifier.BaseUrl)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		ErrorLogger.Panicf("status code error: %s", resp.Status)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	err = notifier.checkHTML(doc)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	boxNotices := notifier.scrapeBoxNotice(doc)

	numNotices := notifier.scrapeNumNotice(doc)

	notices := make([]Notice, 0, len(boxNotices)+len(numNotices))
	for _, notice := range boxNotices {
		notices = append(notices, notice)
	}
	for _, notice := range numNotices {
		notices = append(notices, notice)
	}

	for _, notice := range notices {
		SentNoticeLogger.Println("notice =>", notice)
	}

	return notices
}

func (notifier *Type4Notifier) checkHTML(doc *goquery.Document) error {
	if notifier.isInvalidHTML(doc) {
		errMsg := strings.Join([]string{"HTML structure has changed at ", notifier.KoreanTopic}, "")
		return errors.New(errMsg)
	}
	return nil
}

func (notifier *Type4Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel := doc.Find(notifier.NumNoticeSelector)
	if sel.Nodes == nil ||
		sel.Find("td:nth-child(1)").Nodes == nil ||
		sel.Find("td:nth-child(2)").Nodes == nil ||
		sel.Find("td:nth-child(3) > a").Nodes == nil ||
		sel.Find("td:nth-child(3) > a > span").Nodes == nil ||
		sel.Find("td:nth-child(4)").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type4Notifier) scrapeBoxNotice(doc *goquery.Document) []Notice {
	boxNoticeSels := doc.Find(notifier.BoxNoticeSelector)
	boxCount := boxNoticeSels.Length()

	boxNoticeChan := make(chan Notice, boxCount)
	boxNotices := make([]Notice, 0, boxCount)
	boxNoticeCount := boxCount - notifier.BoxCount

	if boxCount > notifier.BoxCount {
		boxNoticeSels = boxNoticeSels.FilterFunction(func(i int, _ *goquery.Selection) bool {
			return i < boxNoticeCount
		})

		boxNoticeSels.Each(func(_ int, boxNotice *goquery.Selection) {
			go notifier.getNotice(boxNotice, boxNoticeChan)
		})

		for i := 0; i < boxNoticeCount; i++ {
			boxNotices = append(boxNotices, <-boxNoticeChan)
		}

		notifier.BoxCount = boxCount
		_, err := Client.Collection("notice").Doc(notifier.EnglishTopic).Update(context.Background(), []firestore.Update{
			{
				Path:  "box",
				Value: notifier.BoxCount,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
	} else if boxCount < notifier.BoxCount {
		notifier.BoxCount = boxCount
		_, err := Client.Collection("notice").Doc(notifier.EnglishTopic).Update(context.Background(), []firestore.Update{
			{
				Path:  "box",
				Value: notifier.BoxCount,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
	}

	return boxNotices
}

func (notifier *Type4Notifier) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find(notifier.NumNoticeSelector)
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeCount := min(maxNum-notifier.MaxNum, 10)
	numNoticeChan := make(chan Notice, numNoticeCount)
	numNotices := make([]Notice, 0, numNoticeCount)

	if maxNum > notifier.MaxNum {
		numNoticeSels = numNoticeSels.FilterFunction(func(i int, _ *goquery.Selection) bool {
			return i < numNoticeCount
		})

		numNoticeSels.Each(func(_ int, numNotice *goquery.Selection) {
			go notifier.getNotice(numNotice, numNoticeChan)
		})

		for i := 0; i < numNoticeCount; i++ {
			numNotices = append(numNotices, <-numNoticeChan)
		}

		notifier.MaxNum = maxNum
		_, err = Client.Collection("notice").Doc(notifier.EnglishTopic).Update(context.Background(), []firestore.Update{
			{
				Path:  "num",
				Value: notifier.MaxNum,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
	}

	return numNotices
}

func (notifier *Type4Notifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	category := sel.Find("td:nth-child(2)").Text()
	category = strings.TrimSpace(category)

	url, _ := sel.Find("td:nth-child(3) > a").Attr("href")
	split := strings.FieldsFunc(url, func(c rune) bool {
		return c == ' '
	})
	url = split[5]
	url = strings.Join([]string{notifier.BaseUrl[:len(notifier.BaseUrl)-7], "View.do?no=", url}, "")

	title := sel.Find("td:nth-child(3) > a > span").Text()

	date := time.Now().Format(time.RFC3339)
	date = date[:len(date)-6]

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
