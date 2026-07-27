package tui

import (
	"image/color"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Severity

type Severity int

const (
	SevCritical Severity = 5
	SevHigh     Severity = 4
	SevMedium   Severity = 3
	SevLow      Severity = 2
	SevInfo     Severity = 1
)

func (s Severity) String() string {
	switch s {
	case SevCritical:
		return "CRIT"
	case SevHigh:
		return "HIGH"
	case SevMedium:
		return "MED"
	case SevLow:
		return "LOW"
	case SevInfo:
		return "INFO"
	default:
		return "\u2014"
	}
}

func (s Severity) Color() color.Color {
	switch s {
	case SevCritical:
		return Critical()
	case SevHigh:
		return High()
	case SevMedium:
		return Medium()
	case SevLow:
		return Low()
	default:
		return Muted()
	}
}

// Finding

type Finding struct {
	Target      string
	Description string
	Severity    Severity
	Status      string
}

func findingsPlaceholder() []Finding {
	return []Finding{
		{"api.example.com", "Missing rate limiting on /login", SevCritical, "Open"},
		{"app.example.com", "JWT missing expiration validation", SevHigh, "Open"},
		{"staging.example.com", "Exposed .env with DB credentials", SevCritical, "Fixed"},
		{"cdn.example.com", "CORS misconfiguration", SevMedium, "Open"},
		{"admin.example.com", "Default admin credentials active", SevHigh, "Open"},
		{"mail.example.com", "SPF record missing", SevLow, "Open"},
		{"app.example.com", "Server header leaks version", SevInfo, "WontFix"},
		{"api.example.com", "GraphQL introspection enabled", SevMedium, "Open"},
		{"db.internal", "MongoDB no-auth on 27017", SevCritical, "Open"},
	}
}

// Recon Model

type ReconModel struct {
	table       table.Model
	findings    []Finding
	filterLevel Severity
	width       int
	height      int
}

func NewReconModel() *ReconModel {
	cols := []table.Column{
		{Title: "Target", Width: 18},
		{Title: "Finding", Width: 38},
		{Title: "Severity", Width: 10},
		{Title: "Status", Width: 10},
	}

	findings := findingsPlaceholder()
	rows := make([]table.Row, 0, len(findings))
	for _, f := range findings {
		rows = append(rows, table.Row{f.Target, f.Description, f.Severity.String(), f.Status})
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Border()).
		BorderBottom(true).
		Bold(true).
		Foreground(Accent())
	s.Selected = s.Selected.
		Foreground(Text()).
		Background(Surface()).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(Text())
	t.SetStyles(s)

	return &ReconModel{table: t, findings: findings}
}

func (m *ReconModel) Init() tea.Cmd { return nil }

func (m *ReconModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 12)
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.filterLevel = SevCritical
			m.applyFilter()
		case "2":
			m.filterLevel = SevHigh
			m.applyFilter()
		case "3":
			m.filterLevel = SevMedium
			m.applyFilter()
		case "4":
			m.filterLevel = SevLow
			m.applyFilter()
		case "5":
			m.filterLevel = SevInfo
			m.applyFilter()
		case "0":
			m.filterLevel = 0
			m.applyFilter()
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *ReconModel) applyFilter() {
	rows := make([]table.Row, 0, len(m.findings))
	for _, f := range m.findings {
		if m.filterLevel == 0 || f.Severity == m.filterLevel {
			rows = append(rows, table.Row{f.Target, f.Description, f.Severity.String(), f.Status})
		}
	}
	m.table.SetRows(rows)
}

func (m *ReconModel) View() tea.View {
	title := lipgloss.NewStyle().Foreground(Accent()).Bold(true).Render("\U0001F50D  Findings")
	filter := lipgloss.NewStyle().Foreground(Muted()).Render("[1]CRIT [2]HIGH [3]MED [4]LOW [5]INFO [0]ALL")
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", filter)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.table.View(),
	))
}

// Integration

func (m *AppModel) renderReconTab() string {
	if m.reconModel == nil {
		m.reconModel = NewReconModel()
	}
	return m.reconModel.View().Content
}
