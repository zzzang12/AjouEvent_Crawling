package notifier

import (
	"Notifier/src/models"
	. "Notifier/src/utils"
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/slack-go/slack"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Type4Notifier models.BaseNotifier

func (Type4Notifier) New(config models.NotifierConfig) *Type4Notifier {
	documentID := config.DocumentID
	dsnap, err := Client.Collection("notice").Doc(documentID).Get(context.Background())
	if err != nil {
		ErrorLogger.Panic(err)
	}
	dbData := dsnap.Data()

	return &Type4Notifier{
		URL:               config.URL,
		Source:            config.Source,
		ChannelID:         config.ChannelID,
		DocumentID:        documentID,
		BoxCount:          int(dbData["box"].(int64)),
		MaxNum:            int(dbData["num"].(int64)),
		BoxNoticeSelector: "#nil",
		NumNoticeSelector: "#contents > article > section > div > div:nth-child(3) > div.tb_w > table > tbody > tr",
	}
}

func (notifier *Type4Notifier) Notify() {
	defer func() {
		recover()
	}()

	notices := notifier.scrapeNotice()
	for _, notice := range notices {
		notifier.sendNoticeToSlack(notice)
	}
}

func (notifier *Type4Notifier) scrapeNotice() []models.Notice {
	resp, err := http.Get(notifier.URL)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	if resp.StatusCode != 200 {
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

	numNotices := notifier.scrapeNumNotice(doc)

	notices := make([]models.Notice, 0, len(numNotices))
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
		errMsg := strings.Join([]string{"HTML structure has changed at ", notifier.Source}, "")
		return errors.New(errMsg)
	}
	return nil
}

func (notifier *Type4Notifier) isInvalidHTML(doc *goquery.Document) bool {
	sel1 := doc.Find(notifier.NumNoticeSelector)
	if sel1.Nodes == nil ||
		sel1.Find("td:nth-child(1)").Nodes == nil ||
		sel1.Find("td:nth-child(2)").Nodes == nil ||
		sel1.Find("td:nth-child(3) > a").Nodes == nil ||
		sel1.Find("td:nth-child(3) > a > span").Nodes == nil ||
		sel1.Find("td:nth-child(4)").Nodes == nil {
		return true
	}
	return false
}

func (notifier *Type4Notifier) scrapeNumNotice(doc *goquery.Document) []models.Notice {
	numNoticeSels := doc.Find(notifier.NumNoticeSelector)
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeCount := min(maxNum-notifier.MaxNum, MaxNumNoticeCount)
	numNoticeChan := make(chan models.Notice, numNoticeCount)
	numNotices := make([]models.Notice, 0, numNoticeCount)

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
		_, err = Client.Collection("notice").Doc(notifier.DocumentID).Update(context.Background(), []firestore.Update{
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

func (notifier *Type4Notifier) getNotice(sel *goquery.Selection, noticeChan chan models.Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	category := sel.Find("td:nth-child(2)").Text()
	category = strings.TrimSpace(category)

	link, _ := sel.Find("td:nth-child(3) > a").Attr("href")
	split := strings.FieldsFunc(link, func(c rune) bool {
		return c == ' '
	})
	link = split[5:6][0]
	link = strings.Join([]string{notifier.URL[:len(notifier.URL)-7], "View.do?no=", link}, "")

	title := sel.Find("td:nth-child(3) > a > span").Text()

	date := sel.Find("td:nth-child(4)").Text()
	month := date[5:7]
	if month[0] == '0' {
		month = month[1:]
	}
	day := date[8:10]
	if day[0] == '0' {
		day = day[1:]
	}
	date = strings.Join([]string{month, "월", day, "일"}, "")

	notice := models.Notice{ID: id, Category: category, Title: title, Date: date, Link: link}

	noticeChan <- notice
}

func (notifier *Type4Notifier) sendNoticeToSlack(notice models.Notice) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	category := strings.Join([]string{"[", notice.Category, "]"}, "")
	footer := strings.Join([]string{notifier.Source, category}, " ")

	attachment := slack.Attachment{
		Color:      "#0072ce",
		Title:      strings.Join([]string{notice.Date, notice.Title}, " "),
		Text:       notice.Link,
		Footer:     footer,
		FooterIcon: "https://github.com/zzzang12/Notifier/assets/70265177/48fd0fd7-80e2-4309-93da-8a6bc957aacf",
	}

	_, _, err := api.PostMessage(notifier.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		ErrorLogger.Panic(err)
	}
}
