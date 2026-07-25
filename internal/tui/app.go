package tui

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/uphiago/barbarossa-cli/internal/config"
	"github.com/uphiago/barbarossa-cli/internal/docker"
)

type AppModel struct {
	tabs       []string
	activeTab  int
	workers    map[string]WorkerStatusMsg
	activities []ActivityMsg
	w, h       int
	showHelp   bool
	darkBg     bool
	docker     *docker.Client
	reconModel *ReconModel
	logModel   *LogsModel
	cfg        *config.Config
}

var workerC = map[string]color.Color{} // filled after theme

func initWorkerColors() {
	workerC = map[string]color.Color{
		"charlie": Charlie,
		"oscar":   Oscar,
		"papa":    Papa,
	}
}

func NewApp(cli *docker.Client, cfg *config.Config) *AppModel {
	return &AppModel{
		tabs:      []string{"Dashboard", "Recon", "Logs"},
		activeTab: 0,
		workers:   make(map[string]WorkerStatusMsg),
		docker:    cli,
		cfg:       cfg,
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(m.poll(), tick(), tea.RequestBackgroundColor)
}

func tick() tea.Cmd { return tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return TickMsg{} }) }

func (m *AppModel) poll() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if m.docker == nil { return m.offline() }
		cts, err := m.docker.ListContainers(ctx)
		if err != nil { return m.offline() }
		cm := make(map[string]docker.ContainerInfo, len(cts))
		for _, ct := range cts { cm[ct.Name] = ct }
		now := time.Now().UTC().Format("15:04")
		cmds := make([]tea.Cmd, 0, len(m.cfg.Containers.Names)*2)
		for _, n := range m.cfg.Containers.Names {
			ct, ok := cm[n]
			if !ok || ct.State != "running" {
				n := n
				cmds = append(cmds, func() tea.Msg { return WorkerStatusMsg{n, false, "\u2014", "\u2014", "\u2014"} })
				cmds = append(cmds, func() tea.Msg { return ActivityMsg{n, "offline", now} })
				continue
			}
			nc := n
			cmds = append(cmds, func() tea.Msg {
				s, err := m.docker.ContainerStats(ctx, nc)
				if err != nil { return WorkerStatusMsg{nc, true, "0%", "0MB", ct.Status} }
				return WorkerStatusMsg{nc, true, fmt.Sprintf("%.0f%%", s.CPUPercent), fmt.Sprintf("%.0fMB", s.MemoryMB), ct.Status}
			})
			cmds = append(cmds, func() tea.Msg { return ActivityMsg{nc, ct.State + " \u2014 " + ct.Status, now} })
		}
		return tea.Batch(cmds...)
	}
}

func (m *AppModel) offline() tea.Msg {
	now := time.Now().UTC().Format("15:04")
	cmds := make([]tea.Cmd, 0, len(m.cfg.Containers.Names)*2)
	for _, n := range m.cfg.Containers.Names {
		n := n
		cmds = append(cmds, func() tea.Msg { return WorkerStatusMsg{n, false, "\u2014", "\u2014", "\u2014"} })
		cmds = append(cmds, func() tea.Msg { return ActivityMsg{n, "docker unreachable", now} })
	}
	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = v.Width, v.Height
	case tea.KeyMsg:
		switch v.String() {
		case "q", "ctrl+c": return m, tea.Quit
		case "tab":          m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab":    m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		case "?":            m.showHelp = !m.showHelp
		case "r":            return m, m.poll()
		}
	case TickMsg:
		return m, tea.Batch(m.poll(), tick())
	case WorkerStatusMsg:
		m.workers[v.Name] = v
	case ActivityMsg:
		if v.Message == "" { break }
		m.activities = append(m.activities, v)
		if len(m.activities) > 50 { m.activities = m.activities[len(m.activities)-50:] }
	case tea.BackgroundColorMsg:
		m.darkBg = bool(v)
		SetAdaptive(m.darkBg)
		initWorkerColors()
	}

	if m.activeTab == 1 { if m.reconModel == nil { m.reconModel = NewReconModel() }; _, cmd := m.reconModel.Update(msg); return m, cmd }
	if m.activeTab == 2 { if m.logModel == nil { m.logModel = NewLogsModel() }; _, cmd := m.logModel.Update(msg); return m, cmd }
	return m, nil
}

func (m *AppModel) View() tea.View {
	if m.showHelp { v := tea.NewView(m.help()); v.AltScreen = true; return v }
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Top, m.tabBar(), m.body(), m.footer()))
}

