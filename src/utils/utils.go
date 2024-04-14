package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	. "Notifier/models"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var ErrorLogger *log.Logger
var SentNoticeLogger *log.Logger
var Client *firestore.Client

func CreateDir(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func OpenLogFile(path string) *os.File {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func CreateLogger(file *os.File) *log.Logger {
	return log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func ConnectFirebase() *firestore.Client {
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatal(err)
	}
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func LoadNotifierConfig(path string) []NotifierConfig {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var configs []NotifierConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configs)
	if err != nil {
		log.Fatal(err)
	}

	return configs
}

func LoadDbData(documentID string) map[string]interface{} {
	dsnap, err := Client.Collection("notice").Doc(documentID).Get(context.Background())
	if err != nil {
		ErrorLogger.Panic(err)
	}
	dbData := dsnap.Data()
	return dbData
}

func LoadEnv() map[string]string {
	env := make(map[string]string)
	env["CRAWLING_PERIOD"] = os.Getenv("CRAWLING_PERIOD")
	return env
}

func SendCrawlingWebhook(url string, payload any) {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	buff := bytes.NewBuffer(payloadJson)

	_, err = http.Post(url, "application/json", buff)
	if err != nil {
		ErrorLogger.Panic(err)
	}
}
