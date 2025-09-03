package models

import (
	"fmt"
	"grandfather/internal/commands.go"
	"grandfather/internal/ui"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Circle struct {
	ID             bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name           string         `bson:"name" json:"name"`
	OwnerId        int64          `bson:"ownerId" json:"ownerId"`
	Members        []int64        `bson:"members" json:"members"`
	CurrentSession *bson.ObjectID `bson:"currentSession,omitempty" json:"currentSession,omitempty"`
}

func (circle Circle) ToMenu(userID int64) ui.Menu {
	circleName := circle.Name

	// TODO: Get all user names and put into member list

	title := fmt.Sprintf("👥 Circle: %s\nWhat would you like to do?", circleName)

	circleMenu := ui.Menu{
		Title:   title,
		Buttons: [][]ui.MenuButton{},
	}

	isOwner := circle.OwnerId == userID
	if isOwner {
		circleMenu.PrependButtonRow("End session", string(commands.EndSessionCommand)+"@"+circleName)
		circleMenu.PrependButtonRow("Start session", string(commands.StartNewSessionCommand)+"@"+circleName)
		circleMenu.PrependButtonRow("Remove member", string(commands.RemoveUserCommand)+"@"+circleName)
	}

	circleMenu.PrependButtonRow("Member list", string(commands.GetMemberListCommand)+"@"+circleName)
	circleMenu.AddButtonRow("Reveal mortal", string(commands.RevealMortalCommand)+"@"+circleName)
	circleMenu.AddButtonRow("Reveal angel", string(commands.RevealAngelCommand)+"@"+circleName)
	circleMenu.AddButtonRow("Send message to mortal", string(commands.SendMessageCommandToMortal)+"@"+circleName)
	circleMenu.AddButtonRow("Send message to angel", string(commands.SendMessageCommandToAngel)+"@"+circleName)
	circleMenu.AddButtonRow("Back", string(commands.ListCirclesCommand))

	return circleMenu
}
