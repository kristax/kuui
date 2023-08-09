package pages

import (
	"context"
	"fyne.io/fyne/v2"
)

func (u *MainWindow) makeMenu() {
	mnTool := fyne.NewMenu("Tool", fyne.NewMenuItem("diff", func() {
		u.AddTab("Diff", func(ctx context.Context) fyne.CanvasObject {
			return u.DiffPage.Build()
		})
	}))
	main := fyne.NewMainMenu(mnTool)
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
