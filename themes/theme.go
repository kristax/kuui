package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct {
	Variant fyne.ThemeVariant
}

var _ fyne.Theme = (*MyTheme)(nil)

// return bundled font resource
func (*MyTheme) Font(s fyne.TextStyle) fyne.Resource {
	//if s.Monospace {
	//	return theme.DefaultTheme().Font(s)
	//}
	//if s.Bold {
	//	if s.Italic {
	//		return theme.DefaultTheme().Font(s)
	//	}
	//	return resourceMplus1cBoldTtf
	//}
	//if s.Italic {
	//	return theme.DefaultTheme().Font(s)
	//}
	return resourceJetBrainsMonoRegularTtf
}

func (m *MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if m.Variant == theme.VariantDark {
		return theme.DarkTheme().Color(n, v)
	}
	return theme.LightTheme().Color(n, v)
}

func (*MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (*MyTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}

func DarkTheme() fyne.Theme {
	return &MyTheme{Variant: theme.VariantDark}
}

func LightTheme() fyne.Theme {
	return &MyTheme{Variant: theme.VariantLight}
}
