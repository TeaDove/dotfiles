package net_traceroute

import (
	"dotfiles/pkg/cli/gloss_utils"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
	"github.com/teadove/teasutils/utils/time_utils"
)

type keymap struct {
	quit key.Binding
}

type model struct {
	spinner spinner.Model
	help    help.Model
	keymap  keymap

	target string

	traceTable     table.Table
	traceTableData *gloss_utils.MappingData
	service        *Service
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

const (
	colTTL      = "ttl"
	colIP       = "ip"
	colRTT      = "rtt"
	colDomains  = "domains"
	colLocation = "location"
	colISP      = "isp"
)

var tableCols = []string{colTTL, colIP, colRTT, colDomains, colLocation, colISP}

func (r *model) updateTable() {
	r.traceTableData.Locker().Lock()
	defer r.traceTableData.Locker().Unlock()

	for _, hop := range r.service.hops {
		ttlString := strconv.Itoa(int(hop.ttl))
		if r.traceTableData.RowExists(ttlString) {
			continue
		}

		if hop.peer == nil {
			r.traceTableData.Set(colTTL, ttlString, ttlString)
			r.traceTableData.Set(colIP, ttlString, "* * *")

			continue
		}

		r.traceTableData.Set(colTTL, ttlString, ttlString)
		r.traceTableData.Set(colIP, ttlString, hop.peer.String())
		r.traceTableData.Set(colRTT, ttlString, time_utils.RoundDuration(hop.rtt))

		if hop.domains.Err == nil {
			r.traceTableData.Set(colDomains, ttlString, hop.domains.Ok)
		}

		if hop.location.Err == nil {
			r.traceTableData.Set(colLocation, ttlString,
				fmt.Sprintf("%s %s %s",
					color.GreenString(hop.location.Ok.Country),
					hop.location.Ok.RegionName,
					hop.location.Ok.City,
				),
			)
			r.traceTableData.Set(colISP, ttlString,
				fmt.Sprintf("%s %s %s",
					hop.location.Ok.Isp,
					color.BlueString(hop.location.Ok.Org),
					hop.location.Ok.As,
				),
			)
		}
	}
}

func (r *model) View() string {
	r.traceTableData.RLocker().Lock()
	defer r.traceTableData.RLocker().Unlock()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		r.target,
		r.traceTable.String(),
	) + r.helpView()
}

func (r *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "й":
			return r, tea.Quit
		}
	default:
		r.spinner, cmd = r.spinner.Update(msg)
	}

	return r, cmd
}

func (r *model) Init() tea.Cmd {
	return r.spinner.Tick
}
