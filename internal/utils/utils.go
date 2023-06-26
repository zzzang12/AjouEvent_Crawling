package utils

import "cloud.google.com/go/firestore"

const MaxNumCount int = 10

var Client *firestore.Client

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
