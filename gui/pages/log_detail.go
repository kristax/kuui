package pages

import (
	"context"
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
)

type LogDetailPage struct {
	mainWindow *MainWindow
	logData    binding.String
	content    *fyne.Container
}

func newLogDetailPage(mainWindow *MainWindow) *LogDetailPage {
	logData := binding.NewString()
	txtLog := widget.NewMultiLineEntry()

	tbiUndo := widget.NewToolbarAction(theme.ContentUndoIcon(), nil)
	tbiRedo := widget.NewToolbarAction(theme.ContentRedoIcon(), nil)
	tbiFormat := widget.NewToolbarAction(theme.VisibilityIcon(), nil)
	toolbar := widget.NewToolbar()
	content := container.NewBorder(toolbar, nil, nil, nil, txtLog)

	tbiUndo.OnActivated = func() {
	}
	toolbar.Append(tbiUndo)

	tbiRedo.OnActivated = func() {
	}
	toolbar.Append(tbiRedo)

	tbiFormat.OnActivated = func() {
		txt, _ := logData.Get()
		s, err := cutJson(txt)
		if err != nil {
			return
		}
		logData.Set(s)
	}
	toolbar.Append(tbiFormat)

	txtLog.Bind(logData)
	txtLog.Wrapping = fyne.TextWrapWord
	txtLog.OnChanged = func(s string) {
		f, err := formatLog(s)
		if err != nil {
			f = s
		}
		logData.Set(f)
	}
	return &LogDetailPage{
		mainWindow: mainWindow,
		logData:    logData,
		content:    content,
	}
}

func (l *LogDetailPage) Show(pod, log string) {
	l.logData.Set(log)
	l.mainWindow.AddTab("L:"+pod, func(ctx context.Context) fyne.CanvasObject {
		return l.content
	})
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
