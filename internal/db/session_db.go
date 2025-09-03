package db

import (
	"context"
	"errors"
	"grandfather/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	sessionCollectionName = "sessions"
)

func CreateSession(ctx context.Context, circleId bson.ObjectID, members []int64) (*models.Session, error) {

	sessionCollection, collErr := GetCollection(sessionCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	newSession := &models.Session{
		ID:        bson.NewObjectID(),
		CircleId:  circleId,
		Members:   members,
		State:     models.StateActive,
		CreatedAt: time.Now(),
	}

	_, err := sessionCollection.InsertOne(ctx, newSession)
	if err != nil {
		return nil, err
	}
	return newSession, nil
}

func GetSession(ctx context.Context, sessionId bson.ObjectID) (*models.Session, error) {
	sessionCollection, collErr := GetCollection(sessionCollectionName)
	if collErr != nil {
		return nil, collErr
	}
	filter := bson.M{
		"_id": sessionId,
	}

	var session models.Session
	err := sessionCollection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func UpdateSessionToFinished(ctx context.Context, sessionId bson.ObjectID) (*models.Session, error) {
	sessionCollection, collErr := GetCollection(sessionCollectionName)
	if collErr != nil {
		return nil, collErr
	}
	filter := bson.M{"_id": sessionId}
	update := bson.M{
		"$set": bson.M{
			"state": models.StateFinished,
			// "updatedAt": time.Now(), // optional: track update time
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.Session
	err := sessionCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func DeleteSessionByID(ctx context.Context, sessionId bson.ObjectID) (*models.Session, error) {
	coll, collErr := GetCollection(sessionCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	var deleted models.Session
	err := coll.FindOneAndDelete(ctx, bson.M{"_id": sessionId}).Decode(&deleted)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // nothing found
		}
		return nil, err
	}

	return &deleted, nil
}
