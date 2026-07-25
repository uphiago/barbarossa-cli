# Barbarossa CLI — Design Document

## Visão

Terminal UI interativa para controlar e monitorar o cluster Barbarossa (Hermes + 3 workers Docker). Multi-painel, navegação por abas, atualização ao vivo.

## Stack

| Biblioteca | Versão | Uso |
|-----------|--------|-----|
| `charm.land/bubbletea/v2` | latest | Core TUI framework (Elm Architecture) |
| `charm.land/bubbles/v2` | latest | table, viewport, textinput, spinner, list, help, key |
| `charm.land/lipgloss/v2` | latest | Estilização completa |
| `charm.land/glamour/v2` | latest | Renderização Markdown |
| `charm.land/log/v2` | latest | Logging |
| `golang.org/x/crypto/ssh` | latest | Conexão SSH com workers |
| `github.com/docker/docker/client` | latest | Docker API |

## Arquitetura de Telas

```
┌─────────────────────────────────────────────────────┐
│ BARBAROSSA CLI                    [DASH] [RECON] [LOG] │ ← tabs
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ CHARLIE  │  │  OSCAR   │  │   PAPA   │          │ ← worker cards
│  │ ● ONLINE │  │ ● ONLINE │  │ ● ONLINE │          │
│  │ CPU 12%  │  │ CPU 3%   │  │ CPU 1%   │          │
│  │ RAM 256M │  │ RAM 512M │  │ RAM 128M │          │
│  │ UPTIME.. │  │ UPTIME.. │  │ UPTIME.. │          │
│  └──────────┘  └──────────┘  └──────────┘          │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │ Recent Activity                    12:34 UTC │   │ ← log stream
│  │ [charlie] subfinder -d target.com → 47 subs  │   │
│  │ [oscar]   compiled PoC → /tmp/exploit        │   │
│  │ [papa]    torsocks curl check → OK           │   │
│  │ [charlie] httpx probe → 23 live hosts        │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  ████████████░░░░░░  Scan Progress  67%            │ ← progress bar
│                                                     │
│  > _                                                 │ ← command bar
│                                                     │
│  TAB:switch  ↑↓:nav  ENTER:select  Q:quit  ?:help  │ ← help bar
└─────────────────────────────────────────────────────┘
```

## Telas (Tabs)

### 1. DASH — Dashboard
- **3 cards** com status dos workers (Charlie, Oscar, Papa)
  - Indicador online/offline colorido
  - CPU, RAM, uptime
  - Última atividade
- **Activity feed** — últimas 10 ações exibidas em viewport scrollável
- **Progress bar** — scan atual (se houver)

### 2. RECON — Findings
- **Tabela interativa** com colunas: Target, Finding, Severity, Status, Date
- Navegação com setas, ENTER abre detalhes
- Filtro por severity (teclas 1-5)
- Severidade colorida: CRIT=red, HIGH=orange, MED=yellow, LOW=green, INFO=gray

### 3. LOG — Live Logs
- **Viewport com tail -f** dos 3 workers simultâneo
- Cada worker com cor diferente (charlie=cyan, oscar=magenta, papa=yellow)
- Pause/resume streaming (SPACE)
- Scroll livre com mouse/teclado

### 4. TERM — Terminal (acesso rápido)
- Shell SSH embedado no worker selecionado
- Histórico de comandos
- Atalho: Ctrl+T abre popup de seleção de worker

## Paleta de Cores

```
Background: #0a0e14 (dark charcoal)
Surface:    #141b22
Border:     #1e2a33
Accent:     #ff6b35 (orange — cor "barbarossa")
Charlie:    #39bae6 (cyan)
Oscar:      #ff6b35 (orange)
Papa:       #c3f953 (green)
Text:       #bfbab0 (warm gray)
Muted:      #5c6773
CRITICAL:   #ff3333 (red)
HIGH:       #ff8c00 (orange)
MEDIUM:     #ffd700 (yellow)
LOW:        #4caf50 (green)
INFO:       #5c6773 (gray)
```

## Keybindings

| Tecla | Ação |
|-------|------|
| `TAB` / `Shift+TAB` | Navegar entre abas |
| `↑` `↓` | Navegar itens |
| `ENTER` | Selecionar/expandir |
| `Q` / `Ctrl+C` | Sair |
| `?` | Toggle help |
| `1-5` | Filtrar por severity (na tab RECON) |
| `SPACE` | Pause/resume logs |
| `R` | Refresh/atualizar |
| `Ctrl+T` | Abrir terminal SSH |

## Estrutura de Diretórios

```
barbarossa-cli/
├── main.go                 # Entrypoint
├── go.mod
├── DESIGN.md               # Este documento
├── internal/
│   ├── tui/
│   │   ├── app.go          # Model principal + tabs
│   │   ├── dashboard.go    # Tab: status workers
│   │   ├── recon.go        # Tab: findings table
│   │   ├── logs.go         # Tab: live logs
│   │   ├── terminal.go     # Tab: SSH shell
│   │   ├── help.go         # Help overlay
│   │   └── styles.go       # Cores, temas, estilos
│   ├── docker/
│   │   └── client.go       # Docker API wrapper
│   ├── ssh/
│   │   └── client.go       # SSH client wrapper
│   └── config/
│       └── config.go       # .env parsing
└── config/
    └── defaults.go         # Valores padrão
```

## Fluxo de Atualização

- Dashboard: poll Docker API a cada 3s via `tea.Tick`
- Activity: poll via `tea.Tick` a cada 5s
- Logs: streaming contínuo via goroutine + `tea.Cmd`
- Recon: carregamento sob demanda (tab switch)

## Milestones

1. **M1 — Fundação**: go.mod, main.go, app.go, dashboard.go (worker cards estáticos + estilos)
2. **M2 — Conexão**: Docker client real, SSH client, workers ao vivo
3. **M3 — Recon**: Tabela de findings com parsing do output
4. **M4 — Logs**: Viewport com streaming
5. **M5 — Terminal**: Shell SSH embedado
6. **M6 — Polimento**: Animações, transições, help overlay
