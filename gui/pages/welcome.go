package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/kucli"
	"time"
)

type WelcomePage struct {
	MainWindow *MainWindow `wire:""`
	KuCli      kucli.KuCli `wire:""`
}

func NewWelcomePage() *WelcomePage {
	return &WelcomePage{}
}

func (p *WelcomePage) Build() fyne.CanvasObject {
	card := widget.NewCard("Bookmarks", "", p.newList())
	go func() {
		for range time.Tick(time.Second) {
			card.SetContent(p.newList())
			card.Refresh()
		}
	}()
	return card
}

func (p *WelcomePage) newList() *widget.List {
	collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
	data := binding.NewStringList()
	list := widget.NewListWithData(data, func() fyne.CanvasObject {
		return widget.NewLabel("")
	}, func(item binding.DataItem, object fyne.CanvasObject) {
		s, _ := item.(binding.String).Get()
		object.(*widget.Label).SetText(s)
	})
	list.OnSelected = func(id widget.ListItemID) {
		namespacePage := newNamespace(p.MainWindow, collections[id])
		p.MainWindow.AddTab(collections[id], func(ctx context.Context) fyne.CanvasObject {
			return namespacePage.Build(ctx)
		})
	}
	data.Set(collections)
	go func() {
		for range time.Tick(time.Second) {
			collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
			data.Set(collections)
		}
	}()
	return list
}
