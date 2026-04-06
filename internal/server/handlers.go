package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/brandkit"
	"github.com/plexusone/agentplexus/internal/apxerr"
	multiagent "github.com/plexusone/multi-agent-spec/sdk/go"
)

// TeamSummary is the API response for listing teams.
type TeamSummary struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	SpecDir   string `json:"spec_dir"`
	RepoName  string `json:"repo_name"`
	NumAgents int    `json:"num_agents"`
}

// GraphNode represents a node in the DAG visualization.
type GraphNode struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Model       string   `json:"model"`
	Icon        string   `json:"icon,omitempty"`
	Tools       []string `json:"tools"`
	Position    Position `json:"position"`
}

// Position is the x,y position of a node.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// GraphEdge represents an edge in the DAG visualization.
type GraphEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// GraphResponse is the API response for a team's DAG.
type GraphResponse struct {
	Nodes    []GraphNode         `json:"nodes"`
	Edges    []GraphEdge         `json:"edges"`
	Warnings []map[string]string `json:"warnings,omitempty"`
}

// AgentDetail is the full agent response including instructions.
type AgentDetail struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	Icon         string   `json:"icon,omitempty"`
	Tools        []string `json:"tools"`
	Skills       []string `json:"skills,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Namespace    string   `json:"namespace,omitempty"`
	Instructions string   `json:"instructions"`
}

func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) {
	var teams []TeamSummary

	for _, dir := range s.specDirs {
		teamsDir := filepath.Join(dir, "teams")
		entries, err := os.ReadDir(teamsDir)
		if err != nil {
			continue
		}

		repoName := extractRepoName(dir)

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			teamFile := filepath.Join(teamsDir, entry.Name())
			team, err := multiagent.LoadTeamFromFile(teamFile)
			if err != nil {
				continue
			}

			teams = append(teams, TeamSummary{
				Name:      team.Name,
				Version:   team.Version,
				SpecDir:   dir,
				RepoName:  repoName,
				NumAgents: len(team.Agents),
			})
		}
	}

	writeJSON(w, http.StatusOK, teams)
}

func (s *Server) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.PathValue("team")

	team, _, _, err := s.findTeam(teamName)
	if err != nil {
		http.Error(w, "team not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, team)
}

func (s *Server) handleGetTeamGraph(w http.ResponseWriter, r *http.Request) {
	teamName := r.PathValue("team")

	team, specDir, teamFile, err := s.findTeam(teamName)
	if err != nil {
		http.Error(w, "team not found", http.StatusNotFound)
		return
	}

	// Validate and normalize team spec
	warnings := apxerr.ValidateTeam(team, teamFile)
	for _, warn := range warnings {
		warn.LogWarn(s.logger)
	}

	graph := buildGraph(team, specDir, warnings)
	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team")
	agentName := r.PathValue("agent")

	_, specDir, _, err := s.findTeam(teamID)
	if err != nil {
		http.Error(w, "team not found", http.StatusNotFound)
		return
	}

	agentFile := filepath.Join(specDir, "agents", agentName+".md")
	agent, err := multiagent.LoadAgentFromFile(agentFile)
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	// Re-read file to get instructions body
	data, err := os.ReadFile(agentFile)
	if err != nil {
		http.Error(w, "error reading agent", http.StatusInternalServerError)
		return
	}
	instructions := extractInstructions(string(data))

	detail := AgentDetail{
		Name:         agent.Name,
		Description:  agent.Description,
		Model:        string(agent.Model),
		Icon:         agent.Icon,
		Tools:        agent.Tools,
		Skills:       agent.Skills,
		Dependencies: agent.Dependencies,
		Namespace:    agent.Namespace,
		Instructions: instructions,
	}

	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) findTeam(name string) (*multiagent.Team, string, string, error) {
	for _, dir := range s.specDirs {
		teamsDir := filepath.Join(dir, "teams")
		entries, err := os.ReadDir(teamsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			teamFile := filepath.Join(teamsDir, entry.Name())
			team, err := multiagent.LoadTeamFromFile(teamFile)
			if err != nil {
				continue
			}
			if team.Name == name {
				return team, dir, teamFile, nil
			}
		}
	}
	return nil, "", "", os.ErrNotExist
}

func buildGraph(team *multiagent.Team, specDir string, warnings []apxerr.Warning) GraphResponse {
	var nodes []GraphNode
	var edges []GraphEdge

	// Convert warnings to JSON format
	var jsonWarnings []map[string]string
	for _, w := range warnings {
		jsonWarnings = append(jsonWarnings, w.ToJSON())
	}

	if team.Workflow == nil {
		// No workflow defined, show agents as standalone nodes horizontally
		for i, agentName := range team.Agents {
			agent := loadAgentSummary(specDir, agentName)
			nodes = append(nodes, GraphNode{
				ID:          agentName,
				Label:       agentName,
				Description: agent.Description,
				Model:       string(agent.Model),
				Icon:        agent.Icon,
				Tools:       agent.Tools,
				Position:    Position{X: float64(i*140) + 50, Y: 100},
			})
		}
		return GraphResponse{Nodes: nodes, Edges: edges, Warnings: jsonWarnings}
	}

	// Build from workflow steps
	stepPositions := layoutDAG(team.Workflow.Steps)

	for _, step := range team.Workflow.Steps {
		agent := loadAgentSummary(specDir, step.Agent)
		pos := stepPositions[step.Name]

		nodes = append(nodes, GraphNode{
			ID:          step.Name,
			Label:       step.Agent,
			Description: agent.Description,
			Model:       string(agent.Model),
			Icon:        agent.Icon,
			Tools:       agent.Tools,
			Position:    pos,
		})

		for _, dep := range step.DependsOn {
			edges = append(edges, GraphEdge{
				ID:     dep + "->" + step.Name,
				Source: dep,
				Target: step.Name,
			})
		}
	}

	return GraphResponse{Nodes: nodes, Edges: edges, Warnings: jsonWarnings}
}

func loadAgentSummary(specDir, name string) *multiagent.Agent {
	// Handle qualified names (namespace/agent-name)
	fileName := name
	if parts := strings.SplitN(name, "/", 2); len(parts) == 2 {
		fileName = parts[1]
	}

	agentFile := filepath.Join(specDir, "agents", fileName+".md")
	agent, err := multiagent.LoadAgentFromFile(agentFile)
	if err != nil {
		return &multiagent.Agent{Name: name}
	}
	return agent
}

func layoutDAG(steps []multiagent.Step) map[string]Position {
	positions := make(map[string]Position)

	// Compute levels via topological ordering
	levels := make(map[string]int)
	for _, step := range steps {
		computeLevel(step.Name, steps, levels, make(map[string]bool))
	}

	// Group steps by level
	levelGroups := make(map[int][]string)
	maxLevel := 0
	for name, level := range levels {
		levelGroups[level] = append(levelGroups[level], name)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Find max nodes in any level for vertical centering
	maxNodesInLevel := 0
	for _, names := range levelGroups {
		if len(names) > maxNodesInLevel {
			maxNodesInLevel = len(names)
		}
	}

	// Horizontal layout: X = level (left to right), Y = position within level (centered)
	nodeWidth := 100
	nodeHeight := 80
	horizontalGap := 40
	verticalGap := 20

	for level, names := range levelGroups {
		// Center nodes vertically within this level
		totalHeight := len(names)*nodeHeight + (len(names)-1)*verticalGap
		startY := (maxNodesInLevel*nodeHeight + (maxNodesInLevel-1)*verticalGap - totalHeight) / 2

		for i, name := range names {
			positions[name] = Position{
				X: float64(level*(nodeWidth+horizontalGap)) + 50,
				Y: float64(startY + i*(nodeHeight+verticalGap) + 50),
			}
		}
	}

	return positions
}

func computeLevel(name string, steps []multiagent.Step, levels map[string]int, visited map[string]bool) int {
	if l, ok := levels[name]; ok {
		return l
	}
	if visited[name] {
		return 0
	}
	visited[name] = true

	var step *multiagent.Step
	for i := range steps {
		if steps[i].Name == name {
			step = &steps[i]
			break
		}
	}
	if step == nil || len(step.DependsOn) == 0 {
		levels[name] = 0
		return 0
	}

	maxParent := 0
	for _, dep := range step.DependsOn {
		l := computeLevel(dep, steps, levels, visited)
		if l+1 > maxParent {
			maxParent = l + 1
		}
	}
	levels[name] = maxParent
	return maxParent
}

func extractInstructions(content string) string {
	// Split on second "---" to get body after frontmatter
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return content
	}
	return strings.TrimSpace(parts[2])
}

func extractRepoName(specDir string) string {
	// specDir is like /path/to/agent-team-webapp/specs
	parent := filepath.Dir(specDir)
	return filepath.Base(parent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding JSON response: %v", err)
	}
}

// handleGetIcon serves brandkit SVG icons.
// GET /api/icons/{brand} - returns white variant (default for dark UI)
// GET /api/icons/{brand}/{variant} - returns specified variant (white, color, orig)
func (s *Server) handleGetIcon(w http.ResponseWriter, r *http.Request) {
	brand := r.PathValue("brand")
	variant := r.PathValue("variant")

	if variant == "" {
		variant = "white" // Default to white for dark UI backgrounds
	}

	var svg []byte
	var err error

	switch variant {
	case "white":
		svg, err = brandkit.GetIconWhite(brand)
	case "color":
		svg, err = brandkit.GetIconColor(brand)
	case "orig":
		svg, err = brandkit.GetIconOrig(brand)
	default:
		http.Error(w, "invalid variant: use white, color, or orig", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "icon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	if _, err = w.Write(svg); err != nil {
		log.Printf("error writing icon response: %v", err)
	}
}

// handleListIcons returns all available brandkit icon names.
func (s *Server) handleListIcons(w http.ResponseWriter, r *http.Request) {
	icons, err := brandkit.ListIcons()
	if err != nil {
		http.Error(w, "failed to list icons", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, icons)
}
