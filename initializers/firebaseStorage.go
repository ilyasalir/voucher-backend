package initializers

import (
	"context"
	"time"

	storage2 "cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"google.golang.org/api/option"
)

// const firebaseConfigPath = "firebase-carport.json"
const firebaseConfigPath = "firebase-carport.json"

var FirebaseApp *firebase.App
var StorageClient *storage.Client

func initContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return ctx
}

func InitFirebaseStorage() (*storage2.Client, error) {
	ctx := initContext()
	opt := option.WithCredentialsFile(firebaseConfigPath)
	client, err := storage2.NewClient(ctx, opt)
	if err != nil {
		return nil, err
	}
	return client, nil
}
