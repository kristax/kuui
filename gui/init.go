package gui

import (
	"fyne.io/fyne/v2/app"
	"github.com/go-kid/ioc"
	"github.com/kristax/kuui/gui/pages"
)

func init() {
	ioc.Register(
		app.NewWithID("com.kristas.kuui"),
		NewUI(),
		pages.NewMainWindow(),
		pages.NewWelcomePage(),
		pages.NewNav(),
	)
}
