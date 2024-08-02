package notifiers

import (
	"errors"
	"os"
	"strconv"
	"strings"

	. "Notifier/models"
	. "Notifier/src/utils"
	"github.com/PuerkitoBio/goquery"
)

type BaseNotifier struct {
	Type              int
	NoticeUrl         string
	EnglishTopic      string
	KoreanTopic       string
	BoxCount          int
	MaxNum            int
	BoxNoticeSelector string
	NumNoticeSelector string
	ContentSelector   string
	ImagesSelector    string
}

func (BaseNotifier) New(config NotifierConfig) *BaseNotifier {
	boxCount, maxNum := LoadDbData(config.EnglishTopic)

	return &BaseNotifier{
		Type:         config.Type,
		NoticeUrl:    config.NoticeUrl,
		EnglishTopic: config.EnglishTopic,
		KoreanTopic:  config.KoreanTopic,
		BoxCount:     boxCount,
		MaxNum:       maxNum,
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
	doc := NewDocumentFromPage(notifier.NoticeUrl)

	err := notifier.checkHTML(doc)
	if err != nil {
		ErrorLogger.Panic(err, notifier.EnglishTopic)
	}

	boxNotices := notifier.scrapeBoxNotice(doc)
	numNotices := notifier.scrapeNumNotice(doc)

	notices := make([]Notice, 0, len(boxNotices)+len(numNotices))
	notices = append(notices, boxNotices...)
	notices = append(notices, numNotices...)

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

	if boxCount == notifier.BoxCount {
		return make([]Notice, 0)
	}

	if boxCount < notifier.BoxCount {
		notifier.BoxCount = boxCount
		query := "UPDATE notice AS n JOIN topic AS t ON n.topic_id = t.id SET n.value = ? WHERE t.department = ? AND n.type = ?"
		_, err := DB.Exec(query, notifier.BoxCount, notifier.EnglishTopic, "box")
		if err != nil {
			ErrorLogger.Panic(err, notifier.EnglishTopic)
		}
		return make([]Notice, 0)
	}

	boxNoticeChan := make(chan Notice, boxCount)
	boxNotices := make([]Notice, 0, boxCount)
	boxNoticeCount := boxCount - notifier.BoxCount

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
	query := "UPDATE notice AS n JOIN topic AS t ON n.topic_id = t.id SET n.value = ? WHERE t.department = ? AND n.type = ?"
	_, err := DB.Exec(query, notifier.BoxCount, notifier.EnglishTopic, "box")
	if err != nil {
		ErrorLogger.Panic(err, notifier.EnglishTopic)
	}

	return boxNotices
}

func (notifier *BaseNotifier) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find(notifier.NumNoticeSelector)
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err, notifier.EnglishTopic)
	}

	if maxNum == notifier.MaxNum {
		return make([]Notice, 0)
	}

	if maxNum < notifier.MaxNum {
		notifier.MaxNum = maxNum
		query := "UPDATE notice AS n JOIN topic AS t ON n.topic_id = t.id SET n.value = ? WHERE t.department = ? AND n.type = ?"
		_, err = DB.Exec(query, notifier.MaxNum, notifier.EnglishTopic, "num")
		if err != nil {
			ErrorLogger.Panic(err, notifier.EnglishTopic)
		}
		return make([]Notice, 0)
	}

	numNoticeCountReference := GetNumNoticeCountReference(doc, notifier.EnglishTopic, notifier.NumNoticeSelector)
	numNoticeCount := min(maxNum-notifier.MaxNum, numNoticeCountReference)
	numNoticeChan := make(chan Notice, numNoticeCount)
	numNotices := make([]Notice, 0, numNoticeCount)

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
	query := "UPDATE notice AS n JOIN topic AS t ON n.topic_id = t.id SET n.value = ? WHERE t.department = ? AND n.type = ?"
	_, err = DB.Exec(query, notifier.MaxNum, notifier.EnglishTopic, "num")
	if err != nil {
		ErrorLogger.Panic(err, notifier.EnglishTopic)
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
