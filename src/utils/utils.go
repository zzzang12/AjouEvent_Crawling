package utils

import (
	"cloud.google.com/go/firestore"
	"log"
	"os"
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

const (
	MaxNumNoticeCount int = 10
	NotifierCount     int = 4
)

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
