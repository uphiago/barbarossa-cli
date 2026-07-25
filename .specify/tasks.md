# Tasks: Barbarossa CLI v1

**Plan**: [plan.md](../plan.md) | **Spec**: [1-barbarossa-v1.md](../specs/1-barbarossa-v1.md)

## Phase 0: Foundation ✅

- [x] T0.1 — AppModel with tabs, polling, worker cards, activity feed, help
- [x] T0.2 — Centralized styles.go with palette, helpers, message types
- [x] T0.3 — main.go entrypoint compiles and runs

## Phase 1: Docker Integration (M3) — PRIORITY

- [x] T1.1 — Create `internal/docker/client.go` with Client interface and real + mock implementations
  - Interface: ListContainers, ContainerStats, ContainerLogs
  - Real: wraps github.com/docker/docker/client
  - Mock: returns predefined data for testing
- [x] T1.2 — Replace PollWorkers() mock with real Docker API calls
  - Read container names from env/config (default: charlie, oscar, papa)
  - Timeout 5s per call
  - Graceful degradation on Docker unreachable
- [x] T1.3 — Add config package: .barbarossa.yaml or env vars for DOCKER_HOST, container names

## Phase 2: Recon Findings Table (M4)

- [x] T2.1 — Create `internal/tui/recon.go` as tea.Model
  - Wrap bubbles/v2 table with columns: Target, Finding, Severity, Status
  - Severity color mapping in render function
  - Arrow key navigation (↑↓), ENTER to expand
  - Key filter 1-5 for severity levels
- [x] T2.2 — Implement ReconModel's Init/Update/View
  - Init: load mock findings array
  - Update: handle KeyMsg for navigation + filter
  - View: render table with app styling
- [x] T2.3 — Wire recon tab into AppModel.renderContent()

## Phase 3: Log Streaming (M6)

- [x] T3.1 — Create `internal/tui/logs.go` as tea.Model
  - Wrap bubbles/v2 viewport for scrollable content
  - 3 goroutines: one per worker, pushing log lines via tea.Msg channel
  - Color-coded by worker: charlie=cyan, oscar=orange, papa=green
  - SPACE: toggle pause/resume
- [x] T3.2 — Implement LogsModel's Init/Update/View
  - Init: start log goroutines
  - Update: receive LogEntryMsg, append to buffer, update viewport
  - View: render viewport with colorized lines
- [x] T3.3 — Wire log tab into AppModel.renderContent()

## Phase 4: SSH Terminal (M5)

- [x] T4.1 — Create `internal/ssh/client.go` with SSH client interface
  - Connect(host, user, key) → session
  - Run(cmd) → output
  - Interactive session via pty
- [ ] T4.2 — Create `internal/tui/terminal.go` as tea.Model
  - Worker selection popup (huh-style select)
  - Embedded pty terminal
  - Command history
  - Ctrl+T shortcut from any tab

## Phase 5: Polish & Edge Cases

- [ ] T5.1 — tea.WindowSizeMsg handler: all tabs reflow on terminal resize
  - Cards stay centered or switch to vertical on narrow terminals
  - Table/viewport resize correctly
- [ ] T5.2 — Loading states: spinner during Docker connect, SSH connect
- [ ] T5.3 — Error banners for Docker unreachable, SSH timeout, container not found
- [ ] T5.4 — Empty states: "No recent activity", "No findings yet", "No logs"
- [ ] T5.5 — Memory test: verify <50MB RSS after 10min idle
- [ ] T5.6 — Binary check: verify <10MB after each phase

## Implementation Order

```
T1.1 → T1.2 → T1.3   (Docker real)
T2.1 → T2.2 → T2.3   (Recon table)
T3.1 → T3.2 → T3.3   (Log streaming)
T4.1 → T4.2           (SSH)
T5.1 → T5.6           (Polish)
```
Each task compiles independently before moving to next. Commit after each task.
