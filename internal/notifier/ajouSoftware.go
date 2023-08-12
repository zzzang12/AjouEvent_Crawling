package notifier

import (
	. "Notifier/internal/utils"
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/slack-go/slack"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type AjouSoftwareSource NoticeSource

func (AjouSoftwareSource) NewNotifier() *AjouSoftwareSource {
	fsDocID := "ajouSoftware"
	dsnap, err := Client.Collection("notice").Doc(fsDocID).Get(context.Background())
	if err != nil {
		ErrorLogger.Panic(err)
	}
	dbData := dsnap.Data()

	return &AjouSoftwareSource{
		BoxCount:  int(dbData["box"].(int64)),
		MaxNum:    int(dbData["num"].(int64)),
		URL:       "http://software.ajou.ac.kr/bbs/board.php?tbl=notice",
		ChannelID: "소프트웨어학과-공지사항",
		FsDocID:   fsDocID,
	}
}

func (source *AjouSoftwareSource) Notify() {
	defer func() {
		recover()
	}()

	notices := source.scrapeNotice()
	for _, notice := range notices {
		source.sendNoticeToSlack(notice)
	}
}

func (source *AjouSoftwareSource) scrapeNotice() []Notice {
	resp, err := http.Get(source.URL)
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

	err = source.checkHTML(doc)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	boxNotices := source.scrapeBoxNotice(doc)

	numNotices := source.scrapeNumNotice(doc)

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

func (source *AjouSoftwareSource) checkHTML(doc *goquery.Document) error {
	if source.isInvalidHTML(doc) {
		errMsg := strings.Join([]string{"notifier can't work because HTML structure has changed at ", source.ChannelID}, "")
		return errors.New(errMsg)
	}
	return nil
}

func (source *AjouSoftwareSource) isInvalidHTML(doc *goquery.Document) bool {
	sel1 := doc.Find("#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):has(td:first-child > img)")
	sel2 := doc.Find("#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):not(:has(td:first-child > img))")
	if sel1.Nodes == nil || sel2.Nodes == nil ||
		sel1.Find("td:nth-child(1)").Nodes == nil ||
		sel1.Find("td:nth-child(3) > a").Nodes == nil ||
		sel1.Find("td:nth-child(3) > p:first-of-type").Nodes == nil ||
		sel2.Find("td:nth-child(1)").Nodes == nil ||
		sel2.Find("td:nth-child(3) > a").Nodes == nil ||
		sel2.Find("td:nth-child(3) > p:first-of-type").Nodes == nil {
		return true
	}
	return false
}

func (source *AjouSoftwareSource) scrapeBoxNotice(doc *goquery.Document) []Notice {
	boxNoticeSels := doc.Find("#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):has(td:first-child > img)")
	boxCount := boxNoticeSels.Length()

	boxNoticeChan := make(chan Notice, boxCount)
	boxNotices := make([]Notice, 0, boxCount)
	boxNoticeCount := boxCount - source.BoxCount

	if boxCount > source.BoxCount {
		boxNoticeSels = boxNoticeSels.FilterFunction(func(i int, _ *goquery.Selection) bool {
			return i < boxNoticeCount
		})

		boxNoticeSels.Each(func(_ int, boxNotice *goquery.Selection) {
			go source.getNotice(boxNotice, boxNoticeChan)
		})

		for i := 0; i < boxNoticeCount; i++ {
			boxNotices = append(boxNotices, <-boxNoticeChan)
		}

		source.BoxCount = boxCount
		_, err := Client.Collection("notice").Doc(source.FsDocID).Update(context.Background(), []firestore.Update{
			{
				Path:  "box",
				Value: source.BoxCount,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
		BoxCountMaxNumLogger.Println("boxCount =>", source.BoxCount)
	} else if boxCount < source.BoxCount {
		source.BoxCount = boxCount
		_, err := Client.Collection("notice").Doc(source.FsDocID).Update(context.Background(), []firestore.Update{
			{
				Path:  "box",
				Value: source.BoxCount,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
		BoxCountMaxNumLogger.Println("boxCount =>", source.BoxCount)
	}

	return boxNotices
}

func (source *AjouSoftwareSource) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find("#sub_contents > div > div.conbody > table:nth-child(2) > tbody > tr:nth-child(n+4):nth-last-child(n+3):nth-of-type(2n):not(:has(td:first-child > img))")
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeChan := make(chan Notice, MaxNumCount)
	numNotices := make([]Notice, 0, MaxNumCount)
	numNoticeCount := maxNum - source.MaxNum
	numNoticeCount = Min(numNoticeCount, MaxNumCount)

	if maxNum > source.MaxNum {
		numNoticeSels = numNoticeSels.FilterFunction(func(i int, _ *goquery.Selection) bool {
			return i < numNoticeCount
		})

		numNoticeSels.Each(func(_ int, numNotice *goquery.Selection) {
			go source.getNotice(numNotice, numNoticeChan)
		})

		for i := 0; i < numNoticeCount; i++ {
			numNotices = append(numNotices, <-numNoticeChan)
		}

		source.MaxNum = maxNum
		_, err = Client.Collection("notice").Doc(source.FsDocID).Update(context.Background(), []firestore.Update{
			{
				Path:  "num",
				Value: source.MaxNum,
			},
		})
		if err != nil {
			ErrorLogger.Panic(err)
		}
		BoxCountMaxNumLogger.Println("maxNum =>", source.MaxNum)
	}

	return numNotices
}

func (source *AjouSoftwareSource) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	var id string
	if sel.Find("td:nth-child(1):has(img)").Nodes != nil {
		id = "공지"
	} else {
		id = sel.Find("td:nth-child(1)").Text()
		id = strings.TrimSpace(id)
	}

	title := sel.Find("td:nth-child(3) > a").Text()
	title = strings.TrimSpace(title)
	title, _, _ = transform.String(korean.EUCKR.NewDecoder(), title)

	link, _ := sel.Find("td:nth-child(3) > a").Attr("href")
	split := strings.FieldsFunc(link, func(c rune) bool {
		return c == '&'
	})
	link = strings.Join(split[1:3], "&")
	link = strings.Join([]string{source.URL, link}, "&")

	date := sel.Find("td:nth-child(3) > p:first-of-type").Text()
	month := date[5:7]
	if month[0] == '0' {
		month = month[1:]
	}
	day := date[8:10]
	if day[0] == '0' {
		day = day[1:]
	}
	date = strings.Join([]string{month, "월", day, "일"}, "")

	notice := Notice{ID: id, Title: title, Date: date, Link: link}

	noticeChan <- notice
}

func (source *AjouSoftwareSource) sendNoticeToSlack(notice Notice) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	var footer string
	if notice.ID == "공지" {
		footer = "[중요]"
	}
	footer = strings.Join([]string{footer, "[소프트웨어학과]"}, " ")

	attachment := slack.Attachment{
		Color:      "#0072ce",
		Title:      strings.Join([]string{notice.Date, notice.Title}, " "),
		Text:       notice.Link,
		Footer:     footer,
		FooterIcon: "https://github.com/zzzang12/Notifier/assets/70265177/48fd0fd7-80e2-4309-93da-8a6bc957aacf",
	}

	_, _, err := api.PostMessage(source.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		ErrorLogger.Panic(err)
	}
}
