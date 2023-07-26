package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TappableLabel struct {
	widget.Label
	OnTapped func()
}

func NewTappableLabel(text string) *TappableLabel {
	t := &TappableLabel{}
	t.ExtendBaseWidget(t)
	t.SetText(text)
	return t
}

func (t *TappableLabel) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}
