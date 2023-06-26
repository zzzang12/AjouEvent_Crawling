package notifier

import "cloud.google.com/go/firestore"

const MaxNumCount int = 10

var Client *firestore.Client
