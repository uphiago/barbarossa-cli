# 🔱 barbarossa-cli

> Terminal UI para controlar o cluster Barbarossa — 3 workers de segurança ofensiva.

Barbarossa CLI é uma interface TUI (Terminal User Interface) construída com [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2 para gerenciar e monitorar o cluster Barbarossa: três workers Docker especializados em recon, engenharia reversa e operações anônimas.

---

## 🏗️ Arquitetura

```
┌────────────┐     ┌────────────┐     ┌────────────┐
│  CHARLIE   │     │   OSCAR    │     │    PAPA    │
│  (Collect) │     │ (Operate)  │     │ (Persist)  │
├────────────┤     ├────────────┤     ├────────────┤
│ nmap       │     │ gdb        │     │ Tor :9050  │
│ nuclei     │     │ gcc        │     │ python3    │
│ subfinder  │     │ strace     │     │ nmap       │
│ httpx      │     │ ltrace     │     │            │
│ ffuf       │     │ nmap       │     │            │
│ masscan    │     │ nuclei     │     │            │
│ python3    │     │ python3    │     │            │
└────────────┘     └────────────┘     └────────────┘
     ✅                 ✅                 ✅
```

| Worker | Hostname | Função | Ferramentas Principais |
|--------|----------|--------|----------------------|
| Charlie | charlie | Coleta &amp; Recon | nmap, nuclei, subfinder, httpx, ffuf, masscan, amass |
| Oscar | oscar | Engenharia Reversa | gdb, gcc, strace, ltrace, xxd + ferramentas do Charlie |
| Papa | papa | Operações Anônimas | Tor SOCKS5:9050, torsocks, nmap, python3 |

---

## Funcionalidades

- **Dashboard (DASH)** — 3 cards de status dos workers (online/offline, CPU, RAM, uptime) com polling automático via Docker API
- **Activity feed** — Últimas ações com timestamp
- **Recon (RECON)** — Tabela interativa de findings com filtro por severidade (1-5), navegação por setas, cores por criticidade
- **Log Streaming (LOG)** — Viewport scrollável com logs ao vivo dos 3 workers, coloridos por worker, pause/resume com SPACE
- **Navegação por abas:** TAB / Shift+TAB
- **Help overlay** interativo com ?
- **Configuração** via `.barbarossa.yaml` ou variáveis de ambiente
- **Degradação graciosa** quando Docker está indisponível
- Tema escuro com paleta de cores exclusiva

### Roadmap

| Marco | Status | Descrição |
|-------|--------|-----------|
| M1 - Tabs | ✅ | App com 3 abas, navegação, polling |
| M2 - Dashboard | ✅ | Worker cards, activity feed, command bar |
| M3 - Docker real | ✅ | Wrapper API Docker para stats ao vivo |
| M4 - Recon | ✅ | Tabela interativa de findings |
| M5 - SSH | 🚧 | Conexão SSH embedada (interface pronta) |
| M6 - Logs | ✅ | Viewport com streaming de logs dos workers |

---

## Configuração

Arquivo `.barbarossa.yaml` (no diretório atual ou `~/.barbarossa.yaml`):

```yaml
docker:
  host: unix:///var/run/docker.sock
containers:
  names:
    - charlie
    - oscar
    - papa
```

Variáveis de ambiente sobrescrevem o arquivo:

| Variável | Descrição |
|----------|-----------|
| `DOCKER_HOST` | Endereço do daemon Docker |
| `BARBAROSSA_CONTAINERS` | Lista separada por vírgula de nomes de containers |
| `BARBAROSSA_CONFIG` | Caminho completo para o arquivo de configuração |

---

## Como usar

```bash
git clone https://github.com/uphiago/barbarossa-cli.git
cd barbarossa-cli
go mod tidy
go run .
```

### Controles

| Tecla | Ação |
|-------|------|
| TAB / Shift+TAB | Navegar entre abas |
| ↑ ↓ | Navegar itens |
| ENTER | Selecionar / expandir |
| R | Refresh manual |
| ? | Toggle help |
| Q / Ctrl+C | Sair |
| 1-5 | Filtrar por severidade (RECON) |
| SPACE | Pausar/retomar logs (LOG) |

---

## Stack

- **Go** 1.25
- **Bubble Tea** v2 — TUI framework (Elm Architecture)
- **Lip Gloss** v2 — Estilização
- **Bubbles** v2 — Componentes (table, viewport)
- **Moby** — Docker SDK
- **golang.org/x/crypto/ssh** — SSH client

---

## Estrutura

```
barbarossa-cli/
├── main.go                     # Entrypoint
├── go.mod                      # Dependências
├── .barbarossa.yaml            # Configuração (exemplo)
├── internal/
│   ├── config/
│   │   └── config.go           # Leitura de YAML + env vars
│   ├── docker/
│   │   └── client.go           # Docker client (real stats + logs)
│   ├── ssh/
│   │   └── client.go           # SSH client interface (real + mock)
│   └── tui/
│       ├── app.go              # Model principal + navegação por abas
│       ├── styles.go           # Paleta, estilos, helpers, mensagens
│       ├── recon.go            # Tab RECON (tabela de findings)
│       └── logs.go             # Tab LOG (streaming viewport)
└── .specify/                   # Documentação de design e tarefas
```

---

## Licença

MIT
