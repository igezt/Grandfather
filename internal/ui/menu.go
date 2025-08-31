package ui

import "github.com/go-telegram/bot/models"

type Menu struct {
	Title   string
	Buttons [][]MenuButton // Rows of buttons
}

type MenuButton struct {
	Text    string
	Command string // Callback data (command to trigger)
}

func (m *Menu) AddRow(buttons ...MenuButton) *Menu {
	m.Buttons = append(m.Buttons, buttons)
	return m
}

func (m *Menu) AddButtonRow(text, command string) *Menu {
	return m.AddRow(MenuButton{Text: text, Command: command})
}

func (m *Menu) PrependRow(buttons ...MenuButton) *Menu {
	m.Buttons = append([][]MenuButton{buttons}, m.Buttons...)
	return m
}

func (m *Menu) PrependButtonRow(text, command string) *Menu {
	return m.PrependRow(MenuButton{Text: text, Command: command})
}

func (m Menu) ToInlineKeyboard() *models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	for _, row := range m.Buttons {
		var btnRow []models.InlineKeyboardButton
		for _, btn := range row {
			btnRow = append(btnRow, models.InlineKeyboardButton{
				Text:         btn.Text,
				CallbackData: btn.Command,
			})
		}
		rows = append(rows, btnRow)
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}
