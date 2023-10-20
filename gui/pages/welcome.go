package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/channels"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/kucli"
	"github.com/kristax/kuui/streamer"
)

type WelcomePage struct {
	MainWindow *MainWindow     `wire:""`
	KuCli      kucli.KuCli     `wire:""`
	Streamer   streamer.Client `wire:""`

	collectionsData binding.StringList
}

func (p *WelcomePage) Channel() string {
	return channels.CollectionsUpdate
}

func (p *WelcomePage) OnCall(ctx context.Context, msg any) {
	collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
	p.collectionsData.Set(collections)
}

func NewWelcomePage() *WelcomePage {
	return &WelcomePage{
		collectionsData: binding.NewStringList(),
	}
}

func (p *WelcomePage) Init() error {
	p.Streamer.Register(p)
	return nil
}

func (p *WelcomePage) Build() fyne.CanvasObject {
	card := widget.NewCard("Bookmarks", "", p.newList())
	return card
}

func (p *WelcomePage) newList() *widget.List {
	collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
	list := widget.NewListWithData(p.collectionsData, func() fyne.CanvasObject {
		return widget.NewLabel("")
	}, func(item binding.DataItem, object fyne.CanvasObject) {
		s, _ := item.(binding.String).Get()
		object.(*widget.Label).SetText(s)
	})
	list.OnSelected = func(id widget.ListItemID) {
		value, _ := p.collectionsData.GetValue(id)
		namespacePage := newNamespace(p.MainWindow, value)
		p.MainWindow.AddTab(collections[id], func(ctx context.Context) fyne.CanvasObject {
			return namespacePage.Build(ctx)
		})
		list.UnselectAll()
	}
	p.collectionsData.Set(collections)
	return list
}
