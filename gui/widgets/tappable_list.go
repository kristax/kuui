package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type TappableList struct {
	container.Scroll
}

func NewTappableList() *TappableList {
	t := &TappableList{}
	t.ExtendBaseWidget(t)
	t.Content = container.NewVBox()
	return t
}

func (t *TappableList) Get(i int) fyne.CanvasObject {
	return t.Content.(*fyne.Container).Objects[i]
}

func (t *TappableList) Add(i fyne.CanvasObject) {
	t.Content.(*fyne.Container).Add(i)
	t.Content.(*fyne.Container).Refresh()
}

func (t *TappableList) Remove(i fyne.CanvasObject) {
	t.Content.(*fyne.Container).Remove(i)
	t.Content.(*fyne.Container).Refresh()
}

func (t *TappableList) RemoveIndex(i int) {
	t.Remove(t.Get(i))
}

func (t *TappableList) RemoveAll() {
	t.Content.(*fyne.Container).RemoveAll()
	t.Content.(*fyne.Container).Refresh()
}

func (t *TappableList) Length() int {
	return len(t.Content.(*fyne.Container).Objects)
}
