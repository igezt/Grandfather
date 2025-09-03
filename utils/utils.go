package utils

import (
	"context"
	"fmt"
	"grandfather/internal/ui"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/go-telegram/bot"
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

func SendErrorMessage(ctx context.Context, b *bot.Bot, chatId int64) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   "Oops, something went wrong. Please try again later.",
	})
}

func SendCustomErrorMessage(ctx context.Context, b *bot.Bot, chatId int64, errorMessage string) {

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   errorMessage,
	})
}

func EditToMenu(ctx context.Context, b *bot.Bot, messageId int, chatID int64, menu ui.Menu) {
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageId,
		Text:        menu.Title,
		ReplyMarkup: menu.ToInlineKeyboard(),
	})
}

func SendMenu(ctx context.Context, b *bot.Bot, chatID int64, menu ui.Menu) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        menu.Title,
		ReplyMarkup: menu.ToInlineKeyboard(),
	})
}

func IsValidOnlyAlphanumericAndSpaces(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ') {
			return false
		}
	}
	return true
}

func ShuffleInt64(a []int64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(a) - 1; i > 0; i-- {
		j := r.Intn(i + 1) // random index from 0..i
		a[i], a[j] = a[j], a[i]
	}
}

func EscapeMarkdownV2(text string) string {
	specials := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, s := range specials {
		text = strings.ReplaceAll(text, s, "\\"+s)
	}
	return text
}
