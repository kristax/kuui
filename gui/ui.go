package gui

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/kristax/kuui/gui/pages"
	"github.com/kristax/kuui/kucli"
	v1 "k8s.io/api/core/v1"
	"log"
)

type ui struct {
	app        fyne.App
	mainWindow fyne.Window
	KuCli      kucli.KuCli `wire:""`
}

func NewUI() any {
	a := app.NewWithID("com.kristas.kuui")
	return &ui{
		app:        a,
		mainWindow: a.NewWindow("KuUi"),
	}
}

func (u *ui) Order() int {
	return 0
}

func (u *ui) Run() error {
	u.mainWindow.SetMaster()
	u.mainWindow.ShowAndRun()
	return nil
}

func (u *ui) Init() error {
	u.makeTray()
	u.logLifecycle()
	u.makeMenu()

	content := container.NewMax()
	tabs := container.NewDocTabs()
	content.Add(tabs)
	welcomePage := pages.WelcomePage{}
	tabs.Append(container.NewTabItem("Welcome", welcomePage.Build()))

	var existTabItems = make(map[string]*tabContent)
	var addTabFn = func(name string, content func(ctx context.Context) fyne.CanvasObject) {
		tc, ok := existTabItems[name]
		if !ok {
			ctx, cancelFunc := context.WithCancel(context.Background())
			item := container.NewTabItem(name, content(ctx))
			tabs.Append(item)
			tc = &tabContent{
				Item:       item,
				CancelFunc: cancelFunc,
			}
			existTabItems[name] = tc
		}
		tabs.Select(tc.Item)
		tabs.Refresh()
	}
	navPage := pages.NewNav(u.mainWindow, u.KuCli, func(namespace string) {

	}, func(pod *v1.Pod) {
		addTabFn(pod.GetName(), func(ctx context.Context) fyne.CanvasObject {
			return pages.NewPodPage(u.KuCli, pod, addTabFn).Build(ctx)
		})
	}).Build()
	tabs.OnClosed = func(item *container.TabItem) {
		if tc, ok := existTabItems[item.Text]; ok {
			tc.CancelFunc()
			delete(existTabItems, item.Text)
		}
	}

	mainFrame := container.NewHSplit(navPage, content)
	mainFrame.Offset = 0.2

	u.mainWindow.SetContent(mainFrame)

	u.mainWindow.Resize(fyne.NewSize(1920, 1080))
	u.mainWindow.CenterOnScreen()
	return nil
}

func (u *ui) makeTray() {
	if desk, ok := u.app.(desktop.App); ok {
		h := fyne.NewMenuItem("Hello", func() {})
		h.Icon = theme.HomeIcon()
		menu := fyne.NewMenu("Hello World", h)
		h.Action = func() {
			log.Println("System tray menu tapped")
			h.Label = "Welcome"
			menu.Refresh()
		}
		desk.SetSystemTrayMenu(menu)
	}
}

type tabContent struct {
	Item       *container.TabItem
	CancelFunc context.CancelFunc
}