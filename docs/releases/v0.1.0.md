# Release Notes - v0.1.0

**Release Date:** 2026-01-25

## Overview

Initial release of Multi-Agent Designer, a visual workflow designer for exploring and managing multi-agent team specifications.

## Highlights

- **Visual Workflow Designer** - React Flow canvas for visualizing agent teams with custom nodes showing agent icons and model-based coloring
- **Gmail-Style Floating Panels** - View agent details in draggable, resizable, minimizable panels that can be positioned anywhere on screen
- **Persistent Views** - Save and restore your UI layouts including selected teams, node positions, and sidebar width using SQLite or PostgreSQL

## Features

### Frontend

- **React Flow Canvas**
  - Interactive drag-and-drop agent nodes
  - Automatic edge layout showing agent dependencies
  - Custom node styling with model-based colors (opus=red, sonnet=blue, haiku=green)
  - Support for agent icons (react, go, postgresql, etc.)

- **Floating Panels**
  - Open multiple agent detail panels simultaneously
  - Drag panels anywhere on screen
  - Resize from top-left or bottom-right corners
  - Minimize to header bar
  - Panel headers colored to match agent model
  - Staggered positioning for new panels

- **Multi-Team View**
  - Display up to 4 teams in a 2x2 grid layout
  - Each team canvas is independent with its own zoom/pan
  - Team labels in top-left corner

- **Views Management**
  - Save current layout as a named view
  - Load saved views from dropdown
  - Update existing views with current layout
  - Delete views no longer needed

- **Agent Details**
  - Model and namespace metadata
  - Description text
  - Tools and dependencies badges
  - Instructions with HTML/Code toggle
  - Markdown rendering with syntax highlighting

### Backend

- **HTTP Server**
  - Embedded frontend assets
  - RESTful API for teams, agents, and views
  - Graceful shutdown

- **File Watcher**
  - Monitors spec directories for changes
  - Automatically reloads team data
  - No server restart required

- **Database**
  - Ent ORM with auto-migration
  - SQLite for local development (default)
  - PostgreSQL for production deployments
  - Environment-based configuration

## Getting Started

```bash
# Build
make build

# Run with workspace
./specui --workspace ~/path/to/agentplexus

# Open browser
open http://localhost:8090
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DRIVER` | `sqlite` | Database driver |
| `DB_DSN` | `file:data/specui.db?_pragma=foreign_keys(1)` | Connection string |

## Known Limitations

- Maximum 4 teams displayed simultaneously
- Panel positions not persisted in views (only node positions)
- No search/filter functionality yet

## Contributors

- John Wang (@johncwang)
- Claude Opus 4.5 (AI pair programmer)
