# AgentPlexus

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/agentplexus/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/agentplexus/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/agentplexus/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/agentplexus/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/agentplexus/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/agentplexus/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/agentplexus
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/agentplexus
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/agentplexus
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/agentplexus
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fagentplexus
 [loc-svg]: https://tokei.rs/b1/github/plexusone/agentplexus
 [repo-url]: https://github.com/plexusone/agentplexus
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/agentplexus/blob/master/LICENSE

A visual workflow designer for exploring and managing multi-agent team specifications. View agent relationships, inspect configurations, and save custom UI layouts.

![AgentPlexus web console](docs/agentplexus.png)

## Features

- **React Flow Canvas** - Interactive workflow visualization with draggable agent nodes and dependency edges
- **Floating Panels** - Gmail-style panels for viewing agent details with drag, resize, and minimize
- **Multi-Team View** - Display up to 4 agent teams simultaneously in a 2x2 grid layout
- **Persistent Views** - Save and restore UI layouts including selected teams, node positions, and sidebar width
- **Hot Reload** - File watcher automatically updates the UI when spec files change
- **Model Colors** - Panel headers colored by model type (opus=red, sonnet=blue, haiku=green)

## Installation

### Prerequisites

- Go 1.21+
- Node.js 18+

### Build

```bash
# Build frontend and backend
make build

# Or build separately
cd frontend && npm install && npm run build && cd ..
go build -o specui ./cmd/specui
```

## Usage

```bash
# Start with a workspace directory containing agent specs
./specui --workspace ~/path/to/agentplexus

# Or specify individual spec directories
./specui --spec-dirs /path/to/team1,/path/to/team2

# Custom port (default: 8090)
./specui --workspace ~/agentplexus --port 3000
```

Then open http://localhost:8090 in your browser.

## Agent Spec Format

The designer expects agent teams in directories with:

- `team.yaml` - Team metadata (name, version, description)
- `agents/*.yaml` - Individual agent specifications

Example team.yaml:

```yaml
name: fullstack-team
version: "1.0"
description: Full-stack development agent team
```

Example agent spec:

```yaml
name: backend
description: Backend API developer
model: sonnet
icon: go
tools:
  - Read
  - Write
  - Bash
dependencies:
  - architecture
instructions: |
  You are a backend developer specializing in Go APIs...
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DRIVER` | `sqlite` | Database driver: `sqlite` or `postgres` |
| `DB_DSN` | `file:data/specui.db?_pragma=foreign_keys(1)` | Database connection string |

### PostgreSQL Example

```bash
export DB_DRIVER=postgres
export DB_DSN="postgres://user:pass@localhost/specui?sslmode=disable"
./specui --workspace ~/agentplexus
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/teams` | GET | List all agent teams |
| `/api/agents/{team}/{agent}` | GET | Get agent details |
| `/api/views` | GET | List saved views |
| `/api/views` | POST | Create a new view |
| `/api/views/{id}` | GET | Get a specific view |
| `/api/views/{id}` | PUT | Update a view |
| `/api/views/{id}` | DELETE | Delete a view |

## Development

```bash
# Start in development mode (hot reload)
make dev

# This runs:
# - Frontend: npm run dev (Vite with HMR)
# - Backend: go run with air (optional)
```

## Architecture

```
cmd/specui/          # CLI entry point
internal/
  ent/               # Ent ORM schema and generated code
  server/            # HTTP handlers and routing
  storage/           # Database initialization
  watcher/           # File system watcher
frontend/
  src/
    components/
      Canvas.tsx         # React Flow wrapper
      AgentNode.tsx      # Custom node component
      AgentPanel.tsx     # Agent details view
      FloatingPanel.tsx  # Draggable panel container
      TeamSelector.tsx   # Sidebar with team list
      ViewsDropdown.tsx  # View management UI
```

## License

MIT
