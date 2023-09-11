package pages

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/kristax/kuui/kucli"
)

type MainWindow struct {
	App         fyne.App     `wire:""`
	KuCli       kucli.KuCli  `wire:""`
	WelcomePage *WelcomePage `wire:""`
	Nav         *Nav         `wire:""`
	//DiffPage    *DiffPage    `wire:""`
	MenuItems []MenuItem `wire:""`

	mainWindow    fyne.Window
	existTabItems map[string]*tabContent
	tabs          *container.DocTabs
}

func NewMainWindow() *MainWindow {
	return &MainWindow{
		existTabItems: make(map[string]*tabContent),
	}
}

func (u *MainWindow) ShowAndRun() error {
	u.mainWindow.SetMaster()
	u.mainWindow.ShowAndRun()
	return nil
}

func (u *MainWindow) Init() error {
	u.mainWindow = u.App.NewWindow("KuUi")
	u.tabs = container.NewDocTabs()

	u.AddTab("Welcome", func(ctx context.Context) fyne.CanvasObject {
		return u.WelcomePage.Build()
	})
	u.tabs.OnClosed = func(item *container.TabItem) {
		if tc, ok := u.existTabItems[item.Text]; ok {
			tc.CancelFunc()
			delete(u.existTabItems, item.Text)
		}
	}
	nav := u.Nav.Build()

	mainFrame := container.NewHSplit(nav, container.NewStack(u.tabs))
	mainFrame.Offset = 0.2

	u.mainWindow.SetContent(mainFrame)
	u.mainWindow.Resize(fyne.NewSize(1920, 1080))
	u.mainWindow.CenterOnScreen()
	u.makeMenu()
	return nil
}

func (u *MainWindow) AddTab(name string, content func(ctx context.Context) fyne.CanvasObject) {
	tc, ok := u.existTabItems[name]
	if !ok {
		ctx, cancelFunc := context.WithCancel(context.Background())
		item := container.NewTabItem(name, content(ctx))
		u.tabs.Append(item)
		tc = &tabContent{
			Item:       item,
			CancelFunc: cancelFunc,
		}
		u.existTabItems[name] = tc
	}
	u.tabs.Select(tc.Item)
	u.tabs.Refresh()
}

func (u *MainWindow) Content() fyne.Window {
	return u.mainWindow
}

type tabContent struct {
	Item       *container.TabItem
	CancelFunc context.CancelFunc
}
