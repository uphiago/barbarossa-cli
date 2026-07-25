# Lip Gloss v2 — Critical API Changes & Migration Cheat Sheet

> Source: [UPGRADE_GUIDE_V2.md](https://raw.githubusercontent.com/charmbracelet/lipgloss/main/UPGRADE_GUIDE_V2.md)
> Context: barbarossa-cli already uses v2 (`charm.land/lipgloss/v2`) + Bubble Tea v2 (`charm.land/bubbletea/v2`).
> This sheet focuses on patterns relevant to our codebase and gotchas for future development.

---

## 1. Module Path

| v1 | v2 |
|---|---|
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/lipgloss/table` | `charm.land/lipgloss/v2/table` |
| `github.com/charmbracelet/lipgloss/tree` | `charm.land/lipgloss/v2/tree` |
| `github.com/charmbracelet/lipgloss/list` | `charm.land/lipgloss/v2/list` |

**barbarossa-cli status: ✅ Already on v2.** No action needed.

---

## 2. Color System — Biggest Change

### `Color` is now a *function*, not a *type*

```go
// v1 — Color was a string type
var c lipgloss.Color = "#ff00ff"   // type-level

// v2 — Color is a constructor returning image/color.Color
var c color.Color = lipgloss.Color("#ff00ff")  // function call
```

**Return type is `image/color.Color`** (stdlib), NOT `lipgloss.TerminalColor`.

### barbarossa-cli styles.go — already correct

```go
// ✅ v2 pattern (our code)
import "image/color"
import "charm.land/lipgloss/v2"

var Accent = lipgloss.Color("#ff6b35")         // returns color.Color
func RenderBox(title, content string, clr color.Color) string { ... }   // ✅ uses color.Color
```

### 🚫 `TerminalColor` interface — REMOVED

All style setters now accept `image/color.Color`:

```go
// v2 — all these take color.Color
func (s Style) Foreground(c color.Color) Style
func (s Style) Background(c color.Color) Style
func (s Style) BorderForeground(c ...color.Color) Style
func (s Style) UnderlineColor(c color.Color) Style
```

**Our code is already correct** — we use `color.Color` in function signatures.

### `ANSIColor` — now an alias

```go
// v1: type ANSIColor uint
// v2: type ANSIColor = ansi.IndexedColor

// v2 also exports named ANSI constants:
lipgloss.Black, lipgloss.Red, lipgloss.Green, lipgloss.Yellow,
lipgloss.Blue, lipgloss.Magenta, lipgloss.Cyan, lipgloss.White,
lipgloss.BrightBlack, lipgloss.BrightRed, lipgloss.BrightGreen,
lipgloss.BrightYellow, lipgloss.BrightBlue, lipgloss.BrightMagenta,
lipgloss.BrightCyan, lipgloss.BrightWhite
```

---

## 3. `AdaptiveColor`, `CompleteColor` — MOVED

These types are **no longer in the root package**.

### Quick migration path (compat package)

```go
import "charm.land/lipgloss/v2/compat"

color := compat.AdaptiveColor{
    Light: lipgloss.Color("#f1f1f1"),  // ⚠️ values are color.Color, not strings!
    Dark:  lipgloss.Color("#cccccc"),
}
```

### Recommended path (pure v2)

```go
// 1. Detect background
hasDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

// 2. Create a color picker
lightDark := lipgloss.LightDark(hasDark)

// 3. Pick colors at point of use
fg := lightDark(lipgloss.Color("#333333"), lipgloss.Color("#f1f1f1"))
s := lipgloss.NewStyle().Foreground(fg)
```

### With Bubble Tea v2 (our stack)

```go
type model struct {
    styles styles
}

func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor   // triggers async detection
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        m.styles = newStyles(msg.IsDark())   // rebuild styles on response
    }
    return m, nil
}

