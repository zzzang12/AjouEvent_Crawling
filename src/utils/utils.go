package utils

import (
	. "Notifier/models"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
	"os"
)

const (
	MaxNumNoticeCount = 10
	CrawlingPeriod    = 10
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
	Client, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return Client
}

func LoadConfig(fileAddress string) []NotifierConfig {
	file, err := os.Open(fileAddress)
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
