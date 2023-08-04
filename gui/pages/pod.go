package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/kristax/kuui/kucli"
	"github.com/kristax/kuui/util/fas"
	v1 "k8s.io/api/core/v1"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type PodPage struct {
	cli      kucli.KuCli
	pod      *v1.Pod
	addTabFn func(name string, content func(ctx context.Context) fyne.CanvasObject)

	//txtLog         *widget.Entry
	logData binding.String
	txtLog  *widget.Entry

	frameLogDetail *fyne.Container

	list    *fyne.Container
	vScroll *container.Scroll

	tbiPause  *widget.ToolbarAction
	tbiDelete *widget.ToolbarAction
	tbiPrint  *widget.ToolbarAction
	tbiClear  *widget.ToolbarAction

	pause  bool
	line   int
	search string

	logCh   chan string
	temp    []string
	isPrint bool
}

func NewPodPage(cli kucli.KuCli, pod *v1.Pod, addTabFn func(name string, content func(ctx context.Context) fyne.CanvasObject)) *PodPage {
	logData := binding.NewString()
	txtLog := widget.NewMultiLineEntry()
	txtLog.Bind(logData)
	tbLogDetail := widget.NewToolbar(widget.NewToolbarAction(theme.ContentCutIcon(), func() {
		txt, _ := logData.Get()
		s, err := cutJson(txt)
		if err != nil {
			return
		}
		logData.Set(s)
	}))
	list := container.NewVBox()
	isPrint := fyne.CurrentApp().Preferences().Bool(preference.IsPrint)
	return &PodPage{
		cli:            cli,
		pod:            pod,
		addTabFn:       addTabFn,
		logData:        logData,
		txtLog:         txtLog,
		frameLogDetail: container.NewBorder(tbLogDetail, nil, nil, nil, txtLog),
		list:           list,
		vScroll:        container.NewScroll(list),
		tbiPause:       widget.NewToolbarAction(theme.MediaPauseIcon(), nil),
		tbiDelete:      widget.NewToolbarAction(theme.MediaStopIcon(), nil),
		tbiPrint:       widget.NewToolbarAction(fas.TernaryOp(isPrint, theme.RadioButtonCheckedIcon(), theme.RadioButtonIcon()), nil),
		tbiClear:       widget.NewToolbarAction(theme.DeleteIcon(), nil),
		pause:          false,
		line:           100,
		search:         "",
		logCh:          nil,
		temp:           nil,
		isPrint:        isPrint,
	}
}

func (p *PodPage) Build(ctx context.Context) fyne.CanvasObject {
	txtSearch := widget.NewEntry()
	toolbar := widget.NewToolbar(widget.NewToolbarSeparator())
	txtLine := widgets.NewNumericalEntry()
	searchBorder := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel("tail"), txtLine), txtSearch)
	titleBox := container.NewBorder(nil, nil, nil, toolbar, searchBorder)
	border := container.NewBorder(titleBox, nil, nil, nil, p.vScroll)

	{
		//pause
		p.tbiPause.OnActivated = func() {
			p.pause = !p.pause
			if p.pause {
				p.tbiPause.SetIcon(theme.MediaPlayIcon())
			} else {
				if len(p.temp) != 0 {
					for _, log := range p.temp {
						p.logCh <- log
					}
					p.temp = []string{}
				}
				p.tbiPause.SetIcon(theme.MediaPauseIcon())
			}
			toolbar.Refresh()
		}

		//delete
		p.tbiDelete.OnActivated = func() {
			err := p.cli.DeletePod(ctx, p.pod.GetNamespace(), p.pod.GetName())
			if err != nil {
				p.logCh <- fmt.Sprintf("delete pod failed: %v", err)
			}
		}

		//clear
		p.tbiClear.OnActivated = func() {
			p.list.RemoveAll()
			p.temp = []string{}
		}

		//refresh
		p.tbiPrint.OnActivated = func() {
			p.isPrint = !p.isPrint
			fyne.CurrentApp().Preferences().SetBool(preference.IsPrint, p.isPrint)
			if p.isPrint {
				p.tbiPrint.SetIcon(theme.RadioButtonCheckedIcon())
			} else {
				p.tbiPrint.SetIcon(theme.RadioButtonIcon())
			}
			toolbar.Refresh()
		}

		toolbar.Append(p.tbiPause)
		toolbar.Append(p.tbiDelete)
		toolbar.Append(p.tbiClear)
		toolbar.Append(p.tbiPrint)
	}

	p.txtLog.Wrapping = fyne.TextWrapWord
	p.txtLog.OnChanged = func(s string) {
		f, err := formatLog(s)
		if err != nil {
			f = s
		}
		p.logData.Set(f)
	}

	txtLine.SetText(strconv.Itoa(p.line))
	txtLine.OnSubmitted = func(s string) {
		d, _ := strconv.Atoi(s)
		p.line = d
		p.reloadLog(ctx)
	}
	txtSearch.OnSubmitted = func(s string) {
		p.search = s
		p.reloadLog(ctx)
	}

	go p.run(ctx)
	return border
}

