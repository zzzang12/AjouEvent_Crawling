package utils

import (
	"Notifier/src/models"
	"cloud.google.com/go/firestore"
	"encoding/json"
	"log"
	"os"
)

const (
	MaxNumNoticeCount = 10
	CrawlingPeriod    = 10
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

func LoadConfig(fileAddress string) ([]models.NotifierConfig, error) {
	var manifests []models.NotifierConfig
	file, err := os.Open(fileAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&manifests)
	if err != nil {
		log.Fatal(err)
	}
	return manifests, err
}

func CreateSelectorMap() {
	selectorMap := make(map[int]map[string]interface{})
	selectorMap[1] = map[string]interface{}{
		"": "",
	}
	selectorMap[2] = map[string]interface{}{
		"box": "#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr[class$=\"b-top-box\"]",
		"num": "#cms-content > div > div > div.bn-list-common02.type01.bn-common-cate > table > tbody > tr:not([class$=\"b-top-box\"])",
	}
}
