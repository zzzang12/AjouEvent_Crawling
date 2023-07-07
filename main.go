package main

import (
	. "Notifier/internal/notifier"
	. "Notifier/internal/utils"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

func main() {
	boxCountMaxNumLog, err := os.OpenFile("logs/boxCountMaxNumLog.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	defer boxCountMaxNumLog.Close()

	errorLog, err := os.OpenFile("logs/errorLog.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	defer errorLog.Close()

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

	exitChan := make(chan bool, 1)
	go InputExit(exitChan)

	AjouNormal = AjouNormalSource{}.New()
	AjouScholarship = AjouScholarshipSource{}.New()

	for {
		select {
		case <-exitChan:
			return
		case <-noticeTicker.C:
			log.Print("working")
			go AjouNormal.Notify()
			go AjouScholarship.Notify()

		}
	}
}
