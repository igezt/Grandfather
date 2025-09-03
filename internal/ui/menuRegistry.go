package ui

var Menus = map[string]Menu{}

const (
	MenuNameMain    = "main"
	MenuNameCircles = "circles"
)

func RegisterMenu(name string, menu Menu) {
	Menus[name] = menu
}

func GetMenu(name string) (Menu, bool) {
	m, ok := Menus[name]
	return m, ok
}

func RegisterMenus() {
	RegisterMenu(MenuNameMain, Menu{
		Title: "Welcome! Choose an option:",
		Buttons: [][]MenuButton{
			{
				{Text: "Start new circle", Command: "startNewCircle"},
				{Text: "Join circle", Command: "joinCircle"},
			},
			{
				{Text: "My circles", Command: "listCircles"},
			},
		},
	})

	RegisterMenu(MenuNameCircles, Menu{
		Title: "Your Circles:",
		Buttons: [][]MenuButton{
			{{Text: "Back", Command: "main"}},
		},
	})

}
