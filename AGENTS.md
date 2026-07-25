# barbarossa-cli — Agent Context

Terminal UI interativa para controlar o cluster Barbarossa (Hermes + charlie/oscar/papa). Construída com o ecossistema Charmbracelet.


## Build & Install Status

```
go version:   1.25.0 (auto-downloaded via GOTOOLCHAIN)
go build:     OK - 5.1MB binary
go vet:       OK - no issues
go install:   OK - installs to $GOPATH/bin/barbarossa-cli
go mod tidy:  OK - 40 deps in go.sum
binary:       /usr/local/bin/barbarossa-cli
```

### Como compilar e rodar

```bash
cd barbarossa-cli
go mod tidy
go build -o barbarossa-cli .
./barbarossa-cli
```

**Nota:** Go 1.25 e baixado automaticamente via GOTOOLCHAIN - nao precisa instalar manualmente.


## Repositório

```
github.com/uphiago/barbarossa-cli
```

## Stack Definida (Go 1.25)

| Lib | Caminho import | O que faz |
|-----|---------------|-----------|
| Bubble Tea | `charm.land/bubbletea/v2` | Core TUI — Elm Architecture |
| Bubbles | `charm.land/bubbles/v2` | table, viewport, textinput, spinner, list, help, key, paginator, progress, timer, textarea, filepicker |
| Lip Gloss | `charm.land/lipgloss/v2` | Estilização: cores, bordas, padding, margins, alignment, compositing, join, wrap, tables, lists, trees |
| Glamour | — | Renderização Markdown (reports) |
| Log | — | Logger |
| Harmonica | — | Animações físicas (opcional) |
| `golang.org/x/crypto/ssh` | SSH client nativo | Conexão com workers |
| `github.com/docker/docker/client` | Docker API | Status containers, logs |

## Docs de referência do Charm

- Site: https://charm.land/
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Bubbles components: https://github.com/charmbracelet/bubbles
- Lip Gloss API: https://github.com/charmbracelet/lipgloss
- Exemplos oficiais: https://github.com/charmbracelet/bubbletea/tree/main/examples
- Community bubbles: https://github.com/charm-and-friends/additional-bubbles

## Arquivos existentes

```
barbarossa-cli/
├── main.go                     # Entrypoint, cria App + tea.NewProgram + alt screen
├── go.mod                      # Módulo Go (deps resolvidas)
├── DESIGN.md                   # Design completo: telas, cores, keybindings, milestones
└── internal/
    └── tui/
        ├── styles.go           # Paleta de cores, estilos base, mensagens, helpers
        └── app.go              # Model principal (AppModel) + tabs + dashboard view
```

## O que já está pronto

### `internal/tui/styles.go`
- Paleta de cores (Bg, Surface, Border, Accent, Charlie, Oscar, Papa, severidades)
- Estilos base: AppStyle, TitleStyle, BorderStyle, TabStyle, ActiveTabStyle, ErrorStyle, HelpStyle
- Helpers: RenderBox, StatusDot, Truncate (parametrizados com `color.Color`)
- Tipos de mensagem: WorkerStatusMsg, ActivityMsg, TickMsg
- Comando de polling: PollWorkers() (mock — workers fixos)

### `internal/tui/app.go`
- AppModel com tabs ["DASH", "RECON", "LOG"], activeTab, worker state, activity log
- tea.Model interface completa: Init, Update, View
- Keybindings: Tab/Shift+Tab navega abas, Q sai, ? help, R refresh
- Polling automático a cada 3s com tea.Tick
- Dashboard view (tab 0):
  - Header "BARBAROSSA CLI" + clock
  - 3 worker cards lado a lado com status dot, CPU, RAM, uptime
  - Activity feed (últimas 10 ações)
  - Command bar na base
- Help overlay
- Placeholder views para RECON e LOG tabs

### `main.go`
- Inicializa App via `tui.NewApp()`
- Roda com alt screen (via View.AltScreen = true no bubbletea v2)
- Clean exit

## O que falta implementar (ordem)

### M1 — App com Tabs funcional
- [x] `internal/tui/app.go` — Model principal com:
  - `tabs []string` = {"DASH", "RECON", "LOG"}
  - `activeTab int`
  - `workers map[string]WorkerStatusMsg`
  - `tea.Model` interface (Init, Update, View)
  - Update: handle KeyMsg (tab/Q/?) e mensagens de polling

### M2 — Dashboard (tab 0)
- [x] Dashboard embutido em `app.go` (método `renderDashboard()`) — View que renderiza:
  - Header "BARBAROSSA CLI" estilizado
  - 3 worker cards lado a lado (charlie, oscar, papa)
  - Cada card: nome, status dot, CPU, RAM, uptime, border colorida por worker
  - Activity feed na parte inferior (últimas ações)
  - Command bar (textinput) na base

### M3 — Docker real
- [x] `internal/docker/client.go` — wrapper da API Docker:
  - `ListContainers()` → container status
  - `ContainerStats(name)` → CPU/RAM
  - `ContainerLogs(name)` → streaming

### M4 — Recon tab
- [x] `internal/tui/recon.go` — tabela interativa:
  - Colunas: Target, Finding, Severity, Status
  - Severity colorida
  - Navegação ↑↓, filtro 1-5

### M5 — SSH
- [ ] `internal/ssh/client.go` — SSH wrapper

### M6 — Logs tab
- [ ] `internal/tui/logs.go` — viewport com tail -f

## Paleta de Cores (definida em styles.go)

```
Bg:       #0a0e14 (dark charcoal)
Surface:  #141b22
Border:   #1e2a33
Accent:   #ff6b35 (barbarossa orange)
Charlie:  #39bae6 (cyan)
Oscar:    #ff6b35 (orange)
Papa:     #c3f953 (green)
Text:     #bfbab0 (warm gray)
Muted:    #5c6773
```

## O Projeto Barbarossa (contexto)

3 workers Docker no `docker-compose.yml`:
- **charlie** (collect) — recon: nmap, subfinder, httpx, nuclei, ffuf, etc
- **oscar** (operate) — exploit dev: gdb, gcc, strace, + tools do charlie
- **papa** (persist) — anonymous via Tor SOCKS5:9050

O CLI se conecta via Docker API (socket ou TCP) e SSH.

## Como testar

```bash
cd barbarossa-cli
go mod tidy        # resolver deps
go run .           # rodar TUI
```

## Convenções

- Todo componente TUI implementa `tea.Model` (Init, Update, View)
- Estilos reutilizáveis em `styles.go`
- Mensagens são tipos exportados em `styles.go`
- Cada tela/aba é um arquivo separado em `internal/tui/`
- Cores usam a paleta definida, nunca valores hardcoded
