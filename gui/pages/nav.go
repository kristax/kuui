package pages

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/channels"
	"github.com/kristax/kuui/gui/preference"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/kristax/kuui/kucli"
	"github.com/kristax/kuui/streamer"
	"github.com/kristax/kuui/themes"
	"github.com/kristax/kuui/util/fas"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Nav struct {
	MainWindow *MainWindow     `wire:""`
	KuCli      kucli.KuCli     `wire:""`
	Streamer   streamer.Client `wire:""`

	pods                []v1.Pod
	navItems            map[string]*v1.Pod
	onNamespaceSelected func(namespace string)
	onPodSelected       func(pod *v1.Pod)
	treeData            binding.StringTree
}

func NewNav() *Nav {
	return &Nav{
		treeData: binding.NewStringTree(),
	}
}

func (u *Nav) Init() error {
	u.onNamespaceSelected = func(namespace string) {
		u.MainWindow.AddTab(namespace, func(ctx context.Context) fyne.CanvasObject {
			return newNamespace(u.MainWindow, namespace).Build(ctx)
		})
	}
	u.onPodSelected = func(pod *v1.Pod) {
		u.MainWindow.AddTab(pod.GetName(), func(ctx context.Context) fyne.CanvasObject {
			return newLogListPage(u.MainWindow, []*v1.Pod{pod}).Build(ctx)
		})
	}
	return u.loadResources()
}

func (u *Nav) Build() *fyne.Container {
	app := fyne.CurrentApp()

	themeDark := app.Preferences().BoolWithFallback(preference.ThemeDark, app.Settings().ThemeVariant() == theme.VariantDark)
	app.Settings().SetTheme(fas.TernaryOp(themeDark, themes.DarkTheme(), themes.LightTheme()))

	btnTheme := widget.NewButton(fas.TernaryOp(themeDark, "Light", "Dark"), nil)
	btnReload := widget.NewButton("Reload", nil)
	navBottom := container.NewGridWithColumns(2, btnReload, btnTheme)

	navTree, refresh := u.buildNav()
	defer refresh(u.pods)
	nav := container.NewBorder(nil, navBottom, nil, nil, navTree)

	btnTheme.OnTapped = func() {
		go func() {
			themeDark = !themeDark
			app.Preferences().SetBool(preference.ThemeDark, themeDark)
			if themeDark {
				app.Settings().SetTheme(themes.DarkTheme())
				btnTheme.SetText("Light")
			} else {
				app.Settings().SetTheme(themes.LightTheme())
				btnTheme.SetText("Dark")
			}
		}()
	}

	btnReload.OnTapped = func() {
		btnReload.Disable()
		defer btnReload.Enable()
		err := u.loadResources()
		if err != nil {
			panic(err)
		}
		refresh(u.pods)
	}

	return nav
}

func (u *Nav) buildNav() (fyne.CanvasObject, func(pods []v1.Pod)) {
	var (
		pods      = u.pods
		page      = 1
		size      = 20
		searchStr = fyne.CurrentApp().Preferences().String(preference.NavSearchStr)
	)

	var tree = u.buildTree()
	txtSearch := widget.NewEntry()
	btnForward := widget.NewButton("<", nil)
	btnNext := widget.NewButton(">", nil)

	txtPage := widgets.NewNumericalEntry()
	txtPage.SetText(fmt.Sprintf("%v", page))
	lbTotalPage := widget.NewLabel("")
	compPage := container.NewGridWithColumns(2, txtPage, lbTotalPage)
	pagination := container.NewBorder(nil, nil, btnForward, btnNext, compPage)
	nav := container.NewBorder(txtSearch, pagination, nil, nil, tree)

	var refreshTreeFunc = func(pods []v1.Pod) {
		if searchStr != "" {
			pods = lo.Filter(pods, func(item v1.Pod, _ int) bool {
				return strings.Contains(item.GetNamespace(), searchStr) || strings.Contains(item.GetName(), searchStr)
			})
		}
		total := len(lo.GroupBy[v1.Pod, string](pods, func(item v1.Pod) string {
			return item.GetNamespace()
		}))
		sumPage := int(math.Ceil(float64(total) / float64(size)))
		if page < 1 {
			page = 1
		}
		if page > sumPage {
			page = sumPage
		}
		txtPage.SetText(fmt.Sprintf("%v", page))
		lbTotalPage.SetText(fmt.Sprintf("of %d", sumPage))
		u.refreshData(pods, page, size)
		nav.Refresh()
	}

	txtPage.OnChanged = func(s string) {
		page, _ = strconv.Atoi(s)
		refreshTreeFunc(pods)
	}

	btnNext.OnTapped = func() {
		page++
		refreshTreeFunc(pods)
	}
	btnForward.OnTapped = func() {
		page--
		refreshTreeFunc(pods)
	}

	txtSearch.SetPlaceHolder("search")
	txtSearch.SetText(searchStr)
	txtSearch.OnChanged = func(s string) {
		page = 1
		searchStr = s
		fyne.CurrentApp().Preferences().SetString(preference.NavSearchStr, searchStr)
		refreshTreeFunc(u.pods)
	}

	return nav, refreshTreeFunc
}