func newStyles(bgIsDark bool) styles {
    lightDark := lipgloss.LightDark(bgIsDark)
    return styles{
        title: lipgloss.NewStyle().Foreground(
            lightDark(lipgloss.Color("#333333"), lipgloss.Color("#f1f1f1")),
        ),
    }
}
```

**Gotcha:** `HasDarkBackground()` now REQUIRES explicit `os.Stdin, os.Stdout` args — no-arg version is gone.

**barbarossa-cli status: Not yet using adaptive colors.** When we add them, prefer the `LightDark` pattern over `compat`.

---

## 4. Renderer — COMPLETELY REMOVED

| v1 (gone) | v2 replacement |
|---|---|
| `lipgloss.NewRenderer(w, opts...)` | Not needed |
| `lipgloss.DefaultRenderer()` | Not needed |
| `lipgloss.SetDefaultRenderer(r)` | Not needed |
| `renderer.NewStyle()` | `lipgloss.NewStyle()` |
| `lipgloss.ColorProfile()` | `colorprofile.Detect(w, env)` |
| `lipgloss.SetColorProfile(p)` | Set `lipgloss.Writer.Profile` |

**`Style` is now a plain value type** — no renderer pointer, no global state.

**barbarossa-cli status: ✅ Already clean.** No renderer usage.

---

## 5. Printing & Color Downsampling — CRITICAL GOTCHA

This is the most likely source of bugs in migrated code.

### What changed

In v1, `Style.Render()` downsampled colors internally via the renderer.
In v2, **`Render()` always emits full-fidelity ANSI.** Downsampling happens when you print.

### If printing to stdout (standalone CLI)

```go
s := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render("hi")

// ❌ v1 old way — no downsampling
fmt.Println(s)   // terminal might show garbage for unsupported colors

// ✅ v2 — automatic downsampling
lipgloss.Println(s)
lipgloss.Fprintln(os.Stderr, s)
lipgloss.Sprint(s)   // returns string, downsampled for stdout
```

### If using Bubble Tea (our case)

**No changes needed** — Bubble Tea v2 handles downsampling internally.
`lipgloss.Println` is only needed for standalone CLI tools, not Bubble Tea programs.

**barbarossa-cli status: Bubble Tea v2 handles output.** No action needed on printing.

---

## 6. Style API — New Methods & Changes

### New methods in v2

| Method | Description | Signature |
|---|---|---|
| `UnderlineStyle(Underline)` | Fine-grained underline style | `UnderlineCurly`, `UnderlineDouble`, `UnderlineSingle`, etc. |
| `UnderlineColor(color.Color)` | Colored underlines | Takes `color.Color` |
| `PaddingChar(rune)` | Char used for padding fill | Single rune |
| `MarginChar(rune)` | Char used for margin fill | Single rune |
| `Hyperlink(link, params...)` | Clickable hyperlinks | URI + optional params |
| `BorderForegroundBlend(...color.Color)` | Gradient border colors | Variadic colors |
| `BorderForegroundBlendOffset(int)` | Gradient offset | Integer offset |

Each new setter has a `Get*` and `Unset*` counterpart.

### Whitespace options — consolidated

```go
// v1 — separate foreground/background (REMOVED)
lipgloss.WithWhitespaceForeground(lipgloss.Color("#333"))
lipgloss.WithWhitespaceBackground(lipgloss.Color("#000"))

