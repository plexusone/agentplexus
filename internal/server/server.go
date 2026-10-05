package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/plexusone/agentplexus/internal/ent"
	"github.com/plexusone/agentplexus/internal/registry"
	"github.com/plexusone/agentplexus/internal/watcher"
)

//go:embed frontend
var frontendFS embed.FS

// Server serves the API and embedded frontend.
type Server struct {
	specDirs []string
	port     int
	mux      *http.ServeMux
	watcher  *watcher.Watcher
	db       *ent.Client
	logger   *slog.Logger
}

// New creates a new Server for the given spec directories.
func New(specDirs []string, port int, db *ent.Client) *Server {
	s := &Server{
		specDirs: specDirs,
		port:     port,
		mux:      http.NewServeMux(),
		db:       db,
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// API routes
	s.mux.HandleFunc("/api/teams", s.handleListTeams)
	s.mux.HandleFunc("/api/teams/{team}", s.handleGetTeam)
	s.mux.HandleFunc("/api/teams/{team}/graph", s.handleGetTeamGraph)
	s.mux.HandleFunc("/api/agents/{team}/{agent}", s.handleGetAgent)

	// Icon routes (brandkit integration)
	s.mux.HandleFunc("/api/icons", s.handleListIcons)
	s.mux.HandleFunc("/api/icons/{brand}", s.handleGetIcon)
	s.mux.HandleFunc("/api/icons/{brand}/{variant}", s.handleGetIcon)

	// View routes (saved layouts)
	s.mux.HandleFunc("GET /api/views", s.handleListViews)
	s.mux.HandleFunc("POST /api/views", s.handleCreateView)
	s.mux.HandleFunc("GET /api/views/{id}", s.handleGetView)
	s.mux.HandleFunc("PUT /api/views/{id}", s.handleUpdateView)
	s.mux.HandleFunc("DELETE /api/views/{id}", s.handleDeleteView)

	// Registry routes (Agent Workforce inventory)
	s.mux.HandleFunc("GET /api/registry/agents", s.handleListRegistryAgents)
	s.mux.HandleFunc("GET /api/registry/teams", s.handleListRegistryTeams)
	s.mux.HandleFunc("POST /api/registry/sync", s.handleRegistrySync)

	// Serve embedded frontend (catch-all for non-API routes)
	frontendSub, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		panic(fmt.Sprintf("frontend embed: %v", err))
	}
	s.mux.Handle("/", http.FileServer(http.FS(frontendSub)))
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	w, err := watcher.New(s.specDirs)
	if err != nil {
		return fmt.Errorf("starting watcher: %w", err)
	}
	s.watcher = w
	go s.watcher.Run()
	go s.watchRegistrySync()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

// watchRegistrySync re-syncs the agent/team registry whenever the spec
// watcher reports a change, so the registry stays current without
// requiring a server restart. A single worker goroutine runs syncs
// serially; bursts of watcher events coalesce into at most one pending
// sync via the buffered signal channel.
func (s *Server) watchRegistrySync() {
	dirty := make(chan struct{}, 1)

	go func() {
		for range dirty {
			result, err := registry.Sync(context.Background(), s.db, s.specDirs)
			if err != nil {
				s.logger.Error("registry sync failed", "error", err)
				continue
			}
			s.logger.Info("registry synced", "result", result.String())
		}
	}()

	for range s.watcher.Events {
		select {
		case dirty <- struct{}{}:
		default:
		}
	}
}

// Close cleans up server resources.
func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