func (u *Nav) buildTree() *widget.Tree {
	collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
	var create = func(b bool) fyne.CanvasObject {
		if b {
			return container.NewBorder(nil, nil, nil, widget.NewCheck("", nil), widget.NewLabel(""))
		}
		return widget.NewLabel("")
	}
	update := func(item binding.DataItem, b bool, object fyne.CanvasObject) {
		var lb *widget.Label
		bs := item.(binding.String)
		id, _ := bs.Get()
		if b {
			border := object.(*fyne.Container)
			lb = border.Objects[0].(*widget.Label)
			btn := border.Objects[1].(*widget.Check)
			contains := lo.Contains(collections, id)
			if contains {
				btn.SetChecked(true)
			}
			btn.OnChanged = u.manageCollection(id)
		} else {
			lb = object.(*widget.Label)
		}
		lb.SetText(id)
	}
	tree := widget.NewTreeWithData(u.treeData, create, update)
	tree.OnSelected = func(uid widget.TreeNodeID) {
		if t, ok := u.navItems[uid]; ok {
			var err error
			t, err = u.KuCli.GetPod(context.Background(), t.GetNamespace(), t.GetName())
			if err != nil {
				dialog.ShowError(fmt.Errorf("%v, please reload", err), u.MainWindow.Content())
				return
			}
			u.onPodSelected(t)
		} else {
			u.onNamespaceSelected(uid)
		}
		tree.Unselect(uid)
	}
	return tree
}

func (u *Nav) manageCollection(id string) func(b bool) {
	return func(b bool) {
		collections := fyne.CurrentApp().Preferences().StringList(preference.NSCollections)
		if b {
			collections = append(collections, id)
		} else {
			collections = lo.Filter(collections, func(item string, _ int) bool {
				return item != id
			})
		}
		sort.Slice(collections, func(i, j int) bool {
			return collections[i] < collections[j]
		})
		fyne.CurrentApp().Preferences().SetStringList(preference.NSCollections, collections)
		u.Streamer.Send(context.Background(), channels.CollectionsUpdate, nil)
	}
}

func (u *Nav) refreshData(pods []v1.Pod, page, size int) {
	values := make(map[string]string)
	nsGroup := lo.GroupBy[v1.Pod, string](pods, func(item v1.Pod) string {
		return item.GetNamespace()
	})

	var (
		maxLen = len(nsGroup)
		offset int
		limit  = maxLen
	)

	if page > 0 && size > 0 {
		offset = fas.Min((page-1)*size, maxLen)
		limit = fas.Min(page*size, maxLen)
	}

	namespaces := lo.Keys(nsGroup)
	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i] < namespaces[j]
	})

	namespaces = namespaces[offset:limit]

	navIndex := map[string][]string{
		"": namespaces,
	}
	for _, namespace := range namespaces {
		values[namespace] = namespace
		podNames := lo.Map[v1.Pod, string](nsGroup[namespace], func(item v1.Pod, _ int) string {
			values[item.GetName()] = item.GetName()
			return item.GetName()
		})
		navIndex[namespace] = append(navIndex[namespace], podNames...)
	}
	u.treeData.Set(navIndex, values)
}

func (u *Nav) loadResources() error {
	ctx := context.Background()
	pods, err := u.KuCli.GetPods(ctx, "")
	if err != nil {
		return err
	}
	u.pods = pods
	u.navItems = lo.SliceToMap[v1.Pod, string, *v1.Pod](u.pods, func(item v1.Pod) (string, *v1.Pod) {
		return item.GetName(), &item
	})
	return nil
}
