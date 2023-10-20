package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/channels"
	"github.com/kristax/kuui/kucli"
	"github.com/kristax/kuui/streamer"
)

type WelcomePage struct {
	MainWindow *MainWindow     `wire:""`
	KuCli      kucli.KuCli     `wire:""`
	Streamer   streamer.Client `wire:""`
	Collection *collection     `wire:""`

	collectionsData binding.StringList
}

func (p *WelcomePage) Channel() []string {
	return []string{channels.CollectionsUpdate}
}

func (p *WelcomePage) OnCall(ctx context.Context, channel string, msg any) {
	p.collectionsData.Set(p.Collection.GetCollections())
}

func NewWelcomePage() *WelcomePage {
	return &WelcomePage{
		collectionsData: binding.NewStringList(),
	}
}

func (p *WelcomePage) Build() fyne.CanvasObject {
	card := widget.NewCard("Bookmarks", "", p.newList())
	return card
}

func (p *WelcomePage) newList() *widget.List {
	list := widget.NewListWithData(p.collectionsData, func() fyne.CanvasObject {
		return container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("", theme.DeleteIcon(), nil), widget.NewLabel(""))
		//return widget.NewLabel("")
	}, func(item binding.DataItem, object fyne.CanvasObject) {
		s, _ := item.(binding.String).Get()
		border := object.(*fyne.Container)
		label := border.Objects[0].(*widget.Label)
		btn := border.Objects[1].(*widget.Button)
		label.SetText(s)
		btn.OnTapped = func() {
			p.Collection.Remove(s)
		}
	})
	list.OnSelected = func(id widget.ListItemID) {
		value, _ := p.collectionsData.GetValue(id)
		namespacePage := newNamespace(p.MainWindow, value)
		p.MainWindow.AddTab(value, func(ctx context.Context) fyne.CanvasObject {
			return namespacePage.Build(ctx)
		})
		list.UnselectAll()
	}
	p.collectionsData.Set(p.Collection.GetCollections())
	return list
}
