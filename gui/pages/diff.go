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
	result   *widget.TextGrid
	in1Entry *widget.Entry
	in2Entry *widget.Entry
}

func NewDiffPage() *DiffPage {
	result := widget.NewTextGrid()
	result.ShowLineNumbers = true
	result.ShowWhitespace = true
	return &DiffPage{
		result:   result,
		in1Entry: widget.NewMultiLineEntry(),
		in2Entry: widget.NewMultiLineEntry(),
	}
}

func (p *DiffPage) Build() fyne.CanvasObject {
	inSplit := container.NewVSplit(p.in1Entry, p.in2Entry)
	scroll := container.NewScroll(p.result)
	resultSplit := container.NewHSplit(inSplit, scroll)

	p.in1Entry.OnChanged = p.onInputChange
	p.in2Entry.OnChanged = p.onInputChange
	return resultSplit
}

func (p *DiffPage) onInputChange(_ string) {
	in1 := p.in1Entry.Text
	in2 := p.in2Entry.Text
	if in1 == "" || in2 == "" {
		return
	}
	diff := cmp.Diff(in1, in2)
	p.makeResult(diff)
}

func (p *DiffPage) makeResult(diff string) {
	//diff = formatDiff(diff)
	if diff == "" {
		diff = "no differences"
	}
	p.result.SetText(diff)
	defer p.result.Refresh()
	split := strings.Split(diff, "\n")
	for i, s := range split {
		if strings.HasPrefix(s, "+") {
			p.result.SetRowStyle(i, &widget.CustomTextGridStyle{
				BGColor: color.RGBA{
					R: 0,
					G: 255,
					B: 0,
					A: 50,
				},
			})
		} else if strings.HasPrefix(s, "-") {
			p.result.SetRowStyle(i, &widget.CustomTextGridStyle{
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
