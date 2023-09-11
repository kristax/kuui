package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"
	"strings"
)

type RemoveDupPage struct {
	in1  *widget.Entry
	in2  *widget.Entry
	main *fyne.Container
}

func (p *RemoveDupPage) Menu() string {
	return "Tool"
}

func (p *RemoveDupPage) Name() string {
	return "RemoveDup"
}

func NewRemoveDupPage() *RemoveDupPage {
	return &RemoveDupPage{
		in1:  widget.NewMultiLineEntry(),
		in2:  widget.NewMultiLineEntry(),
		main: nil,
	}
}

func (p *RemoveDupPage) Init() error {
	p.in1.SetPlaceHolder("input text per line")
	p.in2.SetPlaceHolder("remove duplicates")

	inSplit := container.NewHSplit(p.in1, p.in2)
	btnCompare := widget.NewButton("Remove Duplicates", func() {
		split := strings.Split(p.in1.Text, "\n")
		split = lo.Uniq(split)
		p.in2.SetText(strings.Join(split, "\n"))
	})
	p.main = container.NewBorder(nil, btnCompare, nil, nil, inSplit)
	return nil
}

func (p *RemoveDupPage) Build() fyne.CanvasObject {
	return p.main
}
