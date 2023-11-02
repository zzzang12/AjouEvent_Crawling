package notifier

import (
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

type AjouMENotifier BaseNotifier

func (AjouMENotifier) New() *AjouMENotifier {
	fsDocID := "ajouME"
	dsnap, err := Client.Collection("notice").Doc(fsDocID).Get(context.Background())
	if err != nil {
		ErrorLogger.Panic(err)
	}
	dbData := dsnap.Data()

	return &AjouMENotifier{
		MaxNum:            int(dbData["num"].(int64)),
		URL:               "https://me.ajou.ac.kr/me/board/board01.jsp",
		Source:            "[기계공학과]",
		ChannelID:         "기계공학과-공지사항",
		FsDocID:           fsDocID,
		NumNoticeSelector: "#jwxe_main_content > div > div.list_wrap > table > tbody > tr",
	}
}

func (notifier *AjouMENotifier) Notify() {
	defer func() {
		recover()
	}()

	notices := notifier.scrapeNotice()
	for _, notice := range notices {
		notifier.sendNoticeToSlack(notice)
	}
}

func (notifier *AjouMENotifier) scrapeNotice() []Notice {
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

	notices := make([]Notice, 0, len(numNotices))
	for _, notice := range numNotices {
		notices = append(notices, notice)
	}

	for _, notice := range notices {
		SentNoticeLogger.Println("notice =>", notice)
	}

	return notices
}

func (notifier *AjouMENotifier) checkHTML(doc *goquery.Document) error {
	if notifier.isInvalidHTML(doc) {
		errMsg := strings.Join([]string{"HTML structure has changed at ", notifier.Source}, "")
		return errors.New(errMsg)
	}
	return nil
}

func (notifier *AjouMENotifier) isInvalidHTML(doc *goquery.Document) bool {
	sel1 := doc.Find(notifier.NumNoticeSelector)
	if sel1.Nodes == nil ||
		sel1.Find("td:nth-child(1)").Nodes == nil ||
		sel1.Find("td:nth-child(2) > a").Nodes == nil ||
		sel1.Find("td:nth-child(5)").Nodes == nil ||
		sel1.Find("td:nth-child(6)").Nodes == nil {
		return true
	}
	return false
}

func (notifier *AjouMENotifier) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find(notifier.NumNoticeSelector)
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeCount := min(maxNum-notifier.MaxNum, MaxNumNoticeCount)
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
		_, err = Client.Collection("notice").Doc(notifier.FsDocID).Update(context.Background(), []firestore.Update{
			{
				Path:  "num",
				Value: notifier.MaxNum,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
		BoxCountMaxNumLogger.Println("maxNum =>", notifier.MaxNum)
	}

	return numNotices
}

func (notifier *AjouMENotifier) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	title := sel.Find("td:nth-child(2) > a").Text()
	title = strings.TrimSpace(title)

	link, _ := sel.Find("td:nth-child(2) > a").Attr("href")
	split := strings.FieldsFunc(link, func(c rune) bool {
		return c == '&'
	})
	link = strings.Join(split[0:2], "&")
	link = strings.Join([]string{notifier.URL, link}, "")

	category := sel.Find("td:nth-child(5)").Text()
	category = strings.TrimSpace(category)

	date := sel.Find("td:nth-child(6)").Text()
	date = strings.TrimSpace(date)
	month := date[5:7]
	if month[0] == '0' {
		month = month[1:]
	}
	day := date[8:10]
	if day[0] == '0' {
		day = day[1:]
	}
	date = strings.Join([]string{month, "월", day, "일"}, "")

	notice := Notice{ID: id, Category: category, Title: title, Date: date, Link: link}

	noticeChan <- notice
}

func (notifier *AjouMENotifier) sendNoticeToSlack(notice Notice) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	category := strings.Join([]string{"[", notice.Category, "]"}, "")
	footer := strings.Join([]string{category, notifier.Source}, " ")

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
