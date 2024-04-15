package notifiers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "Notifier/models"
	. "Notifier/src/utils"
	"cloud.google.com/go/firestore"
	"github.com/PuerkitoBio/goquery"
)

type BaseNotifier struct {
	Type              int
	BaseUrl           string
	EnglishTopic      string
	KoreanTopic       string
	BoxCount          int
	MaxNum            int
	BoxNoticeSelector string
	NumNoticeSelector string
}

func (BaseNotifier) New(config NotifierConfig) *BaseNotifier {
	dbData := LoadDbData(config.EnglishTopic)

	return &BaseNotifier{
		Type:         config.Type,
		BaseUrl:      config.BaseUrl,
		EnglishTopic: config.EnglishTopic,
		KoreanTopic:  config.KoreanTopic,
		BoxCount:     int(dbData["box"].(int64)),
		MaxNum:       int(dbData["num"].(int64)),
	}
}

func (notifier *BaseNotifier) Notify() {
	defer func() {
		recover()
	}()

	notices := notifier.scrapeNotice()
	for _, notice := range notices {
		SendCrawlingWebhook(os.Getenv("WEBHOOK_ENDPOINT"), notice)
		SentNoticeLogger.Println(notice)
	}
}

func (notifier *BaseNotifier) scrapeNotice() []Notice {
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

	return notices
}

func (notifier *BaseNotifier) checkHTML(doc *goquery.Document) error {
	if notifier.isInvalidHTML(doc) {
		errMsg := "HTML structure has changed at " + notifier.KoreanTopic
		return errors.New(errMsg)
	}
	return nil
}

func (notifier *BaseNotifier) isInvalidHTML(doc *goquery.Document) bool {
	switch notifier.Type {
	case 1:
		return Type1Notifier{}.New(notifier).isInvalidHTML(doc)
	case 2:
		return Type2Notifier{}.New(notifier).isInvalidHTML(doc)
	case 3:
		return Type3Notifier{}.New(notifier).isInvalidHTML(doc)
	case 4:
		return Type4Notifier{}.New(notifier).isInvalidHTML(doc)
	case 5:
		return Type5Notifier{}.New(notifier).isInvalidHTML(doc)
	default:
		return false
	}
}

func (notifier *BaseNotifier) scrapeBoxNotice(doc *goquery.Document) []Notice {
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

func (notifier *BaseNotifier) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find(notifier.NumNoticeSelector)
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeCountReference := GetNumNoticeCountReference(doc, notifier.EnglishTopic, notifier.BoxNoticeSelector)
	numNoticeCount := min(maxNum-notifier.MaxNum, numNoticeCountReference)
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

func (notifier *BaseNotifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	switch notifier.Type {
	case 1:
		Type1Notifier{}.New(notifier).getNotice(sel, noticeChan)
	case 2:
		Type2Notifier{}.New(notifier).getNotice(sel, noticeChan)
	case 3:
		Type3Notifier{}.New(notifier).getNotice(sel, noticeChan)
	case 4:
		Type4Notifier{}.New(notifier).getNotice(sel, noticeChan)
	case 5:
		Type5Notifier{}.New(notifier).getNotice(sel, noticeChan)
	}
}
