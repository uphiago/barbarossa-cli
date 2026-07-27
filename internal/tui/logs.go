package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/uphiago/barbarossa-cli/internal/docker"
)

// Log entry

type LogEntryMsg struct {
	Worker    string
	Line      string
	Timestamp time.Time
}

// Logs Model

type LogsModel struct {
	viewport viewport.Model
	entries  []LogEntryMsg
	paused   bool
	width    int
	height   int
	source   logSource
	workers  []string
	events   chan LogEntryMsg
	started  bool
	cancel   context.CancelFunc
}

type logSource interface {
	ContainerLogs(context.Context, string, int) (docker.ContainerLogsResult, error)
}

type logStreamsClosedMsg struct{}

func NewLogsModel(source logSource, workers []string) *LogsModel {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(15))
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border()).
		Padding(0, 1)
	return &LogsModel{
		viewport: vp,
		entries:  make([]LogEntryMsg, 0),
		source:   source,
		workers:  append([]string(nil), workers...),
	}
}

func (m *LogsModel) Init() tea.Cmd {
	if m.started || m.source == nil || len(m.workers) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.events = make(chan LogEntryMsg, 128)
	m.started = true

	var streams sync.WaitGroup
	for _, worker := range m.workers {
		streams.Add(1)
		go func() {
			defer streams.Done()
			m.streamWorker(ctx, worker)
		}()
	}
	go func() {
		streams.Wait()
		close(m.events)
	}()

	return m.waitForLog()
}

func (m *LogsModel) streamWorker(ctx context.Context, worker string) {
	stream, err := m.source.ContainerLogs(ctx, worker, 100)
	if err != nil {
		select {
		case m.events <- LogEntryMsg{
			Worker:    worker,
			Line:      fmt.Sprintf("log stream unavailable: %v", err),
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
		}
		return
	}

	lines := make(chan string)
	go docker.NowStreamLogs(ctx, stream, lines)
	for line := range lines {
		select {
		case m.events <- LogEntryMsg{
			Worker:    worker,
			Line:      line,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}
	}
}

func (m *LogsModel) waitForLog() tea.Cmd {
	return func() tea.Msg {
		entry, ok := <-m.events
		if !ok {
			return logStreamsClosedMsg{}
		}
		return entry
	}
}

func (m *LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width - 4)
		m.viewport.SetHeight(msg.Height - 12)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == " " {
			m.paused = !m.paused
		}

	case LogEntryMsg:
		if !m.paused {
			m.entries = append(m.entries, msg)
			if len(m.entries) > 200 {
				m.entries = m.entries[len(m.entries)-200:]
			}
			m.renderView()
		}
		return m, m.waitForLog()

	case logStreamsClosedMsg:
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *LogsModel) renderView() {
	var b strings.Builder
	for _, e := range m.entries {
		clr := workerC[e.Worker]
		if clr == nil {
			clr = Muted()
		}
		tag := lipgloss.NewStyle().Foreground(clr).Bold(true).Render(e.Worker)
		timeTag := lipgloss.NewStyle().Foreground(Muted()).Width(8).Render(e.Timestamp.Format("15:04:05"))
		line := e.Line
		if len(line) > 60 {
			line = line[:57] + "..."
		}
		b.WriteString(fmt.Sprintf("%s  %-8s  %s\n", timeTag, tag, line))
	}
	if b.Len() == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(Muted()).Render("Waiting for logs..."))
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m *LogsModel) View() tea.View {
	title := lipgloss.NewStyle().Foreground(Accent()).Bold(true).Render("\U0001F4DC  Live Logs")

	status := lipgloss.NewStyle().Foreground(Low()).Bold(true).Render("\u25B6 STREAMING")
	if m.paused {
		status = lipgloss.NewStyle().Foreground(High()).Bold(true).Render("\u23F8 PAUSED")
	}

	legend := lipgloss.NewStyle().Foreground(Muted()).Render("[SPACE] toggle pause")
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", status, "  ", legend)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.viewport.View(),
	))
}

// Integration

func (m *AppModel) renderLogTab() string {
	if m.logModel == nil {
		m.logModel = NewLogsModel(m.docker, m.cfg.Containers.Names)
	}
	return m.logModel.View().Content
}
