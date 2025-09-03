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
	clientOnce sync.Once
	client     *mongo.Client
	clientErr  error
)

func connect() {
	uri := "mongodb://127.0.0.1:27017/?directConnection=true"

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(5 * time.Second). // how long to wait for a suitable server
		SetConnectTimeout(5 * time.Second)          // dial timeout

	cli, err := mongo.Connect(clientOpts)
	if err != nil {
		clientErr = err
		return
	}

	// Verify with a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.Ping(ctx, nil); err != nil {
		clientErr = err
		_ = cli.Disconnect(context.Background())
		return
	}

	log.Println("Connected to MongoDB")
	client = cli
}

func GetMongoClient() (*mongo.Client, error) {
	clientOnce.Do(connect)
	return client, clientErr
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	cli, err := GetMongoClient()
	if err != nil {
		return nil, err
	}
	return cli.Database("grandfather").Collection(collectionName), nil
}
