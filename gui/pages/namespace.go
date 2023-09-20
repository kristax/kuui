package pages

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/samber/lo"
	coreV1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)

type NamespacePage struct {
	mainWindow *MainWindow
	namespace  string
}

func newNamespace(mainWindow *MainWindow, namespace string) *NamespacePage {
	return &NamespacePage{
		mainWindow: mainWindow,
		namespace:  namespace,
	}
}

func (p *NamespacePage) Build(ctx context.Context) fyne.CanvasObject {
	grid := container.NewAdaptiveGrid(1)
	scroll := container.NewVScroll(grid)

	grid.Add(p.buildDeploymentTable(ctx))
	grid.Add(p.buildPodsCard(ctx))
	grid.Add(p.buildServicesTable(ctx))

	return scroll
}

func buildTable(row, col int, updateHeader func(id widget.TableCellID, label *widget.Label),
	updateCell func(id widget.TableCellID, label *widget.Label),
	onSelect func(id widget.TableCellID, table *widget.Table)) *widget.Table {
	table := widget.NewTable(func() (int, int) {
		return row, col
	}, func() fyne.CanvasObject {
		label := widget.NewLabel("")
		return label
	}, nil)

	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		updateHeader(id, template.(*widget.Label))
	}
	table.UpdateCell = func(id widget.TableCellID, template fyne.CanvasObject) {
		updateCell(id, template.(*widget.Label))
	}
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < 0 || id.Row >= row {
			return
		}
		onSelect(id, table)
	}
	table.SetColumnWidth(0, 400)
	for i := 1; i < col; i++ {
		table.SetColumnWidth(i, 200)
	}
	table.ShowHeaderRow = true
	return table
}

func (p *NamespacePage) buildDeploymentTable(ctx context.Context) *widget.Card {
	deployments, err := p.mainWindow.KuCli.GetDeployments(ctx, p.namespace)
	if err != nil {
		fyne.LogError("get deployments", err)
		return nil
	}
	table := buildTable(len(deployments), 4, func(id widget.TableCellID, label *widget.Label) {
		switch id.Col {
		case 0:
			label.SetText("NAME")
		case 1:
			label.SetText("READY")
		case 2:
			label.SetText("UP-TO-DATE")
		case 3:
			label.SetText("AVAILABLE")
		}
	}, func(id widget.TableCellID, label *widget.Label) {
		deployment := deployments[id.Row]
		switch id.Col {
		case 0:
			name := deployment.GetName()
			label.SetText(name)
		case 1:
			label.SetText(fmt.Sprintf("%d / %d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas))
		case 2:
			label.SetText(fmt.Sprintf("%d", deployment.Status.UpdatedReplicas))
		case 3:
			label.SetText(fmt.Sprintf("%d", deployment.Status.AvailableReplicas))
		}
	}, func(id widget.TableCellID, table *widget.Table) {
		deployment := deployments[id.Row]
		sr := fmt.Sprintf("%d", *deployment.Spec.Replicas)
		txtScale := widgets.NewNumericalEntry()
		txtScale.SetText(sr)
		box := container.NewVBox(
			container.NewGridWithColumns(2, widget.NewLabel("Current Replicas: "), widget.NewLabel(sr)),
			container.NewGridWithColumns(2, widget.NewLabel("Scale To: "), txtScale),
		)
		dialog.ShowCustomConfirm("Scale Replicas", "Scale", "Cancel", box, func(b bool) {
			replica, _ := strconv.Atoi(txtScale.Text)
			i32 := int32(replica)
			result, err := p.mainWindow.KuCli.GetDeployment(ctx, p.namespace, deployment.GetName())
			if err != nil {
				fyne.LogError("get deployment", err)
				return
			}
			result.Spec.Replicas = &i32
			result, err = p.mainWindow.KuCli.UpdateDeployment(ctx, p.namespace, result)
			if err != nil {
				fyne.LogError("update deployment", err)
				return
			}
			go func() {
				for result.Status.ReadyReplicas != *result.Spec.Replicas {
					time.Sleep(time.Millisecond * 200)
					result, err = p.mainWindow.KuCli.GetDeployment(ctx, p.namespace, result.GetName())
					if err != nil {
						fyne.LogError("get deployment", err)
						return
					}
					deployments[id.Row] = *result
					table.Refresh()
				}
			}()

		}, p.mainWindow.Content())
		table.Unselect(id)
	})
	return widget.NewCard("Deployments", "", table)
}

func (p *NamespacePage) buildPodsCard(ctx context.Context) *widget.Card {
	pods, err := p.mainWindow.KuCli.GetPods(ctx, p.namespace)
	if err != nil {
		fyne.LogError("get deployments", err)
		return nil
	}

	var selectedPods []*coreV1.Pod
	entryPods := widget.NewEntry()
	var refreshEntryPods = func() {
		names := lo.Map(selectedPods, func(item *coreV1.Pod, _ int) string {
			return item.GetName()
		})
		entryPods.SetText(strings.Join(names, ", "))
	}

	var onSelectFn = func(id widget.TableCellID, table *widget.Table) {
		defer refreshEntryPods()
		defer table.Unselect(id)
		selPod := pods[id.Row]
		contains := lo.ContainsBy(selectedPods, func(item *coreV1.Pod) bool {
			return selPod.GetName() == item.GetName()
		})
		if contains {
			selectedPods = lo.Reject(selectedPods, func(item *coreV1.Pod, _ int) bool {
				return selPod.GetName() == item.GetName()
			})
			return
		}
		pod, err := p.mainWindow.KuCli.GetPod(ctx, p.namespace, selPod.GetName())
		if err != nil {
			fyne.LogError("get pod", err)
			return
		}
		selectedPods = append(selectedPods, pod)
	}

	table := buildPodsTable(pods, onSelectFn)

	var openLogPageFn = func() {
		if len(selectedPods) > 0 {
			var name = selectedPods[0].GetName()
			if len(selectedPods) > 1 {
				name += "..."
			}
			p.mainWindow.AddTab(name, func(ctx context.Context) fyne.CanvasObject {
				return newLogListPage(p.mainWindow, selectedPods).Build(ctx)
			})
		}
	}

	entryPods.OnSubmitted = func(s string) {
		openLogPageFn()
	}
	btnAll := widget.NewButton("All", func() {
		selectedPods = make([]*coreV1.Pod, 0, len(pods))
		selectedPods = lo.Map(pods, func(item coreV1.Pod, _ int) *coreV1.Pod {
			return &item
		})
		refreshEntryPods()
	})
	btnLogs := widget.NewButton("Logs", openLogPageFn)
	btnClear := widget.NewButton("Clear", func() {
		selectedPods = nil
		refreshEntryPods()
	})

	header := container.NewBorder(nil, nil, nil, container.NewHBox(btnAll, btnClear, btnLogs), entryPods)
	stack := container.NewStack(table)
	border := container.NewBorder(header, nil, nil, nil, stack)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Second * 3):
				pods, err = p.mainWindow.KuCli.GetPods(ctx, p.namespace)
				if err != nil {
					fyne.LogError("get deployments", err)
					return
				}
				stack.RemoveAll()
				stack.Add(buildPodsTable(pods, onSelectFn))
				stack.Refresh()
			}
		}
	}()

	return widget.NewCard("Pods", "", border)
}

