package main

import (
	"Notifier/internal/notifier"
	"Notifier/internal/utils"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"time"
)

func main() {
	start := time.Now()

	ctx := context.Background()
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatal(err)
	}
	utils.Client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer utils.Client.Close()

	notifier.AjouNormalBoxCount, notifier.AjouNormalMaxNum = notifier.GetAjouNormalFromDB()

	notifier.AjouNormalFunc(notifier.AjouNormalBoxCount, notifier.AjouNormalMaxNum)

	end := time.Since(start)
	fmt.Println("end =>", end)
}
