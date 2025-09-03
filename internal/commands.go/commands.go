package commands

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Command string

const (
	MainMenuCommand            Command = "main"
	StartNewCircleCommand      Command = "startNewCircle"
	JoinCircleCommand          Command = "joinCircle"
	ListCirclesCommand         Command = "listCircles"
	GetCircleCommand           Command = "getCircle"
	RemoveUserCommand          Command = "removeUserCommand"
	RemoveSpecificUserCommand  Command = "removeSpecificUserCommand"
	StartNewSessionCommand     Command = "startNewSessionCommand"
	RevealMortalCommand        Command = "revealMortalCommand"
	RevealAngelCommand         Command = "revealAngelCommand"
	SendMessageCommandToMortal Command = "sendMessageCommandToMortal"
	SendMessageCommandToAngel  Command = "sendMessageCommandToAngel"
	EndSessionCommand          Command = "endSessionCommand"
	GetMemberListCommand       Command = "getMemberListCommand"
)

type CommandHandler func(ctx context.Context, b *bot.Bot, update *models.Update)

var Router = map[Command]CommandHandler{}
