package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogPage struct {
	txtLog       *widget.Entry
	detailWindow fyne.Window
}

func NewLogPage() *LogPage {
	txtLog := widget.NewMultiLineEntry()
	txtLog.Wrapping = fyne.TextWrapWord
	txtLog.TextStyle = fyne.TextStyle{
		Monospace: true,
		Symbol:    true,
	}
	detailWindow := fyne.CurrentApp().NewWindow("Log Detail")
	detailWindow.SetContent(container.NewBorder(widget.NewEntry(), nil, nil, nil, txtLog))
	detailWindow.Resize(fyne.NewSize(0.1, 0.1))
	//detailWindow.CenterOnScreen()
	detailWindow.SetCloseIntercept(func() {
		detailWindow.Hide()
	})
	return &LogPage{
		txtLog:       txtLog,
		detailWindow: detailWindow,
	}
}

func (p *LogPage) Show() {
	p.detailWindow.Show()
}

func (p *LogPage) Hide() {
	p.detailWindow.Hide()
}

func (p *LogPage) SetText(text string) {
	p.txtLog.SetText(text)
}
