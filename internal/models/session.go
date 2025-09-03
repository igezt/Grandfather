package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SessionState string

const (
	StateActive   SessionState = "active"
	StateFinished SessionState = "inactive"
)

type Session struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CircleId  bson.ObjectID `bson:"circleId" json:"circleId"`
	Members   []int64       `bson:"members" json:"members"`
	State     SessionState  `bson:"state" json:"state"`
	CreatedAt time.Time     `bson:"time" json:"time"`
}
