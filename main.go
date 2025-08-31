package main

import (
	"context"
	"encoding/json"
	"fmt"
	handlers "grandfather/internal/bot"
	"grandfather/internal/ui"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New("8334069842:AAE0GvBFBFPT69R0pPAJvf8n_9PXlkEdQfs", opts...)
	if err != nil {
		panic(err)
	}

	// --- Register command handlers ---
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, handlers.StartCommandHandler)

	// b.RegisterHandler(bot.HandlerTypeMessageText, "startNewCircle", bot.MatchTypeCommand, handlers.StartNewCircleCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "joinCircle", bot.MatchTypeExact, handlers.JoinCircleCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "startNewSession", bot.MatchTypeExact, handlers.StartNewSessionCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "revealMortal", bot.MatchTypeExact, handlers.RevealMortalCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "revealAngel", bot.MatchTypeExact, handlers.RevealAngelCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "sendMessage", bot.MatchTypeExact, handlers.SendMessageCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "endSession", bot.MatchTypeExact, handlers.EndSessionCommandHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "removeUser", bot.MatchTypeExact, handlers.RemoveUserCommandHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, CallbackHandler)

	ui.RegisterMenus()

	// --- Start bot ---
	b.Start(ctx)
}

func CallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	cmd := update.CallbackQuery.Data

	// Marshal to JSON
	data, err := json.MarshalIndent(update.CallbackQuery, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling:", err)
		return
	}
	fmt.Println(string(data))

	if update.CallbackQuery.Message.Message != nil {
		msg := update.CallbackQuery.Message.Message // *models.Message

		if menu, ok := ui.GetMenu(cmd); ok {
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      msg.Chat.ID,
				MessageID:   msg.ID,
				Text:        menu.Title,
				ReplyMarkup: menu.ToInlineKeyboard(),
			})
			return
		}
	}

	// Otherwise treat as normal command
	switch cmd {
	case "startNewCircle":
		// start circle flow
		handlers.StartNewCircleCommandHandler(ctx, b, update)
	case "joinCircle":
		// start session flow
		handlers.JoinCircleCommandHandler(ctx, b, update)
	case "listCircles":
		handlers.ListCirclesCommandHandler(ctx, b, update)
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	state := handlers.UserStates[userID]

	switch state {
	case handlers.StateWaitingCircleName:
		handlers.StartNewCircleWithNameCommandHandler(ctx, b, update)
	case handlers.StateWaitingJoinCircleName:
		handlers.JoinCircleWithNameCommandHandler(ctx, b, update)
	default:

		fmt.Println("ChatID:", update.Message.Chat.ID)
		fmt.Println("Text:", update.Message.Text)

		user := update.Message.From
		fmt.Println("User ID:", user.ID)        // unique int64 ID
		fmt.Println("Username:", user.Username) // may be empty
		fmt.Println("First name:", user.FirstName)
		fmt.Println("Last name:", user.LastName)

		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Send me a command!",
		})
	}
}
