package models

import "go.mongodb.org/mongo-driver/v2/bson"

type MessageState string

const (
	NotDelivered MessageState = "not_delivered"
	Delivered    MessageState = "delivered"
)

type Message struct {
	ID           bson.ObjectID `bson:"_id" json:"id"`
	SenderId     int64         `bson:"senderId" json:"senderId"`
	RecepientId  int64         `bson:"recepientId" json:"recepientId"`
	MessageState MessageState  `bson:"messageState" json:"messageState"`
	Message      string        `bson:"message" json:"message"`
	CircleName   string        `bson:"circleName" json:"circleName"`
	SenderRole   string        `bson:"senderRole" json:"senderRole"`
}