func (p *PodPage) run(ctx context.Context) {
	var err error
	p.logCh, err = p.cli.TailfLogs(ctx, p.pod.GetNamespace(), p.pod.GetName(), int64(p.line))
	if err != nil {
		fyne.LogError("tail logs failed", err)
		p.AddItem(fmt.Sprintf("tail logs failed: %v", err))
		return
	}
	for log := range p.logCh {
		if p.pause {
			p.temp = append(p.temp, log)
			continue
		}
		if p.search != "" && !strings.Contains(strings.ToLower(log), strings.ToLower(p.search)) {
			continue
		}
		if len(p.list.Objects) > p.line {
			p.list.Remove(p.list.Objects[0])
		}
		p.AddItem(log)
	}
}

func (p *PodPage) reloadLog(ctx context.Context) {
	//temporary close pause mode
	if p.pause {
		p.tbiPause.OnActivated()
	}
	//query all logs
	logs, err := p.cli.TailLogs(ctx, p.pod.GetNamespace(), p.pod.GetName(), int64(p.line))
	if err != nil {
		p.logCh <- fmt.Sprintf("Tail Logs failed: %v", err)
		return
	}
	//clear screen
	p.list.RemoveAll()
	for _, log := range logs {
		p.logCh <- log
	}
}

var compile = regexp.MustCompile(`\[\d+(;\d+)*m`)

func (p *PodPage) AddItem(txt string) {
	if p.isPrint {
		os.Stdout.WriteString(txt)
	}
	if compile.MatchString(txt) {
		txt = compile.ReplaceAllString(txt, "**")
	}
	content := widgets.NewTappableLabel(txt)
	content.TextStyle = fyne.TextStyle{
		Monospace: true,
		Symbol:    true,
	}
	content.Wrapping = fyne.TextWrapWord
	content.OnTapped = p.contentTapped(txt)
	p.list.Add(content)
	p.vScroll.ScrollToBottom()
}

func (p *PodPage) contentTapped(txt string) func() {
	return func() {
		p.logData.Set(txt)
		p.addTabFn("L:"+p.pod.GetName(), func(ctx context.Context) fyne.CanvasObject {
			return p.frameLogDetail
		})
	}
}

func formatLog(txt string) (string, error) {
	if strings.HasPrefix(txt, "{") && strings.HasSuffix(txt, "}") {
		var m = make(map[string]interface{})
		err := json.Unmarshal([]byte(txt), &m)
		if err != nil {
			return txt, err
		}
		bytes, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return txt, err
		}
		return string(bytes), nil
	}
	return txt, nil
}

func cutJson(s string) (string, error) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 {
		return s, nil
	}
	s = s[start+1 : end]
	start = strings.Index(s, "{")
	end = strings.LastIndex(s, "}")
	if start == -1 || end == -1 {
		return s, nil
	}
	s = s[start : end+1]
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, ``)
	return s, nil
}
