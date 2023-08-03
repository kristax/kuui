package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	var (
		pause = false
		//history []string
		temp    []string
		search  string
		isPrint = fyne.CurrentApp().Preferences().Bool(preference.IsPrint)
	)

	p.list = container.NewVBox()
	p.vScroll = container.NewScroll(p.list)
	txtSearch := widget.NewEntry()

	toolbar := widget.NewToolbar(widget.NewToolbarSeparator())
	tbiPause := widget.NewToolbarAction(theme.MediaPauseIcon(), nil)
	tbiDelete := widget.NewToolbarAction(theme.MediaStopIcon(), nil)
	tbiPrint := widget.NewToolbarAction(fas.TernaryOp(isPrint, theme.RadioButtonCheckedIcon(), theme.RadioButtonIcon()), nil)
	tbiClear := widget.NewToolbarAction(theme.DeleteIcon(), nil)

	txtLine := widgets.NewNumericalEntry()
	searchBorder := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel("tail"), txtLine), txtSearch)
	titleBox := container.NewBorder(nil, nil, nil, toolbar, searchBorder)

	p.txtLog = widget.NewMultiLineEntry()
	tbLogDetail := widget.NewToolbar(widget.NewToolbarAction(theme.ContentCutIcon(), func() {
		s, err := cutJson(p.txtLog.Text)
		if err != nil {
			return
		}
		p.txtLog.SetText(s)
	}))
	p.frameLogDetail = container.NewBorder(tbLogDetail, nil, nil, nil, p.txtLog)

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

		//delete
		tbiDelete.OnActivated = func() {
			err := p.cli.DeletePod(ctx, p.pod.GetNamespace(), p.pod.GetName())
			if err != nil {
				logCh <- fmt.Sprintf("delete pod failed: %v", err)
			}
		}

		//clear
		tbiClear.OnActivated = func() {
			p.list.RemoveAll()
			temp = []string{}
		}

		//refresh
		tbiPrint.OnActivated = func() {
			isPrint = !isPrint
			fyne.CurrentApp().Preferences().SetBool(preference.IsPrint, isPrint)
			if isPrint {
				tbiPrint.SetIcon(theme.RadioButtonCheckedIcon())
			} else {
				tbiPrint.SetIcon(theme.RadioButtonIcon())
			}
			toolbar.Refresh()
		}

		toolbar.Append(tbiPause)
		toolbar.Append(tbiDelete)
		toolbar.Append(tbiClear)
		toolbar.Append(tbiPrint)
	}

	p.txtLog.TextStyle = fyne.TextStyle{
		Monospace: true,
		Symbol:    true,
	}
	p.txtLog.Wrapping = fyne.TextWrapWord
	p.txtLog.OnChanged = func(s string) {
		f, err := formatLog(s)
		if err != nil {
			f = s
		}
		p.txtLog.SetText(f)
	}

	txtSearch.OnSubmitted = func(s string) {
		search = s
		//temporary close pause mode
		if pause {
			tbiPause.OnActivated()
			defer tbiPause.OnActivated()
		}
		//query all logs
		logs, err := p.cli.TailLogs(ctx, p.pod.GetNamespace(), p.pod.GetName(), int64(line))
		if err != nil {
			logCh <- fmt.Sprintf("Tail Logs failed: %v", err)
			return
		}
		//clear screen
		p.list.RemoveAll()
		for _, log := range logs {
			logCh <- log
		}
	}

	go func() {
		for log := range logCh {
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
			if isPrint {
				os.Stdout.WriteString(log)
				//fmt.Println(log)
			}
		}
	}()

	return border
}

var compile = regexp.MustCompile(`\[\d+(;\d+)*m`)

func (p *PodPage) AddItem(txt string) {
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
		p.txtLog.SetText(txt)
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
