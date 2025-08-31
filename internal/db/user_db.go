package db

import (
	"context"
	"fmt"
	"grandfather/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	userCollectionName = "users"
)

func CreateUser(ctx context.Context, userId int64, chatId int64) (*models.User, error, bool) {
	coll, collErr := GetCollection(userCollectionName)
	if collErr != nil {
		return nil, collErr, false
	}

	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"chat_id": chatId,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	res, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err, false
	}

	alreadyCreatedUser := res.MatchedCount > 0
	if res.MatchedCount > 0 {
		fmt.Println("User already exists, updated chatId")
	} else if res.UpsertedCount > 0 {
		fmt.Println("Inserted new user with ID:", res.UpsertedID)
	}

	return &models.User{
		ID:     userId,
		ChatID: chatId,
	}, nil, alreadyCreatedUser
}
