package models

type UserState string

const (
	StateNone                       UserState = ""
	StateWaitingCircleName          UserState = "waiting_circle_name"
	StateWaitingJoinCircleName      UserState = "waiting_join_circle_name"
	StateWaitingSendMessageToAngel  UserState = "waiting_send_message_to_angel"
	StateWaitingSendMessageToMortal UserState = "waiting_send_message_to_mortal"
)

type User struct {
	ID          int64     `bson:"_id" json:"id"`         // Telegram user ID as the primary key
	ChatID      int64     `bson:"chat_id" json:"chatId"` // Chat ID (can differ from user ID, esp. groups)
	FirstName   string    `bson:"first_name" json:"firstName"`
	LastName    string    `bson:"last_name" json:"lastName"`
	UserHandle  string    `bson:"user_handle" json:"userHandle"`
	State       UserState `bson:"state" json:"state"`
	StateCircle string    `bson:"stateCircle,omitempty" json:"stateCircle,omitempty"`
}
