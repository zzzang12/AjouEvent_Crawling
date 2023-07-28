package utils

import (
	"cloud.google.com/go/firestore"
	"github.com/slack-go/slack"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Notifier interface {
	Notify()
}

type Notice struct {
	ID         string
	Category   string
	Title      string
	Department string
	Date       string
	Link       string
}

type NoticeSource struct {
	BoxCount  int
	MaxNum    int
	URL       string
	ChannelID string
	FsDocID   string
}

const MaxNumCount int = 10
const NotifierCount int = 4

var BoxCountMaxNumLogger *log.Logger
var ErrorLogger *log.Logger
var SentNoticeLogger *log.Logger
var Client *firestore.Client

func CreateDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir("logs", os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func SendErrorToSlack(err error) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	errorTime := time.Now().Format("2006/01/02 15:04:05")
	_, errorFile, errorLine, _ := runtime.Caller(1)
	errorFilePath := strings.Split(errorFile, "/")
	errorFile = errorFilePath[len(errorFilePath)-1]
	errorFile = strings.Join([]string{errorFile, strconv.Itoa(errorLine)}, ":")
	title := strings.Join([]string{errorTime, errorFile}, " ")

	attachment := slack.Attachment{
		Color:      "#cc0000",
		Title:      title,
		Text:       err.Error(),
		Footer:     "[에러]",
		FooterIcon: "https://github.com/zzzang12/Notifier/assets/70265177/48fd0fd7-80e2-4309-93da-8a6bc957aacf",
	}

	_, _, err = api.PostMessage("에러", slack.MsgOptionAttachments(attachment))
	if err != nil {
		ErrorLogger.Fatal(err)
	}
}
