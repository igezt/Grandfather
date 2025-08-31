package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Circle struct {
	ID      bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string        `bson:"name" json:"name"`
	OwnerId int64         `bson:"ownerId" json:"ownerId"`
	Members []int64       `bson:"members" json:"members"`
}
