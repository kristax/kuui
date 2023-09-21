package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type HoverTable struct {
	widget.Table
	OnMouseDown  func(event *desktop.MouseEvent)
	OnMouseUp    func(event *desktop.MouseEvent)
	OnMouseIn    func(event *desktop.MouseEvent)
	OnMouseMoved func(event *desktop.MouseEvent)
	OnMouseOut   func()
}

func NewHoverTable(length func() (int, int), create func() fyne.CanvasObject, update func(widget.TableCellID, fyne.CanvasObject)) *HoverTable {
	entry := &HoverTable{
		Table: *widget.NewTable(length, create, update),
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (h *HoverTable) MouseDown(event *desktop.MouseEvent) {
	h.Table.MouseDown(event)
	if h.OnMouseDown != nil {
		h.OnMouseDown(event)
	}
}

func (h *HoverTable) MouseUp(event *desktop.MouseEvent) {
	h.Table.MouseUp(event)
	if h.OnMouseUp != nil {
		h.OnMouseUp(event)
	}
}

func (h *HoverTable) MouseIn(event *desktop.MouseEvent) {
	h.Table.MouseIn(event)
	if h.OnMouseIn != nil {
		h.OnMouseIn(event)
	}
}

func (h *HoverTable) MouseMoved(event *desktop.MouseEvent) {
	h.Table.MouseMoved(event)
	if h.OnMouseMoved != nil {
		h.OnMouseMoved(event)
	}
}

func (h *HoverTable) MouseOut() {
	h.Table.MouseOut()
	if h.OnMouseOut != nil {
		h.OnMouseOut()
	}
}
