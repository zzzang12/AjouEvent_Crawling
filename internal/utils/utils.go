package utils

import (
	"bufio"
	"cloud.google.com/go/firestore"
	"fmt"
	"log"
	"os"
)

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

func InputExit(exitChan chan bool) {
	var str string
	for {
		_, err := fmt.Fscanln(bufio.NewReader(os.Stdin), &str)
		if err != nil {
			log.Fatal(err)
		}
		if str == "exit" {
			exitChan <- true
		}
	}
}
