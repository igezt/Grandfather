package db

import (
	"context"
	"errors"
	"fmt"
	"grandfather/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

func RemoveUserFromCircle(ctx context.Context, coll *mongo.Collection, circleId string, userId int64) (*models.Circle, error) {
	// Parse ObjectID
	oid, err := bson.ObjectIDFromHex(circleId)
	if err != nil {
		return nil, errors.New("invalid circleId")
	}

	// Build filter and update
	filter := bson.M{"_id": oid}
	update := bson.M{"$pull": bson.M{"members": userId}}

	// Return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedCircle models.Circle
	err = coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedCircle)
	if err != nil {
		return nil, err
	}

	return &updatedCircle, nil
}
