package handlers

import (
	"context"
	"fmt"
	"grandfather/internal/db"
	"grandfather/internal/ui"
	"grandfather/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UserState string

const (
	StateNone                  UserState = ""
	StateWaitingCircleName     UserState = "waiting_circle_name"
	StateWaitingJoinCircleName UserState = "waiting_join_circle_name"
)

var UserStates = map[int64]UserState{} // map[telegramID]state

func StartCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("start")

	user := update.Message.From
	chatId := update.Message.Chat.ID
	_, err, alreadyCreatedUser := db.CreateUser(ctx, user.ID, chatId)
	if err != nil {
		fmt.Printf("failed to create user: %v\n", err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Oops, something went wrong. Please try again later.",
		})
		return
	}

	if alreadyCreatedUser {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You have already been registered in the past 🎉",
		})
	} else {
		// Successful insertion
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Welcome! You’ve been registered 🎉",
		})
	}

	mainMenu, _ := ui.GetMenu("main")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        mainMenu.Title,
		ReplyMarkup: mainMenu.ToInlineKeyboard(),
	})
}

func StartNewCircleCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Start new circle")
	user, chatID, err := utils.ExtractUserAndChat(update)
	if err != nil {
		fmt.Println("Error extracting user/chat:", err)
		return
	}
	UserStates[user.ID] = StateWaitingCircleName

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Great! What would you like to name your circle?",
	})
}

func StartNewCircleWithNameCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Start new circle with name")
	user, _, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		return
	}
	circleName := update.Message.Text
	_, err := db.CreateCircle(ctx, circleName, user.ID)

	// TODO: Give custom message when we see duplicate key error
	if err != nil {
		fmt.Printf("failed to create circle %s: %v\n", circleName, err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Oops, something went wrong. Please try again later.",
		})
		return
	}

	UserStates[user.ID] = StateNone

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Your circle %q has been created!", circleName),
	})
}

func JoinCircleCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Joining circle")

	user, chatID, err := utils.ExtractUserAndChat(update)
	if err != nil {
		fmt.Println("Error extracting user/chat:", err)
		return
	}
	UserStates[user.ID] = StateWaitingJoinCircleName

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Great! What is the name of the circle you want to join?",
	})
}

func JoinCircleWithNameCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Joining circle with name")

	user, _, extractErr := utils.ExtractUserAndChat(update)

	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		return
	}

	circleName := update.Message.Text
	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Oops, something went wrong. Please try again later.",
		})
		return
	}

	updatedCircle, updateCircleErr := db.AddUserToCircle(ctx, circle.ID, user.ID)
	// TODO: Give custom message when we see not found key error
	if updateCircleErr != nil {
		fmt.Printf("failed to update circle %s: %v\n", circleName, updateCircleErr)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Oops, something went wrong. Please try again later.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("You have joined the circle %s", updatedCircle.Name),
	})

}

func ListCirclesCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Listing circles")
	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		return
	}
	circles, getCirclesErr := db.GetCircles(ctx, user.ID)

	if getCirclesErr != nil {
		fmt.Printf("Error getting circles for user: %d\n", user.ID)
		return
	}

	circlesMenu, _ := ui.GetMenu("circles")
	for _, circle := range circles {
		circlesMenu.PrependButtonRow(circle.Name, "getcirclesaction-"+circle.Name)
	}

	// b.SendMessage(ctx, &bot.SendMessageParams{
	// 	ChatID:      chatID,
	// 	Text:        circlesMenu.Title,
	// 	ReplyMarkup: circlesMenu.ToInlineKeyboard(),
	// })

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        circlesMenu.Title,
		ReplyMarkup: circlesMenu.ToInlineKeyboard(),
	})

}

func RemoveUserCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Remove user")
}

func StartNewSessionCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Starting new session")

}

func RevealMortalCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Reveal mortal")

}

func RevealAngelCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Reveal angel")

}

func SendMessageCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Send message")

}

func EndSessionCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("End Session")

}
