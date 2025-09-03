package db

import (
	"context"
	"fmt"
	"grandfather/internal/models"
	"log"

	tlgModels "github.com/go-telegram/bot/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	userCollectionName = "users"
)

func GetUser(ctx context.Context, userId int64) (*models.User, error) {
	filter := bson.M{"_id": userId}

	coll, collErr := GetCollection(userCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	var user models.User
	err := coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No user found
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func CreateUser(ctx context.Context, user *tlgModels.User, chatId int64) (*models.User, error, bool) {
	userId := user.ID

	coll, collErr := GetCollection(userCollectionName)
	if collErr != nil {
		return nil, collErr, false
	}

	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"chat_id":     chatId,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"user_handle": user.Username,
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

func GetUsers(ctx context.Context, userIds []int64) ([]*models.User, error) {

	if len(userIds) == 0 {
		return []*models.User{}, nil
	}

	coll, err := GetCollection(userCollectionName)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bson.M{"$in": userIds}}

	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	foundByID := make(map[int64]*models.User, len(userIds))

	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		foundByID[u.ID] = &u
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	ordered := make([]*models.User, 0, len(userIds))
	for _, id := range userIds {
		if u, ok := foundByID[id]; ok {
			ordered = append(ordered, u)
		}
	}

	return ordered, nil
}

func UpdateState(ctx context.Context, userId int64, state models.UserState) error {

	coll, err := GetCollection(userCollectionName)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userId}
	update := bson.M{
		"$set": bson.M{
			"state": state,
		},
	}

	opts := options.UpdateOne().SetUpsert(false)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		log.Printf("No user found with id %d", userId)
	}

	return nil
}
