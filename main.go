package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"os"
	"strings"
)

type Notice struct {
	id         string
	category   string
	title      string
	department string
	date       string
	link       string
}

func main() {
	notices := scrapeNotice()
	notices = notices[:1]
	for _, notice := range notices {
		sendMessageToSlack(notice)
	}
}

func sendMessageToSlack(notice Notice) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	attachment := slack.Attachment{
		Title: strings.Join([]string{notice.date, notice.title}, " "),
		Text:  notice.link,
	}
	_, _, err := api.PostMessage("아주대학교-공지사항", slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Fatal(err)
	}
}

func scrapeNotice() []Notice {
	noticeURL := "https://ajou.ac.kr/kr/ajou/notice.do"

	resp, err := http.Get(noticeURL)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %s", resp.Status)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sel := doc.Find("#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr")
	noticeNum := sel.Length()
	noticeChan := make(chan Notice, noticeNum)
	sel.Each(func(_ int, sel *goquery.Selection) {
		go getNotice(noticeURL, sel, noticeChan)
	})

	notices := make([]Notice, 0, noticeNum)
	for i := 0; i < noticeNum; i++ {
		notices = append(notices, <-noticeChan)
	}

	return notices
}

func getNotice(noticeURL string, sel *goquery.Selection, noticeChan chan Notice) {
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
	link = strings.Join([]string{noticeURL, link}, "")

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

	notice := Notice{id, category, title, department, date, link}

	noticeChan <- notice
}
