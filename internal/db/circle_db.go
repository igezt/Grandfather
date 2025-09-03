package db

import (
	"context"
	"fmt"
	"grandfather/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	circleCollectionName = "circles"
)

func CreateCircle(ctx context.Context, circleName string, circleOwner int64) (*models.Circle, error) {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	circle := models.Circle{
		Name:    circleName,
		OwnerId: circleOwner,
		Members: []int64{circleOwner},
	}

	res, err := coll.InsertOne(ctx, circle)

	if err != nil {
		return nil, err
	}

	// Type assertion
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		circle.ID = oid
	} else {
		return nil, fmt.Errorf("failed to assert InsertedID to ObjectID, got %T", res.InsertedID)
	}

	return &circle, nil
}

func GetCircle(ctx context.Context, circleName string) (*models.Circle, error) {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	filter := bson.M{
		"name": circleName,
	}

	var circle models.Circle
	res := coll.FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		return nil, err
	}

	// Decode the found document into our struct
	if err := res.Decode(&circle); err != nil {
		return nil, err
	}

	return &circle, nil
}

func GetCircles(ctx context.Context, userId int64) ([]models.Circle, error) {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return nil, collErr
	}
	filter := bson.M{
		"members": userId,
	}
	cursor, err := coll.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	var circles []models.Circle
	if err := cursor.All(ctx, &circles); err != nil {
		return nil, err
	}

	return circles, nil
}

func AddUserToCircle(ctx context.Context, circleId bson.ObjectID, userId int64) (*models.Circle, error) {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return nil, collErr
	}
	fmt.Printf("%s\n", circleId)

	filter := bson.M{"_id": circleId}
	update := bson.M{"$addToSet": bson.M{"members": userId}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedCircle models.Circle
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedCircle)
	if err != nil {
		return nil, err
	}

	return &updatedCircle, nil
}

func RemoveUserFromCircle(ctx context.Context, circleId bson.ObjectID, userId int64) (*models.Circle, error) {

	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	// Build filter and update
	filter := bson.M{"_id": circleId}
	update := bson.M{"$pull": bson.M{"members": userId}}

	// Return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedCircle models.Circle
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedCircle)
	if err != nil {
		return nil, err
	}

	return &updatedCircle, nil
}

func SetCircleCurrentSession(ctx context.Context, circleId bson.ObjectID, sessionId bson.ObjectID) error {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return collErr
	}

	update := bson.M{
		"$set": bson.M{
			"currentSession": sessionId,
		},
	}

	_, err := coll.UpdateByID(ctx, circleId, update)
	if err != nil {
		return err
	}

	return nil
}

func UnsetCircleCurrentSession(ctx context.Context, circleId, sessionId bson.ObjectID) (bool, error) {
	coll, collErr := GetCollection(circleCollectionName)
	if collErr != nil {
		return false, collErr
	}

	// Only unset if the currentSession matches the given sessionId
	filter := bson.M{
		"_id":            circleId,
		"currentSession": sessionId,
	}
	update := bson.M{
		"$unset": bson.M{
			"currentSession": "",
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}

	// MatchedCount == 1 means we successfully unset
	return result.MatchedCount == 1, nil
}
