# barbarossa-cli Constitution

## Core Principles

### I. TUI-First, Keyboard-Centric
Every feature is designed for the terminal with keyboard-first navigation. Mouse support is additive, never the only path. Users must be able to reach every action within 3 keystrokes. The Bubble Tea Elm Architecture (Model-Update-View) is non-negotiable — all UI state flows through tea.Msg.

### II. Worker Independence
The CLI monitors but never depends on workers being online. Every worker card degrades gracefully to OFFLINE state. No blocking operations against remote hosts — SSH and Docker calls are async via tea.Cmd and must respect timeouts (max 5s connect, 30s operation).

### III. Composeable UI Components
Every tab, view, and widget implements tea.Model and lives in its own file under internal/tui/. Styles are centralized in styles.go — no hardcoded colors, borders, or padding anywhere else. Components are independently testable with tea.testing or by asserting View() output.

### IV. Zero-Bloat Dependencies
Only the Charm ecosystem is core (Bubble Tea v2, Lip Gloss v2, Bubbles v2). Every additional dependency (Docker client, SSH, etc.) must be behind an interface in internal/ so it can be mocked in tests and replaced without touching UI code. No ORMs, no frameworks beyond Charm.

### V. Real-Time, Not Poll-Heavy
Dashboard polls Docker API every 3s. Logs stream via goroutines. RECON loads on tab switch. All data flows through tea.Msg — no goroutines touching model state directly. Use tea.Tick, tea.Batch, and tea.Sequence for timing.

## Security Requirements

- No credentials in source code, config files, or environment variables that touch git
- SSH keys read from filesystem only, never embedded
- Docker socket path configurable, never hardcoded to /var/run/docker.sock
- API tokens passed via environment or auth helpers, never in CLI args that show in ps

## Performance Standards

- Cold start (first View render) must complete within 500ms with no workers connected
- Worker polling must not block UI rendering — stale data is acceptable within the 3s tick window
- Memory usage must stay under 50MB RSS under normal operation
- Binary size target: under 10MB (currently 5.1MB ✅)

## Development Workflow

- Conventional commits: feat:, fix:, docs:, chore:, refactor:, test:, style:
- Commit as hfelipe <hfelipe.sh@gmail.com>
- No bot attribution or session trailers in commits
- go vet ./... passes before every commit
- Code comments document non-obvious constraints only — no changelog narration

## Governance

This constitution supersedes all other development practices. Amendments require updating this file, committing with reason, and pushing. All code changes must align with the principles above. The AGENTS.md serves as runtime guidance and is subordinate to this constitution.

**Version**: 1.0.0 | **Ratified**: 2026-07-24 | **Last Amended**: 2026-07-24
