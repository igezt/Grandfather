package db

import (
	"context"
	"grandfather/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	matchCollectionName = "matches"
)

func GetMortalMatch(ctx context.Context, sessionId bson.ObjectID, userId int64) (*models.Match, error) {

	coll, collErr := GetCollection(matchCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	var m models.Match
	err := coll.FindOne(ctx, bson.M{
		"session_id": sessionId,
		"angel_id":   userId,
	}).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetAngelMatch(ctx context.Context, sessionId bson.ObjectID, userId int64) (*models.Match, error) {
	coll, collErr := GetCollection(matchCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	var m models.Match
	err := coll.FindOne(ctx, bson.M{
		"session_id": sessionId,
		"mortal_id":  userId,
	}).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func CreateMatches(ctx context.Context, matches []*models.Match, sessionId bson.ObjectID) ([]*models.Match, error) {

	coll, collErr := GetCollection(matchCollectionName)
	if collErr != nil {
		return nil, collErr
	}
	if len(matches) == 0 {
		return nil, nil
	}
	docs := make([]interface{}, 0, len(matches))

	for _, m := range matches {
		m.SessionId = sessionId
		if m.ID.IsZero() {
			m.ID = bson.NewObjectID()
		}
		docs = append(docs, m)
	}
	_, err := coll.InsertMany(ctx, docs)

	if err != nil {
		return nil, err
	}

	return matches, nil
}