func buildPodsTable(pods []coreV1.Pod, onSelect func(id widget.TableCellID, table *widget.Table)) *widget.Table {
	table := buildTable(len(pods), 4, func(id widget.TableCellID, label *widget.Label) {
		switch id.Col {
		case 0:
			label.SetText("NAME")
		case 1:
			label.SetText("STATUS")
		case 2:
			label.SetText("AGE")
		case 3:
			label.SetText("RESTART")
		}
	}, func(id widget.TableCellID, label *widget.Label) {
		pod := pods[id.Row]
		switch id.Col {
		case 0:
			name := pod.GetName()
			label.SetText(name)
		case 1:
			label.SetText(string(pod.Status.Phase))
		case 2:
			label.SetText(fmtDuration(time.Now().Sub(pod.GetCreationTimestamp().Time)))
		case 3:
			if len(pod.Status.ContainerStatuses) > 0 {
				label.SetText(fmt.Sprintf("%d", pod.Status.ContainerStatuses[0].RestartCount))
			}
		}
	}, onSelect)

	return table
}

func (p *NamespacePage) buildServicesTable(ctx context.Context) *widget.Card {
	services, err := p.mainWindow.KuCli.GetServices(ctx, p.namespace)
	if err != nil {
		fyne.LogError("get deployments", err)
		return nil
	}

	table := buildTable(len(services), 6, func(id widget.TableCellID, label *widget.Label) {
		switch id.Col {
		case 0:
			label.SetText("NAME")
		case 1:
			label.SetText("TYPE")
		case 2:
			label.SetText("CLUSTER-IP")
		case 3:
			label.SetText("EXTERNAL-IP")
		case 4:
			label.SetText("PORT(S)")
		case 5:
			label.SetText("AGE")
		}
	}, func(id widget.TableCellID, label *widget.Label) {
		service := services[id.Row]
		switch id.Col {
		case 0:
			name := service.GetName()
			label.SetText(name)
		case 1:
			label.SetText(string(service.Spec.Type))
		case 2:
			label.SetText(service.Spec.ClusterIP)
		case 3:
			label.SetText(strings.Join(service.Spec.ExternalIPs, ","))
		case 4:
			s := lo.Map[coreV1.ServicePort, string](service.Spec.Ports, func(item coreV1.ServicePort, _ int) string {
				return fmt.Sprintf("%d/%s", item.Port, item.Protocol)
			})
			label.SetText(strings.Join(s, ","))
		case 5:
			label.SetText(fmtDuration(time.Now().Sub(service.GetCreationTimestamp().Time)))
		}
	}, func(id widget.TableCellID, table *widget.Table) {})

	return widget.NewCard("Services", "", table)
}

func fmtDuration(duration time.Duration) string {
	var t []string
	if d := int64(duration.Hours() / 24); d >= 1 {
		t = append(t, fmt.Sprintf("%dd", d))
		duration -= time.Duration(d * 24 * int64(time.Hour))
	}
	if h := int64(duration.Minutes() / 60); h >= 1 {
		t = append(t, fmt.Sprintf("%dh", h))
		duration -= time.Duration(h * 60 * int64(time.Minute))
	}
	if m := int64(duration.Seconds() / 60); m >= 1 {
		t = append(t, fmt.Sprintf("%dm", m))
		duration -= time.Duration(m * 60 * int64(time.Second))
	}
	if sec := int64(duration.Seconds()); sec >= 1 {
		t = append(t, fmt.Sprintf("%ds", sec))
	}
	return strings.Join(t, "")
}
