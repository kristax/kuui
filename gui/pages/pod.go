package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/kristax/kuui/kucli"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"regexp"
	"strconv"
	"strings"
)

type PodPage struct {
	cli      kucli.KuCli
	pod      *v1.Pod
	addTabFn func(name string, content func(ctx context.Context) fyne.CanvasObject)

	txtLog         *widget.Entry
	frameLogDetail *fyne.Container
	list           *fyne.Container
	vScroll        *container.Scroll
}

func NewPodPage(cli kucli.KuCli, pod *v1.Pod, addTabFn func(name string, content func(ctx context.Context) fyne.CanvasObject)) *PodPage {
	return &PodPage{
		cli:      cli,
		pod:      pod,
		addTabFn: addTabFn,
	}
}

func (p *PodPage) Build(ctx context.Context) fyne.CanvasObject {
	p.list = container.NewVBox()
	p.vScroll = container.NewScroll(p.list)
	txtSearch := widget.NewEntry()

	toolbar := widget.NewToolbar(widget.NewToolbarSeparator())
	tbiPause := widget.NewToolbarAction(theme.MediaPauseIcon(), nil)
	tbiDelete := widget.NewToolbarAction(theme.MediaStopIcon(), nil)
	tbiRefresh := widget.NewToolbarAction(theme.ViewRefreshIcon(), nil)
	tbiClear := widget.NewToolbarAction(theme.DeleteIcon(), nil)

	txtLine := widgets.NewNumericalEntry()
	searchBorder := container.NewBorder(nil, nil, nil, txtLine, txtSearch)
	titleBox := container.NewBorder(nil, nil, nil, toolbar, searchBorder)

	p.txtLog = widget.NewMultiLineEntry()
	p.frameLogDetail = container.NewBorder(nil, nil, nil, nil, p.txtLog)

	border := container.NewBorder(titleBox, nil, nil, nil, p.vScroll)

	var line = 100
	txtLine.SetText(strconv.Itoa(line))
	txtLine.OnChanged = func(s string) {
		if s == "" {
			txtLine.SetText(strconv.Itoa(line))
			return
		}
		d, _ := strconv.Atoi(s)
		if d == 0 {
			txtLine.SetText(strconv.Itoa(1))
			return
		}
		line = d
		length := len(p.list.Objects)
		if length >= line {
			sub := length - line
			for i := 0; i <= sub; i++ {
				p.list.Remove(p.list.Objects[0])
			}
			p.list.Refresh()
			p.vScroll.ScrollToBottom()
		}
	}

	logCh, err := p.cli.TailfLogs(ctx, p.pod.GetNamespace(), p.pod.GetName(), int64(line))
	if err != nil {
		fyne.LogError("tail logs failed", err)
		p.AddItem(fmt.Sprintf("tail logs failed: %v", err))
		return border
	}

	var (
		pause = false
		//history []string
		temp   []string
		search string
	)
	{
		//pause
		tbiPause.OnActivated = func() {
			pause = !pause
			if pause {
				tbiPause.SetIcon(theme.MediaPlayIcon())
			} else {
				if len(temp) != 0 {
					for _, log := range temp {
						logCh <- log
					}
					temp = []string{}
				}
				tbiPause.SetIcon(theme.MediaPauseIcon())
			}
			toolbar.Refresh()
		}
		toolbar.Append(tbiPause)

		//refresh
		tbiRefresh.OnActivated = func() {
			p.Build(ctx)
		}
		toolbar.Append(tbiRefresh)

		//delete
		tbiDelete.OnActivated = func() {
			err := p.cli.DeletePod(ctx, p.pod.GetNamespace(), p.pod.GetName())
			if err != nil {
				p.AddItem(fmt.Sprintf("delete pod failed: %v", err))
			}
		}
		toolbar.Append(tbiDelete)

		tbiClear.OnActivated = func() {
			p.list.RemoveAll()
			//history = []string{}
			temp = []string{}
		}
		toolbar.Append(tbiClear)

	}

	p.txtLog.TextStyle = fyne.TextStyle{
		Monospace: true,
		Symbol:    true,
	}
	p.txtLog.Wrapping = fyne.TextWrapWord

	txtSearch.OnSubmitted = func(s string) {
		search = s
		logs, err := p.cli.TailLogs(ctx, p.pod.GetNamespace(), p.pod.GetName(), int64(line))
		if err != nil {
			p.AddItem(fmt.Sprintf("Tail Logs failed: %v", err))
			return
		}
		filtered := lo.Filter(logs, func(item string, _ int) bool {
			return strings.Contains(strings.ToLower(item), strings.ToLower(s))
		})
		p.list.RemoveAll()
		for _, log := range filtered {
			p.AddItem(log)
		}
	}

	go func() {
		for log := range logCh {
			if log == "" || log == "\n" {
				continue
			}
			if pause {
				temp = append(temp, log)
				continue
			}
			if search != "" && !strings.Contains(strings.ToLower(log), strings.ToLower(search)) {
				continue
			}
			if len(p.list.Objects) >= line {
				p.list.Remove(p.list.Objects[0])
			}
			p.AddItem(log)
		}
	}()

	return border
}

var compile = regexp.MustCompile(`\[\d+(;\d+)*m`)

func (p *PodPage) AddItem(txt string) {
	if compile.MatchString(txt) {
		txt = compile.ReplaceAllString(txt, "")
	}
	content := widgets.NewTappableLabel(txt)
	content.TextStyle = fyne.TextStyle{
		Monospace: true,
		Symbol:    true,
	}
	content.Wrapping = fyne.TextWrapWord
	content.OnTapped = func() {
		var formatted = formatLog(txt)
		p.txtLog.SetText(formatted)

		p.addTabFn("L:"+p.pod.GetName(), func(ctx context.Context) fyne.CanvasObject {
			p.txtLog.SetText(txt)
			return p.frameLogDetail
		})
	}
	p.list.Add(content)
	p.vScroll.ScrollToBottom()
}

func (p *PodPage) addItem(object fyne.CanvasObject) {
	p.list.Add(object)
	p.vScroll.ScrollToBottom()
}

func formatLog(txt string) string {
	var formatted = txt
	if strings.HasPrefix(formatted, "{") {
		var m = make(map[string]interface{})
		err := json.Unmarshal([]byte(formatted), &m)
		if err != nil {
			fyne.LogError("", err)
			return err.Error()
		}
		bytes, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			fyne.LogError("", err)
			return err.Error()
		}
		formatted = string(bytes)
	}
	return formatted
}
