package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentplexus/multi-agent-designer/internal/server"
	"github.com/agentplexus/multi-agent-designer/internal/storage"
)

func main() {
	var (
		specDirs  string
		workspace string
		port      int
	)

	flag.StringVar(&specDirs, "spec-dirs", "", "comma-separated list of spec directories")
	flag.StringVar(&workspace, "workspace", "", "parent directory to auto-discover agent-team-* repos")
	flag.IntVar(&port, "port", 8090, "port to serve on")
	flag.Parse()

	dirs, err := resolveSpecDirs(specDirs, workspace)
	if err != nil {
		log.Fatalf("error resolving spec directories: %v", err)
	}

	if len(dirs) == 0 {
		fmt.Fprintln(os.Stderr, "no spec directories found")
		fmt.Fprintln(os.Stderr, "usage:")
		fmt.Fprintln(os.Stderr, "  specui --spec-dirs=path/to/specs,path/to/other/specs")
		fmt.Fprintln(os.Stderr, "  specui --workspace=~/go/src/github.com/agentplexus")
		os.Exit(1)
	}

	// Initialize database
	ctx := context.Background()
	db, err := storage.NewClient(ctx, "")
	if err != nil {
		log.Fatalf("error initializing database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	fmt.Printf("multi-agent-designer v0.1.0\n")
	fmt.Printf("database: sqlite (%s)\n", storage.DefaultPath)
	fmt.Printf("spec directories:\n")
	for _, d := range dirs {
		fmt.Printf("  %s\n", d)
	}
	fmt.Printf("listening on http://localhost:%d\n", port)

	srv := server.New(dirs, port, db)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func resolveSpecDirs(specDirs, workspace string) ([]string, error) {
	if specDirs != "" {
		parts := strings.Split(specDirs, ",")
		var resolved []string
		for _, p := range parts {
			abs, err := filepath.Abs(strings.TrimSpace(p))
			if err != nil {
				return nil, fmt.Errorf("resolving %q: %w", p, err)
			}
			resolved = append(resolved, abs)
		}
		return resolved, nil
	}

	if workspace != "" {
		expanded := workspace
		if strings.HasPrefix(expanded, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			expanded = filepath.Join(home, expanded[1:])
		}
		return discoverAgentTeams(expanded)
	}

	return nil, nil
}

func discoverAgentTeams(workspace string) ([]string, error) {
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return nil, fmt.Errorf("reading workspace %q: %w", workspace, err)
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "agent-team-") {
			continue
		}
		specsDir := filepath.Join(workspace, entry.Name(), "specs")
		if info, err := os.Stat(specsDir); err == nil && info.IsDir() {
			dirs = append(dirs, specsDir)
		}
	}
	return dirs, nil
}
