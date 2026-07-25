# Implementation Plan: Barbarossa CLI v1

**Branch**: `main` | **Date**: 2026-07-24 | **Spec**: [1-barbarossa-v1.md](../specs/1-barbarossa-v1.md)

**Input**: Feature specification for the Barbarossa CLI — a Bubble Tea v2 TUI for monitoring a 3-worker Docker security cluster.

## Summary

Build a keyboard-navigable terminal UI with 3 tabs (DASH, RECON, LOG) using Charm ecosystem v2 libraries. The app polls Docker for worker stats every 3s, renders worker cards, activity feed, findings table, and log streams. All state flows through tea.Msg; UI is pure View() functions referencing centralized styles.

## Technical Context

**Language/Version**: Go 1.25.0 (auto-downloaded via GOTOOLCHAIN)
**Primary Dependencies**: Bubble Tea v2, Lip Gloss v2, Bubbles v2 (table, viewport, textinput)
**Storage**: N/A (stateless UI; findings/data sourced from worker SSH commands)
**Testing**: go test ./... with tea.testing for TUI component assertions
**Target Platform**: Linux/macOS terminal (256-color), 80x24 minimum
**Project Type**: CLI / TUI application (single binary)
**Performance Goals**: <500ms cold start, <50MB RSS, <10MB binary
**Constraints**: Zero blocking I/O; all Docker/SSH operations async via tea.Cmd
**Scale/Scope**: 3 workers, 3 tabs, ~2000 LOC total

## Constitution Check

| Principle | Check |
|-----------|-------|
| I. TUI-First, Keyboard-Centric | ✅ TAB/Shift+TAB focus switching, ? help, Q quit, all via tea.KeyMsg |
| II. Worker Independence | ✅ PollWorkers() mock data; cards degrade to OFFLINE; no blocking calls |
| III. Composeable UI Components | ✅ styles.go central styles; app.go/recon.go/logs.go separate files |
| IV. Zero-Bloat Dependencies | ✅ Only charm ecosystem + x/crypto/ssh + docker/client behind interfaces |
| V. Real-Time, Not Poll-Heavy | ✅ 3s Tick polling; tea.Batch for parallel commands; goroutines for logs |

## Project Structure

### Documentation (this feature)

```text
.specify/
├── memory/constitution.md      # Governing principles
├── specs/1-barbarossa-v1.md    # Feature specification
├── plan.md                     # This file
└── tasks.md                    # Generated tasks (next step)
```

### Source Code (repository root)

```text
barbarossa-cli/
├── main.go                     # Entrypoint: tea.NewProgram(model)
├── go.mod                      # Module + dependencies
├── .gitignore
├── internal/
│   ├── tui/
│   │   ├── styles.go           # Palette, base styles, helpers, message types
│   │   ├── app.go              # AppModel: tabs, workers, activities, help
│   │   ├── recon.go            # RECON tab: findings table (bubbles/table)
│   │   ├── logs.go             # LOG tab: streaming viewport (bubbles/viewport)
│   │   └── terminal.go         # TERM tab: SSH shell (future)
│   ├── docker/
│   │   └── client.go           # Docker API wrapper (real stats)
│   └── ssh/
│       └── client.go           # SSH client wrapper
```

**Structure Decision**: Flat internal/ layout with tui/, docker/, ssh/ subdirectories. Each tab is one file implementing tea.Model for independent testing. Docker and SSH behind interfaces for mocking.

## Implementation Phases

### Phase 0: Verify Foundation ✅ (done)
- [x] app.go with AppModel, tabs, polling, worker cards, activity feed, help
- [x] styles.go with palette, base styles, helpers, message types
- [x] main.go entrypoint
- [x] go mod tidy, go build passes, go vet clean, 5.1MB binary

### Phase 1: Real Docker Integration (M3)
- [ ] `internal/docker/client.go` — Docker client interface:
  - `type Client interface { ListContainers() (...); ContainerStats(name string) (...); ContainerLogs(name string) (... io.ReadCloser) }`
  - Real implementation using `github.com/docker/docker/client`
  - Mock implementation for testing
- [ ] Replace PollWorkers() mock with real Docker API calls
- [ ] Container name config via environment or flag (default: charlie, oscar, papa)

### Phase 2: Recon Findings Table (M4)
- [ ] `internal/tui/recon.go` — findings tab:
  - Uses `charm.land/bubbles/v2/table` for interactive table
  - Columns: Target, Finding, Severity, Status
  - Severity color-coded (CRITICAL=red, HIGH=orange, MEDIUM=yellow, LOW=green)
  - Arrow key navigation, ENTER for details
  - Filter keys 1-5
  - Mock data for testing; later sourced from worker scans

### Phase 3: Log Streaming (M6)
- [ ] `internal/tui/logs.go` — log tab:
  - Uses `charm.land/bubbles/v2/viewport` for scrollable content
  - 3 goroutines reading from Docker container logs
  - Color-coded by worker (charlie=cyan, oscar=orange, papa=green)
  - SPACE pause/resume
  - Scroll tracking (auto-scroll to bottom when at bottom)

### Phase 4: SSH Terminal (M5)
- [ ] `internal/ssh/client.go` — SSH client wrapper
- [ ] `internal/tui/terminal.go` — embedded terminal tab
  - Worker selection popup
  - Interactive shell via pty
  - Command history

### Phase 5: Polish
- [ ] tea.WindowSizeMsg handler for responsive layout on resize
- [ ] Loading states (spinners during Docker connect, SSH connect)
- [ ] Error banners (Docker unreachable, SSH timeout)
- [ ] Config file (.barbarossa.yaml) for Docker socket, SSH key paths, worker names

## Complexity Tracking

No constitution violations. All dependencies are behind interfaces as required by Principle IV.
