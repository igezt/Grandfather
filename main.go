package main

import (
	"context"
	"fmt"
	handlers "grandfather/internal/bot"
	"grandfather/internal/commands.go"
	"grandfather/internal/db"
	appModels "grandfather/internal/models"
	"grandfather/internal/outbox"
	"grandfather/internal/ui"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

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

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, CallbackHandler)

	commands.Router[commands.MainMenuCommand] = handlers.MainMenuCommandHandler
	commands.Router[commands.StartNewCircleCommand] = handlers.StartNewCircleCommandHandler
	commands.Router[commands.JoinCircleCommand] = handlers.JoinCircleCommandHandler
	commands.Router[commands.ListCirclesCommand] = handlers.ListCirclesCommandHandler

	ui.RegisterMenus()
	outbox := outbox.NewOutbox(ctx, b)
	defer outbox.Stop()

	b.Start(ctx)
}

func CallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	cmd := update.CallbackQuery.Data

	if !strings.Contains(cmd, "@") {
		fmt.Printf("Normal command received: %s\n", cmd)
		// Check if we have a registered handler
		if handler, ok := commands.Router[commands.Command(cmd)]; ok {
			handler(ctx, b, update)
			return
		}

		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Unknown action",
			ShowAlert:       false,
		})
		return
	}
	fmt.Printf("Dynamic command received: %s\n", cmd)

	parts := strings.SplitN(cmd, "@", 2)
	cmdPrefix, extraData := parts[0], parts[1]

	switch commands.Command(cmdPrefix) {
	case commands.GetCircleCommand:
		handlers.GetCircleDetailsHandler(ctx, b, update, extraData)
	case commands.RemoveUserCommand:
		handlers.RemoveUserCommandHandler(ctx, b, update, extraData)
	case commands.RemoveSpecificUserCommand:
		data := strings.SplitN(extraData, "@", 2)
		if len(data) != 2 {
			// Invalid payload, bail out gracefully
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Invalid command format.",
			})
			return
		}

		circleName, userIdStr := data[0], data[1]

		// Parse the user ID as int64
		userIdToRemove, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Invalid user ID.",
			})
			return
		}
		handlers.RemoveSpecificUserCommandHandler(ctx, b, update, circleName, userIdToRemove)
	case commands.GetMemberListCommand:
		handlers.GetMemberListCommandHandler(ctx, b, update, extraData)
	case commands.StartNewSessionCommand:
		handlers.StartNewSessionCommandHandler(ctx, b, update, extraData)
	case commands.EndSessionCommand:
		handlers.EndSessionCommandHandler(ctx, b, update, extraData)
	case commands.RevealMortalCommand:
		handlers.RevealMortalCommandHandler(ctx, b, update, extraData)
	case commands.RevealAngelCommand:
		handlers.RevealAngelCommandHandler(ctx, b, update, extraData)
	case commands.SendMessageCommandToAngel:
		handlers.SendMessageToAngelCommandHandler(ctx, b, update, extraData)
	case commands.SendMessageCommandToMortal:
		handlers.SendMessageToMortalCommandHandler(ctx, b, update, extraData)
	default:
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Unknown action",
			ShowAlert:       false,
		})
	}

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	if err != nil {
		log.Println("answer callback error:", err)
	}

}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	user, getUserErr := db.GetUser(ctx, userID)

	if getUserErr != nil {
		fmt.Printf("failed to get user %d: %v\n", userID, getUserErr)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Have you been registered? If you haven't, run the /start command to begin!",
		})
		return
	}

	state := user.State

	switch state {
	case appModels.StateWaitingCircleName:
		handlers.StartNewCircleWithNameCommandHandler(ctx, b, update)
	case appModels.StateWaitingJoinCircleName:
		handlers.JoinCircleWithNameCommandHandler(ctx, b, update)
	case appModels.StateWaitingSendMessageToAngel:
		handlers.SendMessageToAngelWithMessageCommandHandler(ctx, b, update, user)
	case appModels.StateWaitingSendMessageToMortal:
		handlers.SendMessageToMortalWithMessageCommandHandler(ctx, b, update, user)
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
			Text:   "Use /start to load up the menu!",
		})
	}
}
