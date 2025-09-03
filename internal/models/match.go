package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Match struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	SessionId bson.ObjectID `bson:"session_id" json:"session_id"`
	AngelId   int64         `bson:"angel_id" json:"angel_id"`
	MortalId  int64         `bson:"mortal_id" json:"mortal_id"`
}
