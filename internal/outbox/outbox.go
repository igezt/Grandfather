package outbox

import (
	"context"
	"fmt"
	"grandfather/internal/db"
	"grandfather/internal/models"
	"time"

	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
)

type Outbox struct {
	stop chan struct{}
	ctx  context.Context
	bot  *bot.Bot
}

func NewOutbox(ctx context.Context, b *bot.Bot) *Outbox {
	o := &Outbox{stop: make(chan struct{})}
	o.ctx = ctx
	o.bot = b
	go o.run(ctx)
	return o
}

func (o *Outbox) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done, stopping poll")
			return
		case <-o.stop:
			fmt.Println("Stop channel closed, stopping poll")
			return
		default:
			o.poll()
			time.Sleep(5 * time.Second)
		}
	}
}

func (o *Outbox) Stop() {
	close(o.stop)
}

func (o *Outbox) poll() []*models.Message {
	fmt.Println("Polling for messages")

	messages, getMessagesErr := db.GetUndeliveredMessages(o.ctx)

	if getMessagesErr != nil {
		fmt.Printf("There was an error getting undelivered messages: %s\n", getMessagesErr)
		return nil
	}

	userIds := []int64{}

	for _, message := range messages {
		userIds = append(userIds, message.RecepientId)
	}
	usersArr, getUsersErr := db.GetUsers(o.ctx, userIds)

	if getUsersErr != nil {
		fmt.Printf("There was an error getting the users for the receipient messages: %s\n", getUsersErr)
	}

	users := map[int64]*models.User{}

	for _, user := range usersArr {
		users[user.ID] = user
	}

	deliveredCount := 0

	for _, message := range messages {
		deliverMessageErr := o.deliverMessage(message, users[message.RecepientId])
		if deliverMessageErr != nil {
			fmt.Printf("Delivering message error: %s", deliverMessageErr)
		} else {
			deliveredCount++
		}
	}

	return []*models.Message{}
}

func (o Outbox) deliverMessage(message *models.Message, user *models.User) error {
	// 1. Construct the Telegram message payload
	text := fmt.Sprintf(
		"✉️ You received a new message in circle *%s* from your %s:\n\n%s",
		message.CircleName,
		message.SenderRole,
		message.Message,
	)

	// 2. Send it to the recipient
	_, err := o.bot.SendMessage(o.ctx, &bot.SendMessageParams{
		ChatID:    user.ChatID,
		Text:      text,
		ParseMode: tgModels.ParseModeMarkdown, // or MarkdownV2/HTML
	})
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	// 3. Update state in DB to "delivered"
	updateMessageErr := db.UpdateMessageToDelivered(o.ctx, message.ID)
	if updateMessageErr != nil {
		return fmt.Errorf("update message state: %w", updateMessageErr)
	}

	return nil
}
