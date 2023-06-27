package utils

import (
	"cloud.google.com/go/firestore"
	"log"
)

const MaxNumCount int = 10

var BoxCountMaxNumLogger *log.Logger
var ErrorLogger *log.Logger
var SentNoticeLogger *log.Logger
var Client *firestore.Client

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
