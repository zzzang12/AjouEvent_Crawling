package utils

import (
	"cloud.google.com/go/firestore"
	"log"
	"os"
)

const (
	MaxNumNoticeCount = 10
	NotifierCount     = 52
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
