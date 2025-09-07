package db

import (
	"context"
	"grandfather/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	messagesCollectionName = "messages"
)

func CreateMessage(ctx context.Context, senderId int64, receipientId int64, circleName string, message string, senderRole string) (*models.Message, error) {
	messageCollection, collErr := GetCollection(messagesCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	newMessage := &models.Message{
		ID:           bson.NewObjectID(),
		SenderId:     senderId,
		RecepientId:  receipientId,
		MessageState: models.NotDelivered,
		CircleName:   circleName,
		Message:      message,
		SenderRole:   senderRole,
	}

	_, err := messageCollection.InsertOne(ctx, newMessage)
	if err != nil {
		return nil, err
	}
	return newMessage, nil
}

func GetUndeliveredMessages(ctx context.Context) ([]*models.Message, error) {

	messageCollection, collErr := GetCollection(messagesCollectionName)
	if collErr != nil {
		return nil, collErr
	}

	filter := bson.M{"messageState": models.NotDelivered}

	cur, err := messageCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var undeliveredMessages = []*models.Message{}

	for cur.Next(ctx) {
		var u models.Message
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		undeliveredMessages = append(undeliveredMessages, &u)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return undeliveredMessages, nil
}

func UpdateMessageToDelivered(ctx context.Context, messageId bson.ObjectID) error {

	messageCollection, collErr := GetCollection(messagesCollectionName)
	if collErr != nil {
		return nil
	}

	_, err := messageCollection.UpdateOne(ctx,
		bson.M{"_id": messageId},
		bson.M{"$set": bson.M{"messageState": models.Delivered}},
	)
	return err
}
