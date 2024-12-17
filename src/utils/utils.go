package utils

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "github.com/go-redis/redis/v8"
    "os"
    "fmt"
    . "Notifier/models"
    "github.com/PuerkitoBio/goquery"
    "github.com/go-sql-driver/mysql"
)

var ErrorLogger *log.Logger
var SentNoticeLogger *log.Logger
var PostLogger *log.Logger
var DB *sql.DB
var ctx = context.Background()

func CreateDir(path string) {
    _, err := os.Stat(path)
    if os.IsNotExist(err) {
        err = os.Mkdir(path, os.ModePerm)
        if err != nil {
            log.Fatal(err)
        }
    }
}

// Redis에서 crawling-token 가져오기
func GetTokenFromRedis() string {
    // Redis 클라이언트 설정
    redisHost := os.Getenv("REDIS_HOST")
    redisPort := os.Getenv("REDIS_PORT")

    rdb := redis.NewClient(&redis.Options{
        Addr: redisHost + ":" + redisPort, // 환경변수에서 Redis 호스트 정보 가져오기
    })

    // crawling-token 키로 Redis에서 토큰 가져오기
    token, err := rdb.Get(ctx, "crawling-token").Result()
    if err != nil {
        log.Fatalf("Failed to get token from Redis: %v", err)
    }
    return token
}

func OpenLogFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
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
    query := "SELECT n.value FROM notice AS n JOIN topic AS t ON n.topic_id = t.id WHERE t.department = ? AND n.type = ?"

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

// 웹훅을 호출할 때 Redis에서 가져온 토큰을 Bearer로 헤더에 추가
func SendCrawlingWebhook(url string, payload any) {
    payloadJson, err := json.Marshal(payload)
    if err != nil {
        ErrorLogger.Panic(err)
    }
	buff := bytes.NewBuffer(payloadJson)

    // Redis에서 crawling-token 가져오기
    token := GetTokenFromRedis()

    // HTTP 요청 생성
    req, err := http.NewRequest("POST", url, buff)
    if err != nil {
        ErrorLogger.Panic(err)
    }

    // Content-Type 헤더 설정
    req.Header.Set("Content-Type", "application/json")

    // Authorization 헤더에 Bearer 토큰 설정
    req.Header.Set("crawling-token", token)


    // HTTP 클라이언트로 요청 보내기
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        ErrorLogger.Panic(err)
    }
    defer resp.Body.Close()

    // 응답 본문 읽기
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

func NewDocumentFromPage(url string) (*goquery.Document, error) {
    // HTTP GET 요청을 위한 새로운 요청 생성
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        // 요청 생성 에러 발생 시 에러 반환
        ErrorLogger.Printf("Request creation error: %s", err)
        return nil, err
    }

    // User-Agent 헤더 설정
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

    // HTTP 클라이언트 생성
    client := &http.Client{}
    
    // 요청 실행
    resp, err := client.Do(req)
    if err != nil {
        // 네트워크 에러 발생 시 에러 반환
        ErrorLogger.Printf("Network error: %s", err)
        return nil, err
    }
        
    defer resp.Body.Close()

    // 응답 상태 코드 확인
    if resp.StatusCode != http.StatusOK {
        // 상태 코드 에러 발생 시 에러 반환
        ErrorLogger.Printf("Status code error: %d, URL: %s", resp.StatusCode, url)
        return nil, fmt.Errorf("status code error: %d, URL: %s", resp.StatusCode, url)
    }

    // HTML 문서 파싱
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        // 파싱 에러 발생 시 에러 반환
        ErrorLogger.Printf("Error parsing document: %s", err)
        return nil, err
    }

    return doc, nil
}