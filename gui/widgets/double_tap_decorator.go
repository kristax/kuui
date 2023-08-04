package widgets

import (
	"fyne.io/fyne/v2"
)

type DoubleTapDecorator[T fyne.Widget] struct {
	Object         T
	OnDoubleTapped func(event *fyne.PointEvent)
}

func (d *DoubleTapDecorator[T]) DoubleTapped(event *fyne.PointEvent) {
	if d.OnDoubleTapped != nil {
		d.OnDoubleTapped(event)
	}
}

//func (d *DoubleTapDecorator[T]) Object() T {
//	return d.object
//}

func NewDoubleTapObject[T fyne.Widget](object T, onDoubleTapped func(event *fyne.PointEvent)) *DoubleTapDecorator[T] {
	return &DoubleTapDecorator[T]{
		Object:         object,
		OnDoubleTapped: onDoubleTapped,
	}
}
