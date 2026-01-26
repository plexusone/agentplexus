package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/agentplexus/multi-agent-designer/internal/ent"
	"github.com/agentplexus/multi-agent-designer/internal/watcher"
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
}

// New creates a new Server for the given spec directories.
func New(specDirs []string, port int, db *ent.Client) *Server {
	s := &Server{
		specDirs: specDirs,
		port:     port,
		mux:      http.NewServeMux(),
		db:       db,
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

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

// Close cleans up server resources.
func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
