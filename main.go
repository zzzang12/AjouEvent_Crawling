package main

import (
	. "Notifier/src/notifiers"
	. "Notifier/src/utils"
	"log"
	"time"
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

	notifierConfigs := LoadConfig("notifierConfigs.json")

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

	noticeTicker := time.NewTicker(CrawlingPeriod * time.Second)
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