// v2 — single style option
lipgloss.WithWhitespaceStyle(
    lipgloss.NewStyle().
        Foreground(lipgloss.Color("#333")).
        Background(lipgloss.Color("#000")),
)
```

### Color getters return `color.Color`

```go
fg := s.GetForeground()   // v2: returns color.Color (was TerminalColor)
```

---

## 7. Removed APIs — Complete Reference

| Removed | Replacement |
|---|---|
| `type Renderer` | — (removed entirely) |
| `DefaultRenderer()` | — |
| `SetDefaultRenderer(r)` | — |
| `NewRenderer(w, opts...)` | — |
| `ColorProfile()` | `colorprofile.Detect(w, env)` |
| `SetColorProfile(p)` | Set `lipgloss.Writer.Profile` |
| `HasDarkBackground()` (no args) | `HasDarkBackground(in, out)` |
| `SetHasDarkBackground(b)` | Pass bool to `LightDark` |
| `type TerminalColor` | `image/color.Color` |
| `type Color string` | `func Color(string) color.Color` |
| `type ANSIColor uint` | `type ANSIColor = ansi.IndexedColor` |
| `type AdaptiveColor{Light,Dark string}` | `compat.AdaptiveColor{Light,Dark color.Color}` or `LightDark` |
| `type CompleteColor{TrueColor,ANSI256,ANSI string}` | `compat.CompleteColor` or `lipgloss.Complete(profile)` |
| `type CompleteAdaptiveColor` | `compat.CompleteAdaptiveColor` |
| `WithWhitespaceForeground(c)` | `WithWhitespaceStyle(s)` |
| `WithWhitespaceBackground(c)` | `WithWhitespaceStyle(s)` |
| `renderer.NewStyle()` | `lipgloss.NewStyle()` |

---

## 8. barbarossa-cli Specific Audit

### ✅ Already correct
- Module path: `charm.land/lipgloss/v2`
- `lipgloss.Color()` as function call (not type assertion)
- `color.Color` in function signatures
- `lipgloss.NewStyle()` without renderer
- Bubble Tea v2 handles output downsampling
- No `AdaptiveColor`, `TerminalColor`, or `Renderer` usage

### 🔍 Watch for future changes
- **Adaptive colors:** Use `lipgloss.LightDark(hasDark)(light, dark)` pattern, NOT `compat.AdaptiveColor`
- **Dark background detection:** Must pass `os.Stdin, os.Stdout` explicitly
- **Printing helpers:** If adding non-BubbleTea CLI output, use `lipgloss.Println()` not `fmt.Println()`
- **Whitespace:** Consolidate to single `WithWhitespaceStyle()` call
- **Underline:** New `UnderlineStyle()` and `UnderlineColor()` available for rich formatting
- **Hyperlinks:** New `Hyperlink()` method on Style
- **Gradient borders:** `BorderForegroundBlend()` and `BorderForegroundBlendOffset()` for fancy borders

### ⚠️ Common gotchas when adding new features

1. **Adaptive colors via compat require `color.Color` values, not strings**
   ```go
   // WRONG — v1 holdover
   compat.AdaptiveColor{Light: "#fff", Dark: "#000"}
   // RIGHT
   compat.AdaptiveColor{Light: lipgloss.Color("#fff"), Dark: lipgloss.Color("#000")}
   ```

2. **`HasDarkBackground` now requires both std in and out**
   ```go
   // WRONG
   lipgloss.HasDarkBackground()
   // RIGHT
   lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
   ```

3. **`Render()` output loses downsampling protection**
   ```go
   // If you ever Render() a string and print it outside Bubble Tea:
   str := s.Render("hello")
   lipgloss.Println(str)   // ✅ downsamples
   fmt.Println(str)         // ❌ raw ANSI, may garble output
   ```

4. **NewStyle() is fully stateless now** — no global renderer means `NewStyle()` is trivial and safe to call in loops, but renderer-dependent features (color profile, dark detection) must be handled at the output layer.

---

## 9. Quick Reference — Most Common Patterns

| What | v2 Code |
|---|---|
| Hex color | `lipgloss.Color("#ff00ff")` |
| ANSI color | `lipgloss.Color("5")` or `lipgloss.Magenta` |
| Create style | `lipgloss.NewStyle().Foreground(c).Bold(true)` |
| Set foreground | `.Foreground(lipgloss.Color("#hex"))` |
| Adaptive (BubbleTea) | `lightDark := lipgloss.LightDark(isDark)` → `lightDark(light, dark)` |
| Adaptive (standalone) | `hasDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` |
| Print with downsampling | `lipgloss.Println(s.Render("hi"))` |
| Complete color | `lipgloss.Complete(profile)(ansi, ansi256, trueColor)` |
| Whitespace | `lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(c))` |
| Underline style | `.UnderlineStyle(lipgloss.UnderlineCurly)` |
| Underline color | `.UnderlineColor(lipgloss.Color("#FF0000"))` |
| Hyperlink | `.Hyperlink("https://example.com")` |
| Border gradient | `.BorderForegroundBlend(c1, c2, c3)` |
| Padding char | `.PaddingChar('\u00b7')` |
| Margin char | `.MarginChar(' ')` |
