package tui

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// ─── Adaptive palette ─────────────────────────────────────────

var lightDark func(color.Color, color.Color) color.Color

func SetAdaptive(darkBG bool) {
	lightDark = lipgloss.LightDark(darkBG)
}

func ld(light, dark color.Color) color.Color {
	if lightDark == nil { return dark }
	return lightDark(light, dark)
}

// Named colors — both light and dark variants
var (
	bgLight, bgDark             = lipgloss.Color("#FFFFFF"), lipgloss.Color("#0D1117")
	surfaceLight, surfaceDark   = lipgloss.Color("#F6F8FA"), lipgloss.Color("#161B22")
	borderLight, borderDark     = lipgloss.Color("#D0D7DE"), lipgloss.Color("#30363D")
	accentLight, accentDark     = lipgloss.Color("#CF3D1A"), lipgloss.Color("#F78166")
	accentDimL, accentDimD      = lipgloss.Color("#B5330F"), lipgloss.Color("#DA5B41")
	textLight, textDark         = lipgloss.Color("#24292F"), lipgloss.Color("#C9D1D9")
	mutedLight, mutedDark       = lipgloss.Color("#656D76"), lipgloss.Color("#8B949E")

	charlieL, charlieD = lipgloss.Color("#0550AE"), lipgloss.Color("#79C0FF")
	oscarL,   oscarD   = lipgloss.Color("#BF5B00"), lipgloss.Color("#FFA657")
	papaL,    papaD    = lipgloss.Color("#1A7F37"), lipgloss.Color("#7EE787")

	critL, critD = lipgloss.Color("#CF222E"), lipgloss.Color("#FF7B72")
	highL, highD = lipgloss.Color("#BF5B00"), lipgloss.Color("#FFA657")
	medL,  medD  = lipgloss.Color("#9A6700"), lipgloss.Color("#E3B341")
	lowL,  lowD  = lipgloss.Color("#1A7F37"), lipgloss.Color("#7EE787")
)

func Bg() color.Color       { return ld(bgLight, bgDark) }
func Surface() color.Color  { return ld(surfaceLight, surfaceDark) }
func Border() color.Color   { return ld(borderLight, borderDark) }
func Accent() color.Color   { return ld(accentLight, accentDark) }
func AccentDim() color.Color { return ld(accentDimL, accentDimD) }
func Text() color.Color     { return ld(textLight, textDark) }
func Muted() color.Color    { return ld(mutedLight, mutedDark) }
func Charlie() color.Color  { return ld(charlieL, charlieD) }
func Oscar() color.Color    { return ld(oscarL, oscarD) }
func Papa() color.Color     { return ld(papaL, papaD) }
func Critical() color.Color { return ld(critL, critD) }
func High() color.Color     { return ld(highL, highD) }
func Medium() color.Color   { return ld(medL, medD) }
func Low() color.Color      { return ld(lowL, lowD) }

// ─── Base styles ───────────────────────────────────────────────

var (
	AppStyle = func() lipgloss.Style { return lipgloss.NewStyle().Background(Bg()).Foreground(Text()) }

	Panel = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Background(Surface()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border()).
			Padding(1, 2)
	}

	// Tabs — connected borders
	tabActiveBorder = func() lipgloss.Border {
		b := lipgloss.RoundedBorder(); b.BottomLeft = "\u2518"; b.Bottom = " "; b.BottomRight = "\u2514"; return b
	}
	tabInactiveBorder = func() lipgloss.Border {
		b := lipgloss.RoundedBorder(); b.BottomLeft = "\u2534"; b.Bottom = "\u2500"; b.BottomRight = "\u2534"; return b
	}

	TabActive = func() lipgloss.Style {
		return lipgloss.NewStyle().Border(tabActiveBorder(), true).BorderForeground(Accent()).Foreground(Accent()).Bold(true).Padding(0, 2)
	}
	TabInactive = func() lipgloss.Style {
		return lipgloss.NewStyle().Border(tabInactiveBorder(), true).BorderForeground(Border()).Foreground(Muted()).Padding(0, 2)
	}

	WindowFrame = func() lipgloss.Style {
		return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(Accent()).Padding(0, 2).UnsetBorderTop()
	}
)

// ─── Typography ────────────────────────────────────────────────

func Bold() lipgloss.Style   { return lipgloss.NewStyle().Bold(true) }
func Sub() lipgloss.Style    { return lipgloss.NewStyle().Foreground(Muted()) }
func Mono() lipgloss.Style   { return lipgloss.NewStyle().Foreground(Accent()).Bold(true) }
func Success() lipgloss.Style { return lipgloss.NewStyle().Foreground(Low()).Bold(true) }
func Error() lipgloss.Style   { return lipgloss.NewStyle().Foreground(Critical()).Bold(true) }

// ─── Components ────────────────────────────────────────────────

func Dot(on bool, c color.Color) string {
	if on { return lipgloss.NewStyle().Foreground(c).Render("\u25CF") }
	return lipgloss.NewStyle().Foreground(Muted()).Render("\u25CB")
}

func Bar(ratio float64, w int, c color.Color) string {
	if w <= 0 { w = 1 }
	ratio = max(0, min(1, ratio))
	n := int(math.Round(ratio * float64(w)))
	return lipgloss.NewStyle().Foreground(c).Render(
		strings.Repeat("\u2588", n) + strings.Repeat("\u2592", w-n),
	)
}

func WorkerCard(name string, ws WorkerStatusMsg, clr color.Color, w int) string {
	if w < 26 { w = 26 }
	dot := Dot(ws.Online, clr)
	status := "ONLINE"
	st := lipgloss.NewStyle().Foreground(clr).Bold(true)
	if !ws.Online { status = "OFFLINE"; st = Sub() }

	hdr := lipgloss.NewStyle().Foreground(clr).Bold(true).Width(w-4).Align(lipgloss.Center).Render(strings.ToUpper(name))
	cr := min(1, max(0, pct(ws.CPU)))
	rr := min(1, max(0, ramPct(ws.RAM)))
	bw := w - 14; if bw < 8 { bw = 8 }

	body := lipgloss.JoinVertical(lipgloss.Left,
		st.Render(dot+" "+status),
		"",
		Sub().Width(3).Render("CPU")+" "+Bar(cr, bw, clr)+" "+ws.CPU,
		Sub().Width(3).Render("RAM")+" "+Bar(rr, bw, clr)+" "+ws.RAM,
		"",
		Sub().Render("\u23F1  "+ws.Uptime),
	)
	return Panel().BorderForeground(clr).Width(w).Render(lipgloss.JoinVertical(lipgloss.Center, hdr, body))
}

func pct(s string) float64 {
	v := 0.0
	for _, ch := range s { if ch>='0'&&ch<='9' { v=v*10+float64(ch-'0') }; if ch=='%'||ch=='M'{break} }
	return v/100
}
func ramPct(s string) float64 {
	v := 0.0
	for _, ch := range s { if ch>='0'&&ch<='9' { v=v*10+float64(ch-'0') }; if ch=='M'||ch=='%'{break} }
	return v/512
}

// ─── Messages ─────────────────────────────────────────────────

type WorkerStatusMsg struct { Name string; Online bool; CPU, RAM, Uptime string }
type ActivityMsg     struct { Worker, Message, Time string }
type TickMsg         struct{}
type DarkBgMsg       bool
