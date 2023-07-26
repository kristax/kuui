package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type WelcomePage struct {
}

func (p *WelcomePage) Build() fyne.CanvasObject {
	return widget.NewLabel("Welcome")
}