func (m *AppModel) tabBar() string {
	var ts []string
	for i, t := range m.tabs {
		if i == m.activeTab { ts = append(ts, TabActive.Render(" "+t+" ")) } else { ts = append(ts, TabInactive.Render(" "+t+" ")) }
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, ts...)
	gw := m.w - lipgloss.Width(row) - 2; if gw < 0 { gw = 0 }
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, lipgloss.NewStyle().Foreground(BorderClr).Render(strings.Repeat("\u2500", gw)))
}

func (m *AppModel) body() string {
	var c string
	switch m.activeTab { case 0: c = m.dash(); case 1: c = m.recon(); case 2: c = m.logs() }
	return WindowFrame.Width(m.w-2).Height(m.h-6).Render(c)
}
func (m *AppModel) recon() string { if m.reconModel == nil { m.reconModel = NewReconModel() }; return m.reconModel.View().Content }
func (m *AppModel) logs() string  { if m.logModel == nil { m.logModel = NewLogsModel() }; return m.logModel.View().Content }

func (m *AppModel) dash() string {
	l := Bold.Foreground(Accent).Render("\u2694  BARBAROSSA")
	s := Sub().Render("  \u00b7  cluster monitor")
	clk := Sub().Render(time.Now().UTC().Format("15:04:05 UTC"))
	hdr := Panel.Background(Bg).Width(m.w-20).Render(lipgloss.JoinHorizontal(lipgloss.Center, l+s+"                ", clk))

	cardW := (m.w - 8) / 3; if cardW < 28 { cardW = (m.w - 6) / 2 }; if m.w < 70 { cardW = m.w - 6 }
	var cards []string
	for _, n := range m.cfg.Containers.Names {
		ws, ok := m.workers[n]; if !ok { ws = WorkerStatusMsg{n, false, "\u2014", "\u2014", "\u2014"} }
		cards = append(cards, WorkerCard(n, ws, workerC[n], cardW))
	}
	cRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	if m.w < 70 { cRow = lipgloss.JoinVertical(lipgloss.Left, cards...) }

	ft := Bold.Foreground(Accent).Render("\u25B8 ACTIVITY")
	var fl []string
	st := 0; if len(m.activities) > 8 { st = len(m.activities) - 8 }
	for _, a := range m.activities[st:] {
		tag := lipgloss.NewStyle().Foreground(workerC[a.Worker]).Bold(true).Width(8).Render(a.Worker)
		fl = append(fl, Sub().Width(5).Render(a.Time)+" "+tag+" "+a.Message)
	}
	if len(fl) == 0 { fl = append(fl, Sub().Render("  waiting for first poll...")) }
	fb := Panel.Render(lipgloss.JoinVertical(lipgloss.Left, ft, "", lipgloss.JoinVertical(lipgloss.Left, fl...)))

	return lipgloss.JoinVertical(lipgloss.Left, hdr, "", cRow, "", fb)
}

func (m *AppModel) footer() string {
	k := func(s string) string { return Mono().Render(s) }
	sep := Sub().Render("\u2502")
	left := k("TAB")+" tab  "+sep+"  "+k("R")+" refresh  "+sep+"  "+k("?")+" help"
	right := k("Q")+" quit"
	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right) - 4; if gap < 1 { gap = 1 }
	return lipgloss.NewStyle().Foreground(MutedClr).Padding(0,1).BorderTop(true).BorderForeground(BorderClr).Width(m.w-2).Render(left+strings.Repeat(" ",gap)+right)
}

func (m *AppModel) help() string {
	row := func(k, d string) string { return lipgloss.NewStyle().Foreground(Accent).Width(18).Render(k)+"  "+Sub().Render(d) }
	var b strings.Builder
	b.WriteString(Bold.Foreground(Accent).Render("\u2694  BARBAROSSA CLI \u2014 Help")+"\n\n")
	b.WriteString(Bold.Foreground(AccentDim).Render("Navigation")+"\n")
	b.WriteString(row("TAB / Shift+TAB","Switch tabs")+"\n")
	b.WriteString(row("\u2191 \u2193 Enter","Navigate + select")+"\n\n")
	b.WriteString(Bold.Foreground(AccentDim).Render("Actions")+"\n")
	b.WriteString(row("R","Refresh workers")+"\n")
	b.WriteString(row("?","Toggle help")+"\n")
	b.WriteString(row("Q / Ctrl+C","Quit")+"\n\n")
	b.WriteString(Bold.Foreground(AccentDim).Render("Per Tab")+"\n")
	b.WriteString(row("1-5 (Recon)","Filter severity")+"\n")
	b.WriteString(row("Space (Logs)","Pause/resume")+"\n\n")
	b.WriteString(Sub().Render("Press ? to close"))
	return Panel.BorderForeground(Accent).Padding(2,4).Width(55).Background(Bg).Render(b.String())
}
