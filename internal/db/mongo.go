package db

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	clientInstance    *mongo.Client
	clientInstanceErr error
	mongoOnce         sync.Once
)

func GetMongoClient() (*mongo.Client, error) {
	getMongoConnection := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(clientOptions)
		if err != nil {
			clientInstanceErr = err
			return
		}

		// Ping to verify connection
		if err = client.Ping(ctx, nil); err != nil {
			clientInstanceErr = err
			return
		}

		log.Println("Connected to MongoDB")
		clientInstance = client
	}
	mongoOnce.Do(getMongoConnection)
	if clientInstanceErr != nil {
		getMongoConnection()
	}

	return clientInstance, clientInstanceErr
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}
	database := client.Database("grandfather")
	collection := database.Collection(collectionName)
	return collection, err
}
