# Feature Specification: Barbarossa CLI v1

**Feature Branch**: `main`
**Created**: 2026-07-24
**Status**: In Progress (M1-M4 and M6 done; interactive portion of M5 pending)

**Input**: A terminal UI for monitoring and controlling a 3-worker offensive security Docker cluster (charlie, oscar, papa).

## User Scenarios & Testing

### User Story 1 - Dashboard: Worker Status Overview (Priority: P1)

An operator opens the CLI to check if all three workers are online, their resource usage, and recent activity.

**Why this priority**: Core value — the primary reason to open the CLI.

**Independent Test**: Launch CLI, verify 3 worker cards render with status dots, CPU, RAM, uptime.

**Acceptance Scenarios**:
1. **Given** all workers online, **When** launch CLI, **Then** cards show colored status dot with ONLINE and metrics.
2. **Given** a worker unreachable, **When** launch CLI, **Then** card shows GRAY dot with OFFLINE and dashes.
3. **Given** any state, **When** 3s pass, **Then** cards refresh via tea.Tick.
4. **Given** user presses R, **When** any time, **Then** manual refresh triggers.

---

### User Story 2 - Multi-Tab Navigation (Priority: P2)

Operator switches between DASH, RECON, LOG tabs via keyboard.

**Why this priority**: Navigation enables all features.

**Independent Test**: Press TAB/Shift+TAB, verify correct tab renders.

**Acceptance Scenarios**:
1. **Given** DASH active, **When** TAB, **Then** RECON renders.
2. **Given** DASH active, **When** Shift+TAB, **Then** wraps to LOG.
3. **Given** any tab, **When** ?, **Then** help overlay with keybindings.

---

### User Story 3 - Recon Findings Table (Priority: P3)

Operator views findings in a navigable severity-color-coded table.

**Why this priority**: Core workflow for offensive security review.

**Acceptance Scenarios**:
1. **Given** findings exist, **When** RECON tab, **Then** table with Target, Finding, Severity, Status.
2. CRITICAL=red, HIGH=orange, MEDIUM=yellow, LOW=green, INFO=gray.
3. Arrow keys navigate rows; 1-5 filter by severity.

---

### User Story 4 - Live Log Streaming (Priority: P4)

Operator watches real-time color-coded logs from all workers.

**Acceptance Scenarios**:
1. Log entries stream color-coded by worker.
2. SPACE pauses/resumes; scroll works independently.

---

### User Story 5 - Docker Stats Integration (Priority: P5)

Real CPU/RAM from Docker API instead of mock data.

**Acceptance Scenarios**:
1. ContainerStats shows real values for charlie/oscar/papa.
2. OFFLINE within 3s if container stops.
3. Docker unreachable => all OFFLINE, no crash.

---

### User Story 6 - SSH Terminal (Priority: P6)

Embedded SSH to worker from TERM tab.

**Acceptance Scenarios**:
1. Interactive shell inside TUI viewport.
2. Connection failure shows error, no crash.

### Edge Cases

- **Terminal resize**: all components reflow via tea.WindowSizeMsg.
- **No Docker**: cards show OFFLINE with "Docker unreachable".
- **No containers**: cards show OFFLINE with "not found".
- **Rapid tab switch**: no flicker, atomic transitions.
- **Binary >10MB**: current known gap; approximately 13 MB.
- **Empty states**: "No recent activity", "No findings yet".
- **SSH timeout 5s**: show error, keep UI responsive.
- **Concurrent poll + render**: tea.Msg queue serializes.

## Requirements

### Functional Requirements

- FR-001: 3 worker cards with name, status dot, CPU%, RAM, uptime
- FR-002: Auto-poll every 3s via tea.Tick
- FR-003: 3 tabs (DASH/RECON/LOG) via TAB/Shift+TAB
- FR-004: Help overlay with keybindings on ?
- FR-005: Clean exit on Q/Ctrl+C restoring terminal
- FR-006: Activity feed: last 10 actions with worker name + timestamp
- FR-007: RECON: severity-color-coded findings table
- FR-008: LOG: color-coded streaming from 3 workers
- FR-009: Docker API for real container stats
- FR-010: SSH to workers for TERM tab
- FR-011: Graceful terminal resize
- FR-012: Graceful degradation when Docker/SSH unavailable
- FR-013: Binary <10MB
- FR-014: Memory <50MB RSS

### Key Entities

- **WorkerStatus**: name, online, cpu, ram, uptime
- **Activity**: worker, message, time
- **Finding**: target, description, severity, status, date
- **LogEntry**: worker, line, timestamp

## Success Criteria

- SC-001: Dashboard renders <500ms cold start
- SC-002: Cards update within 3s of container state change
- SC-003: Binary <10MB (currently approximately 13 MB; unmet)
- SC-004: Zero unhandled panics
- SC-005: All shortcuts documented in help overlay
- SC-006: Memory <50MB RSS after 10min idle

## Assumptions

- 256-color terminal support
- Docker accessible via socket or TCP
- SSH keys pre-configured at known path
- Containers named charlie, oscar, papa
- Go 1.25+ via GOTOOLCHAIN
- Terminal >=80x24
