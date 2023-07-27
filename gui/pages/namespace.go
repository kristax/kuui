package pages

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kristax/kuui/gui/widgets"
	"github.com/kristax/kuui/kucli"
	"github.com/samber/lo"
	coreV1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)

type NamespacePage struct {
	mainWindow fyne.Window
	cli        kucli.KuCli
	namespace  string
	addTabFn   func(name string, content func(ctx context.Context) fyne.CanvasObject)
}

func NewNamespace(mainWindow fyne.Window, cli kucli.KuCli, namespace string, addTabFn func(name string, content func(ctx context.Context) fyne.CanvasObject)) *NamespacePage {
	return &NamespacePage{
		mainWindow: mainWindow,
		cli:        cli,
		namespace:  namespace,
		addTabFn:   addTabFn,
	}
}

func (p *NamespacePage) Build(ctx context.Context) fyne.CanvasObject {
	grid := container.NewAdaptiveGrid(1)
	scroll := container.NewVScroll(grid)

	grid.Add(p.buildDeploymentTable(ctx))
	grid.Add(p.buildServicesTable(ctx))
	grid.Add(p.buildPodsTable(ctx))

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Second):
				grid.Objects[2] = p.buildPodsTable(ctx)
				grid.Refresh()
			}
		}
	}()

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
	deployments, err := p.cli.GetDeployments(ctx, p.namespace)
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
			result, err := p.cli.GetDeployment(ctx, p.namespace, deployment.GetName())
			if err != nil {
				fyne.LogError("get deployment", err)
				return
			}
			result.Spec.Replicas = &i32
			result, err = p.cli.UpdateDeployment(ctx, p.namespace, result)
			if err != nil {
				fyne.LogError("update deployment", err)
				return
			}
			go func() {
				for result.Status.ReadyReplicas != *result.Spec.Replicas {
					time.Sleep(time.Millisecond * 200)
					result, err = p.cli.GetDeployment(ctx, p.namespace, result.GetName())
					if err != nil {
						fyne.LogError("get deployment", err)
						return
					}
					deployments[id.Row] = *result
					table.Refresh()
				}
			}()

		}, p.mainWindow)
		table.Unselect(id)
	})
	return widget.NewCard("Deployments", "", table)
}

func (p *NamespacePage) buildPodsTable(ctx context.Context) *widget.Card {
	pods, err := p.cli.GetPods(ctx, p.namespace)
	if err != nil {
		fyne.LogError("get deployments", err)
		return nil
	}

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
	}, func(id widget.TableCellID, table *widget.Table) {
		selPod := pods[id.Row]
		pod, err := p.cli.GetPod(ctx, p.namespace, selPod.GetName())
		if err != nil {
			fyne.LogError("get pod", err)
			return
		}
		p.addTabFn(pod.GetName(), func(ctx context.Context) fyne.CanvasObject {
			return NewPodPage(p.cli, pod, p.addTabFn).Build(ctx)
		})
		table.Unselect(id)
	})

	return widget.NewCard("Pods", "", table)
}

func (p *NamespacePage) buildServicesTable(ctx context.Context) *widget.Card {
	services, err := p.cli.GetServices(ctx, p.namespace)
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
