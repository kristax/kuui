package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"github.com/samber/lo"
)

func (u *MainWindow) makeMenu() {
	itemsGroup := lo.GroupBy[MenuItem, string](u.MenuItems, func(item MenuItem) string {
		return item.Menu()
	})

	var menus []*fyne.Menu
	for menu, items := range itemsGroup {
		menuItems := lo.Map[MenuItem, *fyne.MenuItem](items, func(item MenuItem, _ int) *fyne.MenuItem {
			return fyne.NewMenuItem(item.Name(), func() {
				u.AddTab(item.Name(), func(ctx context.Context) fyne.CanvasObject {
					return item.Build()
				})
			})
		})
		menus = append(menus, fyne.NewMenu(menu, menuItems...))
	}
	main := fyne.NewMainMenu(menus...)
	u.mainWindow.SetMainMenu(main)
}

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	switch sh := s.(type) {
	case *fyne.ShortcutCopy:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutCut:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutPaste:
		sh.Clipboard = w.Clipboard()
	}
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}
