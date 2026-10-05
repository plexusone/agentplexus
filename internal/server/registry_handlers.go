package server

import (
	"net/http"

	"github.com/plexusone/agentplexus/internal/registry"
)

// RegistryAgent is the API response for a registered agent.
type RegistryAgent struct {
	ID            string   `json:"id"`
	RegistryXRN   string   `json:"registry_xrn,omitempty"`
	SourceRef     string   `json:"source_ref,omitempty"`
	RepoName      string   `json:"repo_name"`
	Namespace     string   `json:"namespace,omitempty"`
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualified_name"`
	Description   string   `json:"description,omitempty"`
	Model         string   `json:"model,omitempty"`
	Role          string   `json:"role,omitempty"`
	Tools         []string `json:"tools,omitempty"`
	Skills        []string `json:"skills,omitempty"`
	SourcePath    string   `json:"source_path"`
	FirstSeenAt   string   `json:"first_seen_at"`
	LastSeenAt    string   `json:"last_seen_at"`
}

// RegistryTeam is the API response for a registered team.
type RegistryTeam struct {
	ID           string   `json:"id"`
	RegistryXRN  string   `json:"registry_xrn,omitempty"`
	SourceRef    string   `json:"source_ref,omitempty"`
	RepoName     string   `json:"repo_name"`
	Name         string   `json:"name"`
	Version      string   `json:"version,omitempty"`
	Description  string   `json:"description,omitempty"`
	Orchestrator string   `json:"orchestrator,omitempty"`
	AgentNames   []string `json:"agent_names,omitempty"`
	SourcePath   string   `json:"source_path"`
	FirstSeenAt  string   `json:"first_seen_at"`
	LastSeenAt   string   `json:"last_seen_at"`
}

// handleListRegistryAgents returns every agent in the registry.
// GET /api/registry/agents
func (s *Server) handleListRegistryAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.db.AgentDefinition.Query().All(r.Context())
	if err != nil {
		http.Error(w, "error querying agents", http.StatusInternalServerError)
		return
	}

	out := make([]RegistryAgent, 0, len(agents))
	for _, a := range agents {
		out = append(out, RegistryAgent{
			ID:            a.ID.String(),
			RegistryXRN:   a.RegistryXrn,
			SourceRef:     a.SourceRef,
			RepoName:      a.RepoName,
			Namespace:     a.Namespace,
			Name:          a.Name,
			QualifiedName: a.QualifiedName,
			Description:   a.Description,
			Model:         a.Model,
			Role:          a.Role,
			Tools:         a.Tools,
			Skills:        a.Skills,
			SourcePath:    a.SourcePath,
			FirstSeenAt:   a.FirstSeenAt.Format(timeFormat),
			LastSeenAt:    a.LastSeenAt.Format(timeFormat),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

// handleListRegistryTeams returns every team in the registry, including
// which registered agents each team resolved to.
// GET /api/registry/teams
func (s *Server) handleListRegistryTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := s.db.TeamDefinition.Query().WithAgents().All(r.Context())
	if err != nil {
		http.Error(w, "error querying teams", http.StatusInternalServerError)
		return
	}

	out := make([]RegistryTeam, 0, len(teams))
	for _, t := range teams {
		out = append(out, RegistryTeam{
			ID:           t.ID.String(),
			RegistryXRN:  t.RegistryXrn,
			SourceRef:    t.SourceRef,
			RepoName:     t.RepoName,
			Name:         t.Name,
			Version:      t.Version,
			Description:  t.Description,
			Orchestrator: t.Orchestrator,
			AgentNames:   t.AgentNames,
			SourcePath:   t.SourcePath,
			FirstSeenAt:  t.FirstSeenAt.Format(timeFormat),
			LastSeenAt:   t.LastSeenAt.Format(timeFormat),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

// handleRegistrySync re-scans all spec directories and upserts the
// registry immediately, returning a summary of what changed.
// POST /api/registry/sync
func (s *Server) handleRegistrySync(w http.ResponseWriter, r *http.Request) {
	result, err := registry.Sync(r.Context(), s.db, s.specDirs)
	if err != nil {
		http.Error(w, "error syncing registry: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
