package gui

import (
	"fyne.io/fyne/v2"
	"github.com/kristax/kuui/gui/pages"
)

type ui struct {
	App        fyne.App          `wire:""`
	MainWindow *pages.MainWindow `wire:""`
}

func NewUI() any {
	return &ui{}
}

func (u *ui) Order() int {
	return 0
}

func (u *ui) Run() error {
	return u.MainWindow.ShowAndRun()
}

func (u *ui) Init() error {
	//u.makeTray()
	//u.logLifecycle()
	return nil
}

//func (u *ui) makeTray() {
//	desk, ok := u.app.(desktop.App)
//	if !ok {
//		return
//	}
//	h := fyne.NewMenuItem("Hello", func() {})
//	h.Icon = theme.HomeIcon()
//	menu := fyne.NewMenu("Hello World", h)
//	h.Action = func() {
//		log.Println("System tray menu tapped")
//		h.Label = "Welcome"
//		menu.Refresh()
//	}
//	desk.SetSystemTrayMenu(menu)
//}
