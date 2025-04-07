package kwatch

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/kube_supplier"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/fatih/color"
	"github.com/monirz/romnum"
	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/conv_utils"
	"github.com/teadove/teasutils/utils/time_utils"
	"strings"
	"time"
)

type KWatch struct {
	kubeSupplier *kube_supplier.Supplier
	model        *model

	crs map[string]string
}

func buildTable(cols ...string) (*table.Table, *gloss_utils.MappingData) {
	data := gloss_utils.NewMappingData(cols...)
	tableView := table.New().
		Wrap(true).
		Headers(cols...).
		Border(lipgloss.ThickBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#df8e1d"))).
		Data(data)

	return tableView, data
}

func New(kubeSupplier *kube_supplier.Supplier) *KWatch {
	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Spinner.FPS = time.Second / 2
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	r := &KWatch{kubeSupplier: kubeSupplier, crs: make(map[string]string)}
	r.model = &model{
		spinner: s,
		help:    help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
	}

	r.model.statefulsetTable, r.model.statefulsetTableData = buildTable("name", "image", "replicas")
	r.model.cronjobTable, r.model.cronjobTableData = buildTable("name", "image", "replicas")

	r.model.deploymentsTable, r.model.deploymentsTableData = buildTable("name", "replicas", "updated at", "images", "prevision images", "images updated at")
	r.model.containersTable, r.model.containersTableData = buildTable("name", "image", "is ok", "age", "cpu", "mem")

	return r
}

func (r *KWatch) Run(ctx context.Context) error {
	p := tea.NewProgram(r.model, tea.WithContext(ctx), tea.WithAltScreen())
	go r.update(ctx)

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run tea")
	}

	return nil
}

func (r *KWatch) update(ctx context.Context) {
	for {
		r.viewContainers(ctx)
		r.viewDeployments(ctx)
		time.Sleep(time.Millisecond * 300)
	}
}

func (r *KWatch) getCRKeys(images []string) []string {
	newImages := make([]string, 0)
	for _, image := range images {
		newImages = append(newImages, r.getCRKey(image))
	}

	return newImages
}

func (r *KWatch) getCRKey(image string) string {
	fields := strings.Split(image, "/")
	if len(fields) < 3 {
		return image
	}

	crIdx := len(fields) - 1
	cr := strings.Join(fields[:crIdx], "/")
	k, ok := r.crs[cr]
	if ok {
		return strings.Join(append([]string{k}, fields[crIdx:]...), "/")
	}

	k, err := romnum.Convert(len(r.crs) + 1)
	if err != nil {
		return image
	}
	k = color.WhiteString(fmt.Sprintf("(%s)", k))

	r.crs[cr] = k
	return strings.Join(append([]string{k}, fields[crIdx:]...), "/")
}

func makeContainerName(containers []*kube_supplier.ContainerInfo) map[kube_supplier.ContainerInfo]string {
	podNames := mapset.NewSet[string]()
	res := make(map[kube_supplier.ContainerInfo]string)

	for _, container := range containers {
		name := container.PodName

		exists := podNames.Add(name)
		if !exists {
			name += fmt.Sprintf(" (%s)", container.Name)
			podNames.Add(name)
		}
		res[*container] = name
	}

	return res
}

func colorIfLoaded(state float64, max float64) func(string, ...any) string {
	load := state / max
	switch {
	case load >= 0.45:
		return color.YellowString
	case load >= 0.9:
		return color.RedString
	case load >= 1:
		return color.HiRedString
	default:
		return fmt.Sprintf
	}
}

func (r *KWatch) viewContainers(ctx context.Context) {
	containers, err := r.kubeSupplier.GetContainersInfo(ctx)
	if err != nil {
		return
	}

	r.model.drawLock.Lock()
	defer r.model.drawLock.Unlock()

	containerNames := makeContainerName(containers)
	r.model.containersTableData.Clear()

	for _, container := range containers {
		state := "ok"
		if !container.Ready {
			state = color.RedString(container.State)
		}

		cpuState := colorIfLoaded(float64(container.CpuUsageMilli), float64(container.CpuRequestedMilli))(fmt.Sprintf("%.1f", float64(container.CpuUsageMilli)/10.0))
		memState := colorIfLoaded(float64(container.MemUsageKiloByte), float64(container.MemRequestedKiloByte))(conv_utils.ClosestByte(container.MemUsageKiloByte * 1024))

		r.model.containersTableData.SetMappingRow(
			containerNames[*container],
			gloss_utils.M{
				"name":  containerNames[*container],
				"image": r.getCRKey(container.Image),
				"is ok": state,
				"age":   time_utils.RoundDuration(time.Since(container.CreatedAt)),
				"cpu":   fmt.Sprintf("%s%%/%.1f%%", cpuState, float64(container.CpuRequestedMilli)/10.0),
				"mem":   fmt.Sprintf("%s/%s", memState, conv_utils.ClosestByte(container.MemRequestedKiloByte*1024)),
			},
		)
	}
}

func (r *KWatch) viewDeployments(ctx context.Context) {
	deployemnts, err := r.kubeSupplier.GetDeploymentInfo(ctx)
	if err != nil {
		return
	}

	for _, deployment := range deployemnts {
		readyReplicas := fmt.Sprintf("%d", deployment.ReadyReplicas)
		if !deployment.Ready {
			readyReplicas = color.RedString(readyReplicas)
		}

		// "name", "replicas", "updated at", "images", "prevision images", "images updated at"
		glossMapping := gloss_utils.M{
			"name":       deployment.Name,
			"replicas":   fmt.Sprintf("%s/%d", readyReplicas, deployment.RequestedReplicas),
			"updated at": time_utils.RoundDuration(time.Since(deployment.UpdatedAt)),
			"images":     strings.Join(r.getCRKeys(deployment.Images), ","),
		}
		if len(deployment.PrevImages) != 0 {
			glossMapping["prevision images"] = strings.Join(r.getCRKeys(deployment.PrevImages), ",")
			glossMapping["images updated at"] = time_utils.RoundDuration(time.Since(deployment.ImageUpdatedAt))
		}

		r.model.deploymentsTableData.SetMappingRow(deployment.Name, glossMapping)
	}
}
