package pages

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/kristax/kuui/util/fas"
	v1 "k8s.io/api/core/v1"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LogListPage struct {
	mainWindow *MainWindow
	pods       []*v1.Pod

	logDetail *LogDetailPage

	list              *fyne.Container
	vScroll           *container.Scroll
	minScrollDuration time.Duration

	tbiPause  *widget.ToolbarAction
	tbiDelete *widget.ToolbarAction
	tbiPrint  *widget.ToolbarAction
	tbiClear  *widget.ToolbarAction

	pause  bool
	line   int
	search string

	logCh   chan string
	pauseWg sync.WaitGroup
	isPrint bool
}

func newLogListPage(mainWindow *MainWindow, pods []*v1.Pod) *LogListPage {
	list := container.NewVBox()
	isPrint := fyne.CurrentApp().Preferences().Bool(preference.IsPrint)
	return &LogListPage{
		mainWindow:        mainWindow,
		pods:              pods,
		list:              list,
		logDetail:         newLogDetailPage(mainWindow),
		vScroll:           container.NewScroll(list),
		minScrollDuration: time.Millisecond * 5,
		tbiPause:          widget.NewToolbarAction(theme.MediaPauseIcon(), nil),
		tbiDelete:         widget.NewToolbarAction(theme.MediaStopIcon(), nil),
		tbiPrint:          widget.NewToolbarAction(fas.TernaryOp(isPrint, theme.RadioButtonCheckedIcon(), theme.RadioButtonIcon()), nil),
		tbiClear:          widget.NewToolbarAction(theme.DeleteIcon(), nil),
		pause:             false,
		line:              100 * len(pods),
		search:            "",
		logCh:             make(chan string, 0),
		isPrint:           isPrint,
	}
}

func (p *LogListPage) Build(ctx context.Context) fyne.CanvasObject {
	txtSearch := widget.NewEntry()
	toolbar := widget.NewToolbar(widget.NewToolbarSeparator())
	txtLine := widgets.NewNumericalEntry()
	searchBorder := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel("line"), txtLine), txtSearch)
	titleBox := container.NewBorder(nil, nil, nil, toolbar, searchBorder)
	border := container.NewBorder(titleBox, nil, nil, nil, p.vScroll)

	{
		//pause
		p.tbiPause.OnActivated = func() {
			p.pause = !p.pause
			if p.pause {
				p.pauseWg.Add(1)
				p.tbiPause.SetIcon(theme.MediaPlayIcon())
			} else {
				p.pauseWg.Done()
				p.tbiPause.SetIcon(theme.MediaPauseIcon())
			}
			toolbar.Refresh()
		}

		//delete
		p.tbiDelete.OnActivated = func() {
			for _, pod := range p.pods {
				err := p.mainWindow.KuCli.DeletePod(ctx, pod.GetNamespace(), pod.GetName())
				if err != nil {
					p.logCh <- fmt.Sprintf("delete pod failed: %v", err)
				}
				p.logCh <- fmt.Sprintf("delete pod succeeded: %s/%s", pod.GetNamespace(), pod.GetName())
			}
		}

		//clear
		p.tbiClear.OnActivated = func() {
			p.list.RemoveAll()
			//p.temp = []string{}
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

	txtLine.SetText(strconv.Itoa(p.line))
	txtLine.OnSubmitted = func(s string) {
		p.reloadLog(ctx)
	}
	txtLine.OnChanged = func(s string) {
		d, _ := strconv.Atoi(s)
		p.line = d * len(p.pods)
	}
	txtSearch.OnSubmitted = func(s string) {
		p.search = s
		p.reloadLog(ctx)
	}

	go p.run(ctx)
	return border
}

func (p *LogListPage) run(ctx context.Context) {
	for _, pod := range p.pods {
		go func(pod *v1.Pod) {
			ch, err := p.mainWindow.KuCli.TailfLogs(ctx, pod.GetNamespace(), pod.GetName(), int64(p.line/len(p.pods)))
			if err != nil {
				fyne.LogError("tail logs failed", err)
				p.logCh <- fmt.Sprintf("tail logs failed: %v", err)
				return
			}
			for s := range ch {
				var els []string
				if len(p.pods) > 1 {
					els = append(els, pod.GetName())
				}
				els = append(els, s)
				p.logCh <- strings.Join(els, " | ")
			}
		}(pod)
	}

	for log := range p.logCh {
		p.pauseWg.Wait()
		if p.search != "" && !strings.Contains(strings.ToLower(log), strings.ToLower(p.search)) {
			continue
		}
		if len(p.list.Objects) > p.line {
			p.list.Remove(p.list.Objects[0])
		}
		p.addItem(log)
	}
}

func (p *LogListPage) reloadLog(ctx context.Context) {
	//temporary close pause mode
	if p.pause {
		p.tbiPause.OnActivated()
	}
	//clear screen
	p.list.RemoveAll()
	//query all logs
	for _, pod := range p.pods {
		logs, err := p.mainWindow.KuCli.TailLogs(ctx, pod.GetNamespace(), pod.GetName(), int64(p.line/len(p.pods)))
		if err != nil {
			p.logCh <- fmt.Sprintf("Tail Logs failed: %v", err)
			return
		}
		for _, log := range logs {
			var els []string
			if len(p.pods) > 1 {
				els = append(els, pod.GetName())
			}
			els = append(els, log)
			p.logCh <- strings.Join(els, " | ")
		}
	}
}

var compile = regexp.MustCompile(`\[\d+(;\d+)*m`)

func (p *LogListPage) addItem(txt string) {
	start := time.Now()
	if p.isPrint {
		os.Stdout.WriteString(txt + "\n")
	}
	if compile.MatchString(txt) {
		txt = compile.ReplaceAllString(txt, "**")
	}
	//txt = strings.ReplaceAll(txt, " ", "\n")
	content := widgets.NewTappableLabel(txt)
	content.Wrapping = fyne.TextWrapBreak
	content.OnTapped = p.contentTapped(txt)
	p.list.Add(content)
	dur := time.Now().Sub(start)
	if sub := p.minScrollDuration - dur; sub > 0 {
		time.Sleep(sub)
	}
	p.vScroll.ScrollToBottom()
}

func (p *LogListPage) contentTapped(txt string) func() {
	return func() {
		p.logDetail.Show(fas.TernaryOp(len(p.pods) == 1, p.pods[0].GetName(), "M:"+p.pods[0].GetNamespace()), txt)
	}
}
