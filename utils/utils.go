package utils

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

func ExtractUserAndChat(update *models.Update) (*models.User, int64, error) {
	if update.Message != nil {
		// Normal message
		return update.Message.From, update.Message.Chat.ID, nil
	}

	if update.CallbackQuery != nil {
		// Button press
		msg := update.CallbackQuery.Message
		if msg.Message != nil {
			// Message is accessible
			return &update.CallbackQuery.From, msg.Message.Chat.ID, nil
		}
		return &update.CallbackQuery.From, 0, fmt.Errorf("callback has no accessible chat")
	}

	return nil, 0, fmt.Errorf("no user/chat found in update")
}
