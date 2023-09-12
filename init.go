package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/go-kid/ioc"
	"github.com/kristax/kuui/gui"
	"github.com/kristax/kuui/gui/pages"
	"github.com/kristax/kuui/streamer"
)

func init() {
	ioc.Register(
		app.NewWithID("com.kristas.kuui"),
		gui.NewUI(),
		pages.NewMainWindow(),
		pages.NewWelcomePage(),
		pages.NewNav(),
		pages.NewDiffPage(),
		pages.NewRemoveDupPage(),
		streamer.NewStreamer(),
	)
}
