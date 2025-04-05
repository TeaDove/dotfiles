package kwatch

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/kube_supplier"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/pkg/errors"
	"sync"
)

type KWatch struct {
	kubeSupplier *kube_supplier.Supplier
	model        *model
}

func buildTable(cols ...string) (*table.Table, *gloss_utils.MappingData) {
	data := gloss_utils.NewMappingData(cols...)
	tableView := table.New().
		Wrap(true).
		Headers(cols...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#df8e1d"))).
		Data(data)

	return tableView, data
}

func New(kubeSupplier *kube_supplier.Supplier) *KWatch {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	r := &KWatch{kubeSupplier: kubeSupplier}
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

	r.model.deploymentsTable, r.model.deploymentsTableData = buildTable("name", "image", "replicas")
	r.model.statefulsetTable, r.model.statefulsetTableData = buildTable("name", "image", "replicas")
	r.model.cronjobTable, r.model.cronjobTableData = buildTable("name", "image", "replicas")
	r.model.containersTable, r.model.containersTableData = buildTable("name", "image", "status")

	return r
}

func (r *KWatch) Run(ctx context.Context) error {
	p := tea.NewProgram(r.model, tea.WithContext(ctx))
	var wg sync.WaitGroup
	wg.Add(1)
	go r.viewContainers(ctx, &wg)

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run tea")
	}

	return nil
}

type ContainersAgg struct {
	Image     string
	BadStates []string

	Ready                 uint64
	Count                 uint64
	CpuUsageTotalMicro    uint64
	MemUsageTotalMegaByte uint64
}

func (r *KWatch) viewContainers(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	containers, err := r.kubeSupplier.GetContainerInfo(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to get container info"))
		return
	}

	for _, container := range containers {
		r.model.containersTableData.SetMappingRow(container.PodName,
			gloss_utils.M{
				"name":   container.PodName,
				"image":  container.Image,
				"status": container.State,
			})
	}
}
