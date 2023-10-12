package main

import (
	. "Notifier/src/notifier"
	. "Notifier/src/utils"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

func main() {
	CreateDir("./logs")

	errorLog, err := os.OpenFile("logs/errorLog.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	defer errorLog.Close()

	boxCountMaxNumLog, err := os.OpenFile("logs/boxCountMaxNumLog.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	defer boxCountMaxNumLog.Close()

	sentNoticeLog, err := os.OpenFile("logs/sentNoticeLog.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	defer sentNoticeLog.Close()

	BoxCountMaxNumLogger = log.New(boxCountMaxNumLog, "", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorLog, "", log.Ldate|log.Ltime|log.Lshortfile)
	SentNoticeLogger = log.New(sentNoticeLog, "", log.Ldate|log.Ltime|log.Lshortfile)

	ctx := context.Background()
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		ErrorLogger.Fatal(err)
	}
	Client, err = app.Firestore(ctx)
	if err != nil {
		ErrorLogger.Fatal(err)
	}
	defer Client.Close()

	noticeTicker := time.NewTicker(10 * time.Second)
	defer noticeTicker.Stop()

	notifiers := make([]Notifier, 0, NotifierCount)
	notifiers = append(notifiers, CreateNotifier("AjouNormal"))
	notifiers = append(notifiers, CreateNotifier("AjouScholarship"))
	notifiers = append(notifiers, CreateNotifier("AjouSw"))
	notifiers = append(notifiers, CreateNotifier("AjouSoftware"))

	for {
		select {
		case <-noticeTicker.C:
			log.Print("working")
			for _, notifier := range notifiers {
				go notifier.Notify()
			}
		}
	}
}
