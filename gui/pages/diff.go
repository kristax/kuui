package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/google/go-cmp/cmp"
	"image/color"
	"strings"
)

type DiffPage struct {
	in1  *widget.Entry
	in2  *widget.Entry
	rst1 *widget.TextGrid
	rst2 *widget.TextGrid
	main *fyne.Container
}

func NewDiffPage() *DiffPage {
	return &DiffPage{
		in1:  widget.NewMultiLineEntry(),
		in2:  widget.NewMultiLineEntry(),
		rst1: widget.NewTextGrid(),
		rst2: widget.NewTextGrid(),
		main: nil,
	}
}

func (p *DiffPage) Init() error {
	p.in1.SetPlaceHolder("current")
	p.in2.SetPlaceHolder("new")

	inSplit := container.NewHSplit(p.in1, p.in2)

	p.rst1.ShowWhitespace = true
	p.rst2.ShowWhitespace = true
	rstSplit := container.NewHSplit(container.NewHScroll(p.rst1), container.NewHScroll(p.rst2))
	rstScroll := container.NewVScroll(rstSplit)

	var inputMode = true
	btnCompare := widget.NewButton("Compare", nil)
	p.main = container.NewBorder(nil, btnCompare, nil, nil, inSplit)

	var changed = false
	p.in1.OnChanged = func(s string) {
		changed = true
	}
	p.in2.OnChanged = p.in1.OnChanged

	btnCompare.OnTapped = func() {
		inputMode = !inputMode
		if inputMode {
			btnCompare.SetText("Compare")
			p.main.Objects[0] = inSplit
		} else {
			if changed {
				go p.cmpDiff()
				changed = false
			}
			btnCompare.SetText("Back")
			p.main.Objects[0] = rstScroll
		}
	}
	return nil
}

func (p *DiffPage) Build() fyne.CanvasObject {
	return p.main
}

func (p *DiffPage) cmpDiff() {
	diff := cmp.Diff(p.in1.Text, p.in2.Text)
	b1 := strings.Builder{}
	b2 := strings.Builder{}
	split := strings.Split(diff, "\n")
	for _, s := range split {
		if strings.HasPrefix(s, "+") {
			b2.WriteString(s)
			b2.WriteString("\n")
		} else if strings.HasPrefix(s, "-") {
			b1.WriteString(s)
			b1.WriteString("\n")
		} else {
			b1.WriteString(s)
			b1.WriteString("\n")
			b2.WriteString(s)
			b2.WriteString("\n")
		}
	}
	p.makeResult(b1.String(), p.rst1)
	p.makeResult(b2.String(), p.rst2)
}

func (p *DiffPage) makeResult(diff string, result *widget.TextGrid) {
	//diff = formatDiff(diff)
	if diff == "" {
		diff = "no differences"
	}
	result.SetText(diff)
	defer result.Refresh()
	split := strings.Split(diff, "\n")
	for i, s := range split {
		if strings.HasPrefix(s, "+") {
			result.SetRowStyle(i, &widget.CustomTextGridStyle{
				BGColor: color.RGBA{
					R: 0,
					G: 255,
					B: 0,
					A: 50,
				},
			})
		} else if strings.HasPrefix(s, "-") {
			result.SetRowStyle(i, &widget.CustomTextGridStyle{
				BGColor: color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 50,
				},
			})
		}
	}
}

func formatDiff(s string) string {
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start == -1 || end == -1 {
		return s
	}
	s = s[start+1 : end]
	s = strings.Trim(s, "\n")
	s = strings.TrimSpace(s)
	return s
}
