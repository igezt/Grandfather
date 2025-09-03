package handlers

import (
	"context"
	"errors"
	"fmt"
	"grandfather/internal/commands.go"
	"grandfather/internal/db"
	"grandfather/internal/ui"
	"grandfather/utils"
	"slices"
	"strings"

	appModels "grandfather/internal/models"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func MainMenuCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("main menu")
	_, chatID, extractErr := utils.ExtractUserAndChat(update)

	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}
	msg := update.CallbackQuery.Message.Message // *models.Message

	if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			Text:        menu.Title,
			ReplyMarkup: menu.ToInlineKeyboard(),
		})
		return
	}
	utils.SendErrorMessage(ctx, b, chatID)
}

func StartCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("start")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)

	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	_, err, alreadyCreatedUser := db.CreateUser(ctx, user, chatID)
	if err != nil {
		fmt.Printf("failed to create user: %v\n", err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if alreadyCreatedUser {
		// _, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		// 	ChatID: update.Message.Chat.ID,
		// 	Text:   "You have already been registered in the past 🎉",
		// })
	} else {
		// Successful insertion
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Welcome! You’ve been registered 🎉",
		})
	}

	mainMenu, _ := ui.GetMenu(ui.MenuNameMain)

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
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	updateUserStateErr := db.UpdateState(ctx, user.ID, appModels.StateWaitingCircleName)

	if updateUserStateErr != nil {
		fmt.Println("Error updating user state:", err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Great! What would you like to name your circle?",
	})
}

func StartNewCircleWithNameCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Start new circle with name")
	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}
	circleName := update.Message.Text

	if !utils.IsValidOnlyAlphanumericAndSpaces(circleName) {
		utils.SendCustomErrorMessage(ctx, b, chatID, "Your circle name must not contain something other than alphabets, numbers and spaces!")
		return
	}

	circle, err := db.CreateCircle(ctx, circleName, user.ID)

	// TODO: Give custom message when we see duplicate key error
	if err != nil {
		fmt.Printf("failed to create circle %s: %v\n", circleName, err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	updateUserStateErr := db.UpdateState(ctx, user.ID, appModels.StateNone)

	if updateUserStateErr != nil {
		fmt.Println("Error updating user state:", err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Your circle %q has been created!", circleName),
	})

	circleMenu := circle.ToMenu(user.ID)

	utils.SendMenu(ctx, b, chatID, circleMenu)

}

func JoinCircleCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Joining circle")

	user, chatID, err := utils.ExtractUserAndChat(update)
	if err != nil {
		fmt.Println("Error extracting user/chat:", err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	updateUserStateErr := db.UpdateState(ctx, user.ID, appModels.StateWaitingJoinCircleName)

	if updateUserStateErr != nil {
		fmt.Println("Error updating user state:", err)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Great! What is the name of the circle you want to join?",
	})
}

func JoinCircleWithNameCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Joining circle with name")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)

	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circleName := update.Message.Text
	circle, getCircleErr := db.GetCircle(ctx, circleName)

	// TODO: Give custom message when we see not found key error
	if getCircleErr != nil {
		if errors.Is(getCircleErr, mongo.ErrNoDocuments) {
			// Custom user-friendly message
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("No circle found with the name '%s'.", circleName),
			})

			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	updatedCircle, updateCircleErr := db.AddUserToCircle(ctx, circle.ID, user.ID)
	if updateCircleErr != nil {
		fmt.Printf("failed to update circle %s: %v\n", circleName, updateCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("You have joined the circle %s", updatedCircle.Name),
	})

	circleMenu := circle.ToMenu(user.ID)

	utils.SendMenu(ctx, b, chatID, circleMenu)
}

func ListCirclesCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("Listing circles")
	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}
	circles, getCirclesErr := db.GetCircles(ctx, user.ID)

	if getCirclesErr != nil {
		fmt.Printf("Error getting circles for user: %d\n", user.ID)
		return
	}

	circlesMenu, _ := ui.GetMenu("circles")
	for _, circle := range circles {
		circlesMenu.PrependButtonRow(circle.Name, string(commands.GetCircleCommand)+"@"+circle.Name)
	}

	utils.EditToMenu(ctx, b, update.CallbackQuery.Message.Message.ID, chatID, circlesMenu)
}

func GetCircleDetailsHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Getting details of a circle")
	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		fmt.Println("Could not find the circle specified.")
		utils.SendCustomErrorMessage(ctx, b, chatID, "Could not find the circle specified. Are you sure the circle still exists?")
		return
	}

	memberIds := circle.Members
	if !slices.Contains(memberIds, user.ID) {
		fmt.Println("User is not a member of the circle.")
		utils.SendCustomErrorMessage(ctx, b, chatID, "You don't seem to be a part of this circle!")
		return
	}

	circleMenu := circle.ToMenu(user.ID)

	utils.EditToMenu(ctx, b, update.CallbackQuery.Message.Message.ID, chatID, circleMenu)
}

func GetMemberListCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Getting member list of a circle")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	// TODO: Give custom message when we see not found key error
	if getCircleErr != nil {
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("Circle %s was not found!", circleName))
		return
	}

	memberIds := circle.Members
	if !slices.Contains(memberIds, user.ID) {
		fmt.Println("User is not a member of the circle.")
		utils.SendCustomErrorMessage(ctx, b, chatID, "You don't seem to be a part of this circle!")
		return
	}

	members, getUsersErr := db.GetUsers(ctx, circle.Members)

	if getUsersErr != nil {
		fmt.Printf("failed to get users for circle %s: %v\n", circleName, getUsersErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	names := make([]string, 0, len(members))
	for _, m := range members {
		if m.UserHandle != "" {
			names = append(names, "@"+m.UserHandle)
		} else {
			names = append(names, m.FirstName)
		}
	}

	memberList := strings.Join(names, ", ")

	title := fmt.Sprintf("👥 Circle: %s\nMembers: %s", circleName, memberList)

	membersMenu := ui.Menu{
		Title:   title,
		Buttons: [][]ui.MenuButton{},
	}
	membersMenu.AddButtonRow("Remove Member", string(commands.RemoveUserCommand)+"@"+circle.Name)
	membersMenu.AddButtonRow("Back", string(commands.GetCircleCommand)+"@"+circle.Name)
	utils.EditToMenu(ctx, b, update.CallbackQuery.Message.Message.ID, chatID, membersMenu)
}

func RemoveUserCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Remove user")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	// TODO: Give custom message when we see not found key error
	if getCircleErr != nil {
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("Circle %s was not found!", circleName))
		return
	}

	ownerId := circle.OwnerId
	if ownerId != user.ID {
		fmt.Printf("Non-owner tried to remove member %s: %v\n", circleName, user.Username)
		utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("You cannot carry out this action. It doesn't seem like you are the owner of the circle %s!", circleName))
		return
	}

	members, getUsersErr := db.GetUsers(ctx, circle.Members)

	if getUsersErr != nil {
		fmt.Printf("failed to get users for circle %s: %v\n", circleName, getUsersErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	title := fmt.Sprintf("👥 Circle: %s\nWho would you like to remove?", circleName)

	removeMembersMenu := ui.Menu{
		Title:   title,
		Buttons: [][]ui.MenuButton{},
	}

	for _, member := range members {
		removeMembersMenu.AddButtonRow(fmt.Sprintf("%s %s @%s", member.FirstName, member.LastName, member.UserHandle), fmt.Sprintf("%s@%s@%d", string(commands.RemoveSpecificUserCommand), circleName, member.ID))
	}
	removeMembersMenu.AddButtonRow("Back", string(commands.GetCircleCommand)+"@"+circle.Name)
	utils.EditToMenu(ctx, b, update.CallbackQuery.Message.Message.ID, chatID, removeMembersMenu)
}

func RemoveSpecificUserCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string, userIdToRemove int64) {
	fmt.Println("Removing specific user")
	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	// TODO: Give custom message when we see not found key error
	if getCircleErr != nil {
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("Circle %s was not found!", circleName))
		return
	}

	ownerId := circle.OwnerId
	if ownerId != user.ID {
		fmt.Printf("Non-owner tried to remove member %s: %v\n", circleName, user.Username)
		utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("You cannot carry out this action. It doesn't seem like you are the owner of the circle %s!", circleName))
		return
	}

	if userIdToRemove == ownerId {
		fmt.Printf("Owner tried to remove themselves %s: %v\n", circleName, user.Username)
		utils.SendCustomErrorMessage(ctx, b, chatID, "The owner cannot be removed from the circle!")
		circleMenu := circle.ToMenu(user.ID)
		utils.SendMenu(ctx, b, chatID, circleMenu)
		return
	}

	_, updatedCircleErr := db.RemoveUserFromCircle(ctx, circle.ID, userIdToRemove)

	if updatedCircleErr != nil {
		fmt.Printf("failed to remove user for circle %s: %v\n", circleName, updatedCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	utils.SendCustomErrorMessage(ctx, b, chatID, fmt.Sprintf("User has been removed from %s", circleName))

	circleMenu := circle.ToMenu(user.ID)
	utils.SendMenu(ctx, b, chatID, circleMenu)
}

func StartNewSessionCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Starting new session")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)

	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		if errors.Is(getCircleErr, mongo.ErrNoDocuments) {
			// Custom user-friendly message
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("No circle found with the name '%s'.", circleName),
			})

			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if !slices.Contains(circle.Members, user.ID) {
		fmt.Println("User is not a member of the circle.")
		utils.SendCustomErrorMessage(ctx, b, chatID, "You don't seem to be a part of this circle!")
		return
	}

	if circle.CurrentSession != nil {
		if s, err := db.GetSession(ctx, *circle.CurrentSession); err == nil {
			if s.State == appModels.StateActive {
				_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "There’s already an active session running for this circle. You can’t start a new one until it ends.",
				})
				return
			}
		}
	}

	session, createSessionErr := db.CreateSession(ctx, circle.ID, circle.Members)

	if createSessionErr != nil {
		fmt.Printf("failed to create sessions for circle %s: %v\n", circleName, createSessionErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	// TODO: create matches
	memberIds := circle.Members

	utils.ShuffleInt64(memberIds)

	matches := make([]*appModels.Match, 0, len(memberIds))
	for i, angel := range memberIds {
		mortal := memberIds[(i+1)%len(memberIds)]
		matches = append(matches, &appModels.Match{
			SessionId: session.ID,
			AngelId:   angel,
			MortalId:  mortal,
		})
	}

	if _, createMatchesErr := db.CreateMatches(ctx, matches, session.ID); createMatchesErr != nil {
		// clean up
		_, _ = db.DeleteSessionByID(ctx, session.ID)
		// (optional) unset circle.currentSession if you set it earlier
		_, _ = db.UnsetCircleCurrentSession(ctx, circle.ID, session.ID)

		fmt.Printf("failed to create matches for circle %s: %v\n", circleName, createMatchesErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if setSessionErr := db.SetCircleCurrentSession(ctx, circle.ID, session.ID); setSessionErr != nil {
		fmt.Printf("failed to set session for circle %s: %v\n", circleName, setSessionErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Your session has been started. Enjoy yourself! 🎉",
	})

}

func RevealMortalCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Reveal mortal")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		if errors.Is(getCircleErr, mongo.ErrNoDocuments) {
			// Custom user-friendly message
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("No circle found with the name '%s'.", circleName),
			})

			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if circle.CurrentSession == nil {
		utils.SendCustomErrorMessage(ctx, b, chatID, "There is no active session for this circle!")
		return
	}

	session, getSessErr := db.GetSession(ctx, *circle.CurrentSession)
	if getSessErr != nil {
		if errors.Is(getSessErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch session:", getSessErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	match, getMatchErr := db.GetMortalMatch(ctx, session.ID, user.ID)

	if getMatchErr != nil {
		if errors.Is(getMatchErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch match:", getSessErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	mortal, getUserErr := db.GetUser(ctx, match.MortalId)
	if getUserErr != nil {
		if errors.Is(getUserErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch user:", getUserErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	fullName := mortal.FirstName
	if mortal.LastName != "" {
		fullName += " " + mortal.LastName
	}

	mortalInfo := fullName
	if mortal.UserHandle != "" {
		mortalInfo += fmt.Sprintf(" (@%s)", mortal.UserHandle)
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		// ReplyMarkup: fmt.Sprintf("Your mortal is: ||%s||", mortalInfo),
		ParseMode: models.ParseModeHTML,
		Text:      fmt.Sprintf(`Your mortal is: <span class="tg-spoiler">%s</span>`, mortalInfo),
	})
}

func RevealAngelCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Reveal angel")

	user, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		if errors.Is(getCircleErr, mongo.ErrNoDocuments) {
			// Custom user-friendly message
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("No circle found with the name '%s'.", circleName),
			})

			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if circle.CurrentSession == nil {
		utils.SendCustomErrorMessage(ctx, b, chatID, "There is no active session for this circle!")
		return
	}

	session, getSessErr := db.GetSession(ctx, *circle.CurrentSession)
	if getSessErr != nil {
		if errors.Is(getSessErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch session:", getSessErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if session.State != appModels.StateFinished {
		fmt.Printf("Session is not complete yet: %s\n", session.ID)
		utils.SendCustomErrorMessage(ctx, b, chatID, "The session is not yet over! You cant see your angel yet.")
		return
	}

	match, getMatchErr := db.GetAngelMatch(ctx, session.ID, user.ID)

	if getMatchErr != nil {
		if errors.Is(getMatchErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch match:", getSessErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	angel, getUserErr := db.GetUser(ctx, match.AngelId)
	if getUserErr != nil {
		if errors.Is(getUserErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch user:", getUserErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	fullName := angel.FirstName
	if angel.LastName != "" {
		fullName += " " + angel.LastName
	}

	angelInfo := fullName
	if angel.UserHandle != "" {
		angelInfo += fmt.Sprintf(" (@%s)", angel.UserHandle)
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		ParseMode: models.ParseModeHTML,
		Text:      fmt.Sprintf(`Your angel is: <span class="tg-spoiler">%s</span>`, angelInfo),
	})
}

func SendMessageToAngelCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Send angel message")

}

func SendMessageToMortalCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("Send mortal message")

}

func EndSessionCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, circleName string) {
	fmt.Println("End Session")

	_, chatID, extractErr := utils.ExtractUserAndChat(update)
	if extractErr != nil {
		fmt.Println("Error extracting user/chat:", extractErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	circle, getCircleErr := db.GetCircle(ctx, circleName)

	if getCircleErr != nil {
		if errors.Is(getCircleErr, mongo.ErrNoDocuments) {
			// Custom user-friendly message
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("No circle found with the name '%s'.", circleName),
			})

			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Printf("failed to get circle %s: %v\n", circleName, getCircleErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if circle.CurrentSession == nil {
		utils.SendCustomErrorMessage(ctx, b, chatID, "There is no active session for this circle!")
		return
	}

	session, getSessErr := db.GetSession(ctx, *circle.CurrentSession)
	if getSessErr != nil {
		if errors.Is(getSessErr, mongo.ErrNoDocuments) || session == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "You don't seem to have an active session.",
			})
			if menu, ok := ui.GetMenu(ui.MenuNameMain); ok {
				utils.SendMenu(ctx, b, chatID, menu)
			}
			return
		}
		fmt.Println("failed to fetch session:", getSessErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	if session.State == appModels.StateFinished {
		fmt.Println("Session already finished:", session.ID)
		utils.SendCustomErrorMessage(ctx, b, chatID, "Session has already been completed!")
		return
	}

	// Mark session as finished
	updated, finishErr := db.UpdateSessionToFinished(ctx, session.ID)
	if finishErr != nil {
		fmt.Println("failed to finish session:", finishErr)
		utils.SendErrorMessage(ctx, b, chatID)
		return
	}

	// Success message
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Your session has been ended. Thanks for playing! 🎉",
	})

	_ = updated
}
