package main

import (
	"log"
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

	Client = ConnectFirebase()
	defer Client.Close()

	env := LoadEnv()

	notifierConfigs := LoadNotifierConfig("config/notifierConfigs.json")

	notifiers := make([]Notifier, 0, len(notifierConfigs))
	for _, notifierConfig := range notifierConfigs {
		var notifier Notifier
		switch notifierConfig.Type {
		case 1:
			notifier = Type1Notifier{}.New(notifierConfig)
		case 2:
			notifier = Type2Notifier{}.New(notifierConfig)
		case 3:
			notifier = Type3Notifier{}.New(notifierConfig)
		case 4:
			notifier = Type4Notifier{}.New(notifierConfig)
		}
		notifiers = append(notifiers, notifier)
	}

	crawlingPeriod, err := strconv.Atoi(env["CRAWLING_PERIOD"])
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
