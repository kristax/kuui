package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"github.com/go-kid/ioc"
	"github.com/kristax/kuui/gui/channels"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/streamer"
	"github.com/samber/lo"
	"sort"
)

func init() {
	ioc.Register(new(collection))
}

type collection struct {
	Streamer streamer.Client `wire:""`
}

func (c *collection) GetCollections() []string {
	return fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
}

func (c *collection) Add(item string) {
	collections := c.GetCollections()
	collections = append(collections, item)
	c.update(collections)
}

func (c *collection) Remove(item string) {
	collections := c.GetCollections()
	collections = lo.Filter(collections, func(col string, _ int) bool {
		return col != item
	})
	c.update(collections)
}

func (c *collection) update(collections []string) {
	sort.Slice(collections, func(i, j int) bool {
		return collections[i] < collections[j]
	})
	fyne.CurrentApp().Preferences().SetStringList(preference.NSCollections, collections)
	c.Streamer.Send(context.Background(), channels.CollectionsUpdate, collections)
}
