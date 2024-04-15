package main

import (
	"log"
	"os"
	"strconv"
	"time"

	. "Notifier/src/notifiers"
	. "Notifier/src/utils"
)

func main() {
	CreateDir("logs")

	errorLogFile := OpenLogFile("logs/errorLog.txt")
	defer errorLogFile.Close()
	ErrorLogger = CreateLogger(errorLogFile)

	sentNoticeLogFile := OpenLogFile("logs/sentNoticeLog.txt")
	defer sentNoticeLogFile.Close()
	SentNoticeLogger = CreateLogger(sentNoticeLogFile)

	postLogFile := OpenLogFile("logs/postLog.txt")
	defer postLogFile.Close()
	PostLogger = CreateLogger(postLogFile)

	Client = ConnectFirebase()
	defer Client.Close()

	notifierConfigs := LoadNotifierConfig("config/notifierConfigs.json")

	notifiers := make([]Notifier, 0, len(notifierConfigs))
	for _, notifierConfig := range notifierConfigs {
		baseNotifier := BaseNotifier{}.New(notifierConfig)
		var notifier Notifier
		switch notifierConfig.Type {
		case 1:
			notifier = Type1Notifier{}.New(baseNotifier)
		case 2:
			notifier = Type2Notifier{}.New(baseNotifier)
		case 3:
			notifier = Type3Notifier{}.New(baseNotifier)
		case 4:
			notifier = Type4Notifier{}.New(baseNotifier)
		case 5:
			notifier = Type5Notifier{}.New(baseNotifier)
		}
		notifiers = append(notifiers, notifier)
	}

	crawlingPeriod, err := strconv.Atoi(os.Getenv("CRAWLING_PERIOD"))
	if err != nil {
		ErrorLogger.Panic(err)
	}
	noticeTicker := time.NewTicker(time.Duration(crawlingPeriod) * time.Second)
	defer noticeTicker.Stop()

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
