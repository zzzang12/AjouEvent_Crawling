package utils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	. "Notifier/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-sql-driver/mysql"
)

var ErrorLogger *log.Logger
var SentNoticeLogger *log.Logger
var PostLogger *log.Logger
var DB *sql.DB

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

func ConnectDB() *sql.DB {
	config := mysql.Config{
		User:                 os.Getenv("DB_USER"),
		Passwd:               os.Getenv("DB_PW"),
		Net:                  "tcp",
		Addr:                 os.Getenv("DB_IP") + ":" + os.Getenv("DB_PORT"),
		DBName:               os.Getenv("DB_NAME"),
		AllowNativePasswords: true,
	}
	connector, err := mysql.NewConnector(&config)
	if err != nil {
		ErrorLogger.Panic(err)
	}

	db := sql.OpenDB(connector)
	err = db.Ping()
	if err != nil {
		ErrorLogger.Panic(err)
	}

	return db
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

func LoadDbData(topic string) (int, int) {
	var boxCount int
	query := "SELECT n.value FROM notice AS n JOIN topic AS t ON n.topic_id = t.id WHERE t.topic = ? AND n.type = ?"

	err := DB.QueryRow(query, topic, "box").Scan(&boxCount)
	if err != nil {
		log.Fatal(err)
	}

	var maxNum int
	err = DB.QueryRow(query, topic, "num").Scan(&maxNum)
	if err != nil {
		log.Fatal(err)
	}

	return boxCount, maxNum
}

func SendCrawlingWebhook(url string, payload any) {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	buff := bytes.NewBuffer(payloadJson)

	resp, err := http.Post(url, "application/json", buff)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ErrorLogger.Panic(err)
	}
	PostLogger.Println(string(body))
}

func GetNumNoticeCountReference(doc *goquery.Document, englishTopic, boxNoticeSelector string) int {
	if englishTopic != "Software" {
		return 10
	}
	boxNoticeSels := doc.Find(boxNoticeSelector)
	boxCount := boxNoticeSels.Length()
	return 15 - boxCount
}
