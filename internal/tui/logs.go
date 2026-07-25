package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
}

func NewLogsModel() *LogsModel {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(15))
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorderClr).
		BorderForeground(BorderClr).
		Padding(0, 1)
	return &LogsModel{viewport: vp, entries: make([]LogEntryMsg, 0)}
}

func (m *LogsModel) Init() tea.Cmd {
	return m.simulateStream()
}

func (m *LogsModel) simulateStream() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		workers := []string{"charlie", "oscar", "papa"}
		msgs := []string{
			"nmap scan complete \u2014 47 ports open",
			"subfinder: 12 new subdomains discovered",
			"nuclei: CVE-2024-1234 matched",
			"httpx probe: 23 live hosts",
			"ffuf: /admin /api /backup found",
			"torsocks: circuit established",
			"masscan sweep: /24 done in 2.3s",
			"gdb: breakpoint hit @ 0x401234",
			"strace: openat(/etc/passwd) = 3",
			"health check: all services OK",
		}
		return LogEntryMsg{
			Worker:    workers[t.UnixNano()%3],
			Line:      msgs[t.UnixNano()%int64(len(msgs))],
			Timestamp: t,
		}
	})
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
		return m, m.simulateStream()
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
			clr = MutedClr
		}
		tag := lipgloss.NewStyle().Foreground(clr).Bold(true).Render(e.Worker)
		timeTag := lipgloss.NewStyle().Foreground(MutedClr).Width(8).Render(e.Timestamp.Format("15:04:05"))
		line := e.Line
		if len(line) > 60 {
			line = line[:57] + "..."
		}
		b.WriteString(fmt.Sprintf("%s  %-8s  %s\n", timeTag, tag, line))
	}
	if b.Len() == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(MutedClr).Render("Waiting for logs..."))
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m *LogsModel) View() tea.View {
	title := lipgloss.NewStyle().Foreground(Accent).Bold(true).Render("\U0001F4DC  Live Logs")

	status := lipgloss.NewStyle().Foreground(Low).Bold(true).Render("\u25B6 STREAMING")
	if m.paused {
		status = lipgloss.NewStyle().Foreground(High).Bold(true).Render("\u23F8 PAUSED")
	}

	legend := lipgloss.NewStyle().Foreground(MutedClr).Render("[SPACE] toggle pause")
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
		m.logModel = NewLogsModel()
	}
	return m.logModel.View().Content
}
