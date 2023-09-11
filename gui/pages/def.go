package pages

import "fyne.io/fyne/v2"

type MenuItem interface {
	Menu() string
	Name() string
	Build() fyne.CanvasObject
}
