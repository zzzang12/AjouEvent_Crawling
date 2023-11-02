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
	//notifiers = append(notifiers, AjouNormalNotifier{}.New())
	//notifiers = append(notifiers, AjouScholarshipNotifier{}.New())
	//notifiers = append(notifiers, AjouSwNotifier{}.New())
	//notifiers = append(notifiers, AjouSoftwareNotifier{}.New())
	//notifiers = append(notifiers, AjouMediaNotifier{}.New())
	//notifiers = append(notifiers, AjouAAINotifier{}.New())
	//notifiers = append(notifiers, AjouSecurityNotifier{}.New())
	//notifiers = append(notifiers, AjouMDCNotifier{}.New())
	//notifiers = append(notifiers, AjouPharmacyNotifier{}.New())
	//notifiers = append(notifiers, AjouMedicineNotifier{}.New())
	//notifiers = append(notifiers, AjouNursingNotifier{}.New())
	//notifiers = append(notifiers, AjouECENotifier{}.New())
	//notifiers = append(notifiers, AjouITNotifier{}.New())
	//notifiers = append(notifiers, AjouAISemiNotifier{}.New())
	//notifiers = append(notifiers, AjouEngineeringNotifier{}.New())
	//notifiers = append(notifiers, AjouMENotifier{}.New())
	//notifiers = append(notifiers, AjouIENotifier{}.New())
	//notifiers = append(notifiers, AjouChembioNotifier{}.New())
	//notifiers = append(notifiers, AjouCheNotifier{}.New())
	//notifiers = append(notifiers, AjouMSENotifier{}.New())
	//notifiers = append(notifiers, AjouEnvNotifier{}.New())
	//notifiers = append(notifiers, AjouCENotifier{}.New())
	//notifiers = append(notifiers, AjouTSENotifier{}.New())
	//notifiers = append(notifiers, AjouArchNotifier{}.New())
	//notifiers = append(notifiers, AjouISENotifier{}.New())
	//notifiers = append(notifiers, AjouMobilityNotifier{}.New())
	//notifiers = append(notifiers, AjouBizNotifier{}.New())
	//notifiers = append(notifiers, AjouAbizNotifier{}.New())
	//notifiers = append(notifiers, AjouEBizNotifier{}.New())
	//notifiers = append(notifiers, AjouFENotifier{}.New())
	//notifiers = append(notifiers, AjouGBNotifier{}.New())
	//notifiers = append(notifiers, AjouNSNotifier{}.New())
	//notifiers = append(notifiers, AjouMathNotifier{}.New())
	//notifiers = append(notifiers, AjouPhysicsNotifier{}.New())
	//notifiers = append(notifiers, AjouChemNotifier{}.New())
	//notifiers = append(notifiers, AjouBiologyNotifier{}.New())
	//notifiers = append(notifiers, AjouCossNotifier{}.New())
	//notifiers = append(notifiers, AjouEconomicsNotifier{}.New())
	//notifiers = append(notifiers, AjouPbaNotifier{}.New())
	//notifiers = append(notifiers, AjouPsychologyNotifier{}.New())
	//notifiers = append(notifiers, AjouSociologyNotifier{}.New())
	//notifiers = append(notifiers, AjouPolNotifier{}.New())
	//notifiers = append(notifiers, AjouSportsLeisureNotifier{}.New())
	//notifiers = append(notifiers, AjouHumanitiesNotifier{}.New())
	//notifiers = append(notifiers, AjouKoreanNotifier{}.New())
	//notifiers = append(notifiers, AjouEnglishNotifier{}.New())
	//notifiers = append(notifiers, AjouFrenchNotifier{}.New())
	//notifiers = append(notifiers, AjouHistoryNotifier{}.New())
	//notifiers = append(notifiers, AjouCultureContentsNotifier{}.New())
	//notifiers = append(notifiers, AjouDasanNotifier{}.New())
	//notifiers = append(notifiers, AjouInternationalNotifier{}.New())
	notifiers = append(notifiers, AjouGraduateNotifier{}.New())

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
