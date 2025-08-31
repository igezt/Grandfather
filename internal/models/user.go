package models

type User struct {
	ID     int64 `bson:"_id" json:"id"`         // Telegram user ID as the primary key
	ChatID int64 `bson:"chat_id" json:"chatId"` // Chat ID (can differ from user ID, esp. groups)
}
