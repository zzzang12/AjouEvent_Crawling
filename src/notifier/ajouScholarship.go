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

type AjouScholarshipSource NoticeSource

func (AjouScholarshipSource) NewNotifier() *AjouScholarshipSource {
	fsDocID := "ajouScholarship"
	dsnap, err := Client.Collection("notice").Doc(fsDocID).Get(context.Background())
	if err != nil {
		ErrorLogger.Panic(err)
	}
	dbData := dsnap.Data()

	return &AjouScholarshipSource{
		BoxCount:  int(dbData["box"].(int64)),
		MaxNum:    int(dbData["num"].(int64)),
		URL:       "https://ajou.ac.kr/kr/ajou/notice_scholarship.do",
		ChannelID: "아주대학교-공지사항",
		FsDocID:   fsDocID,
	}
}

func (source *AjouScholarshipSource) Notify() {
	defer func() {
		recover()
	}()

	notices := source.scrapeNotice()
	for _, notice := range notices {
		source.sendNoticeToSlack(notice)
	}
}

func (source *AjouScholarshipSource) scrapeNotice() []Notice {
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

func (source *AjouScholarshipSource) checkHTML(doc *goquery.Document) error {
	if source.isInvalidHTML(doc) {
		errMsg := strings.Join([]string{"notifier can't work because HTML structure has changed at ", "아주대학교-장학공지사항"}, "")
		return errors.New(errMsg)
	}
	return nil
}

func (source *AjouScholarshipSource) isInvalidHTML(doc *goquery.Document) bool {
	sel1 := doc.Find("#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr[class$=\"b-top-box\"]")
	sel2 := doc.Find("#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr:not([class$=\"b-top-box\"])")
	if sel1.Nodes == nil || sel2.Nodes == nil ||
		sel1.Find("td:nth-child(1)").Nodes == nil ||
		sel1.Find("td:nth-child(2)").Nodes == nil ||
		sel1.Find("td:nth-child(3) > div > a").Nodes == nil ||
		sel1.Find("td:nth-child(3) > div > a").Nodes == nil ||
		sel1.Find("td:nth-child(5)").Nodes == nil ||
		sel1.Find("td:nth-child(6)").Nodes == nil ||
		sel2.Find("td:nth-child(1)").Nodes == nil ||
		sel2.Find("td:nth-child(2)").Nodes == nil ||
		sel2.Find("td:nth-child(3) > div > a").Nodes == nil ||
		sel2.Find("td:nth-child(3) > div > a").Nodes == nil ||
		sel2.Find("td:nth-child(5)").Nodes == nil ||
		sel2.Find("td:nth-child(6)").Nodes == nil {
		return true
	}
	return false
}

func (source *AjouScholarshipSource) scrapeBoxNotice(doc *goquery.Document) []Notice {
	boxNoticeSels := doc.Find("#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr[class$=\"b-top-box\"]")
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

func (source *AjouScholarshipSource) scrapeNumNotice(doc *goquery.Document) []Notice {
	numNoticeSels := doc.Find("#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr:not([class$=\"b-top-box\"])")
	maxNumText := numNoticeSels.First().Find("td:first-child").Text()
	maxNumText = strings.TrimSpace(maxNumText)
	maxNum, err := strconv.Atoi(maxNumText)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	numNoticeCount := min(maxNum-source.MaxNum, MaxNumNoticeCount)
	numNoticeChan := make(chan Notice, numNoticeCount)
	numNotices := make([]Notice, 0, numNoticeCount)

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

func (source *AjouScholarshipSource) getNotice(sel *goquery.Selection, noticeChan chan Notice) {
	id := sel.Find("td:nth-child(1)").Text()
	id = strings.TrimSpace(id)

	category := sel.Find("td:nth-child(2)").Text()
	category = strings.TrimSpace(category)

	title, _ := sel.Find("td:nth-child(3) > div > a").Attr("title")
	title = title[:len(title)-17]

	link, _ := sel.Find("td:nth-child(3) > div > a").Attr("href")
	split := strings.FieldsFunc(link, func(c rune) bool {
		return c == '&'
	})
	link = strings.Join(split[0:2], "&")
	link = strings.Join([]string{source.URL, link}, "")

	department := sel.Find("td:nth-child(5)").Text()

	date := sel.Find("td:nth-child(6)").Text()
	month := date[5:7]
	if month[0] == '0' {
		month = month[1:]
	}
	day := date[8:10]
	if day[0] == '0' {
		day = day[1:]
	}
	date = strings.Join([]string{month, "월", day, "일"}, "")

	notice := Notice{ID: id, Category: category, Title: title, Department: department, Date: date, Link: link}

	noticeChan <- notice
}

func (source *AjouScholarshipSource) sendNoticeToSlack(notice Notice) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	var footer string
	if notice.ID == "공지" {
		footer = "[중요]"
	}
	category := strings.Join([]string{"[", notice.Category, "]"}, "")
	department := strings.Join([]string{"[", notice.Department, "]"}, "")
	footer = strings.Join([]string{footer, category, department}, " ")

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
