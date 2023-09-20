package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/themes"
	"github.com/kristax/kuui/util/fas"
	"image/color"
)

type ThemePage struct {
}

func (p *ThemePage) Menu() string {
	return "Preference"
}

func (p *ThemePage) Name() string {
	return "Theme"
}

func NewThemePage() *ThemePage {
	return &ThemePage{}
}

func (p *ThemePage) Init() error {
	app := fyne.CurrentApp()
	isDark := app.Preferences().BoolWithFallback(preference.ThemeDark, app.Settings().ThemeVariant() == theme.VariantDark)
	app.Settings().SetTheme(fas.TernaryOp(isDark, themes.DarkTheme(), themes.LightTheme()))
	return nil
}

func (p *ThemePage) Build() fyne.CanvasObject {
	app := fyne.CurrentApp()
	var changeTheme = func(isDark bool) {
		app.Preferences().SetBool(preference.ThemeDark, isDark)
		app.Settings().SetTheme(fas.TernaryOp(isDark, themes.DarkTheme(), themes.LightTheme()))
	}
	themeDark := container.NewBorder(nil, widget.NewButton("Set", func() { changeTheme(true) }), nil, nil, canvas.NewRectangle(color.Black))
	themeLight := container.NewBorder(nil, widget.NewButton("Set", func() { changeTheme(false) }), nil, nil, canvas.NewRectangle(color.White))
	return container.NewAdaptiveGrid(2,
		widget.NewCard("Default Theme", "Dark", themeDark),
		widget.NewCard("Default Theme", "Light", themeLight),
	)
}
