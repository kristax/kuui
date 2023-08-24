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
	"regexp"
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
	tbiCut := widget.NewToolbarAction(theme.ContentCutIcon(), nil)
	toolbar := widget.NewToolbar()
	content := container.NewBorder(toolbar, nil, nil, nil, txtLog)

	var (
		histories    []string
		historyIndex = 0
		undoMode     = false
	)

	tbiUndo.OnActivated = func() {
		if historyIndex-1 >= 0 {
			undoMode = true
			historyIndex -= 1
			logData.Set(histories[historyIndex])
		}
	}
	toolbar.Append(tbiUndo)

	tbiRedo.OnActivated = func() {
		if historyIndex+1 < len(histories) {
			undoMode = true
			historyIndex += 1
			logData.Set(histories[historyIndex])
		}
	}
	toolbar.Append(tbiRedo)

	tbiCut.OnActivated = func() {
		txt, _ := logData.Get()
		s, err := cutJson(txt)
		if err != nil {
			return
		}
		logData.Set(s)
	}
	toolbar.Append(tbiCut)
	tbiFormat.OnActivated = func() {
		s, _ := logData.Get()
		f, err := formatLog(s)
		if err != nil {
			f = s
		}
		logData.Set(f)
	}
	toolbar.Append(tbiFormat)

	txtLog.Bind(logData)
	txtLog.Wrapping = fyne.TextWrapWord

	txtLog.OnChanged = func(s string) {
		if undoMode {
			undoMode = false
		} else {
			histories = append(histories, s)
			historyIndex = len(histories) - 1
			fmt.Println(len(histories))
		}
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
	txt = prettySQL(txt)
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

func prettySQL(sql string) string {
	sql = strings.TrimSpace(sql)
	re := regexp.MustCompile(`(?i)(SELECT|FROM|WHERE|LEFT\s+JOIN|INNER\s+JOIN|RIGHT\s+JOIN|ORDER\s+BY|GROUP\s+BY|LIMIT|SET|VALUES|UPDATE|HAVING|ADD\s+COLUMN|DROP\s+COLUMN|CREATE\s+TABLE|ALTER\s+TABLE|DELETE\s+FROM|UNION\s+ALL|UNION|EXCEPT|INTERSECT)`)
	sql = re.ReplaceAllString(sql, "\n$1")

	re2 := regexp.MustCompile(`(?i)(\,|\)')`)
	sql = re2.ReplaceAllString(sql, "$1\n")

	return sql
}
