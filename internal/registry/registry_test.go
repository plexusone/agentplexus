package registry

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/plexusone/agentplexus/internal/ent"
	entteamdef "github.com/plexusone/agentplexus/internal/ent/teamdefinition"
)

func newTestClient(t *testing.T) *ent.Client {
	t.Helper()

	// Matches internal/storage.NewClient's driver: modernc.org/sqlite
	// registers itself under "sqlite", not the CGO "sqlite3" name that
	// entsql.Open(dialect.SQLite, ...) assumes.
	db, err := sql.Open("sqlite", "file:ent?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return client
}

// writeAgent writes a minimal valid multi-agent-spec agent markdown file.
func writeAgent(t *testing.T, dir, name, description string) {
	t.Helper()
	content := "---\nname: " + name + "\ndescription: " + description + "\nmodel: sonnet\ntools: [Read]\n---\n\nInstructions for " + name + ".\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0o600); err != nil {
		t.Fatalf("write agent %s: %v", name, err)
	}
}

// writeTeam writes a minimal valid multi-agent-spec team JSON file named
// "fixture-team" with the given roster.
func writeTeam(t *testing.T, dir string, agents []string) {
	t.Helper()
	const name, version = "fixture-team", "1.0.0"
	agentsJSON := `"` + agents[0] + `"`
	for _, a := range agents[1:] {
		agentsJSON += `,"` + a + `"`
	}
	content := `{"name":"` + name + `","version":"` + version + `","agents":[` + agentsJSON + `]}`
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(content), 0o600); err != nil {
		t.Fatalf("write team %s: %v", name, err)
	}
}

func setupSpecDir(t *testing.T) string {
	t.Helper()
	specDir := filepath.Join(t.TempDir(), "agent-team-fixture", "specs")
	if err := os.MkdirAll(filepath.Join(specDir, "agents"), 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specDir, "teams"), 0o755); err != nil {
		t.Fatalf("mkdir teams: %v", err)
	}
	return specDir
}

func TestSync_CreatesAgentsAndTeams(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	writeAgent(t, filepath.Join(specDir, "agents"), "qa", "Quality checks")
	writeAgent(t, filepath.Join(specDir, "agents"), "release", "Ships the release")
	writeTeam(t, filepath.Join(specDir, "teams"), []string{"qa", "release"})

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.AgentsRegistered != 2 {
		t.Errorf("AgentsRegistered = %d, want 2", result.AgentsRegistered)
	}
	if result.TeamsRegistered != 1 {
		t.Errorf("TeamsRegistered = %d, want 1", result.TeamsRegistered)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", result.Warnings)
	}

	team, err := client.TeamDefinition.Query().Where(entteamdef.Name("fixture-team")).WithAgents().Only(ctx)
	if err != nil {
		t.Fatalf("query team: %v", err)
	}
	if len(team.Edges.Agents) != 2 {
		t.Errorf("team resolved %d agent edges, want 2", len(team.Edges.Agents))
	}
}

func TestSync_IsIdempotent(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	writeAgent(t, filepath.Join(specDir, "agents"), "qa", "Quality checks")
	writeTeam(t, filepath.Join(specDir, "teams"), []string{"qa"})

	if _, err := Sync(ctx, client, []string{specDir}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if result.AgentsRegistered != 0 || result.AgentsUpdated != 0 {
		t.Errorf("second sync of unchanged files should be a no-op, got %+v", result)
	}
	if result.TeamsRegistered != 0 || result.TeamsUpdated != 0 {
		t.Errorf("second sync of unchanged files should be a no-op, got %+v", result)
	}
}

func TestSync_DetectsContentChanges(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	agentsDir := filepath.Join(specDir, "agents")
	writeAgent(t, agentsDir, "qa", "Quality checks")
	if _, err := Sync(ctx, client, []string{specDir}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}

	writeAgent(t, agentsDir, "qa", "Quality checks, revised")
	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if result.AgentsUpdated != 1 {
		t.Errorf("AgentsUpdated = %d, want 1", result.AgentsUpdated)
	}

	agent, err := client.AgentDefinition.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query agent: %v", err)
	}
	if agent.Description != "Quality checks, revised" {
		t.Errorf("Description = %q, want updated value", agent.Description)
	}
}

func TestSync_WarnsOnDanglingOrchestratorAndWorkflowStepReferences(t *testing.T) {
	// The roster alone is valid (matches the old, narrower check); the
	// orchestrator and a workflow step reference agents that don't
	// exist anywhere. This is the class of drift the plain roster
	// check could never see, and the reason syncTeams now delegates to
	// multiagentspec.Team.ValidateAgentReferences instead of just
	// walking team.Agents.
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	writeAgent(t, filepath.Join(specDir, "agents"), "qa", "Quality checks")
	teamJSON := `{
		"name": "fixture-team",
		"version": "1.0.0",
		"agents": ["qa"],
		"orchestrator": "release-coordinator",
		"workflow": {
			"type": "chain",
			"steps": [
				{"name": "docs-validation", "agent": "documentation"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(specDir, "teams", "fixture-team.json"), []byte(teamJSON), 0o600); err != nil {
		t.Fatalf("write team: %v", err)
	}

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(result.Warnings) != 2 {
		t.Fatalf("Warnings = %v, want exactly 2 (orchestrator + workflow step)", result.Warnings)
	}
}

func TestSync_WarnsOnDanglingTeamReference(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	writeAgent(t, filepath.Join(specDir, "agents"), "qa", "Quality checks")
	writeTeam(t, filepath.Join(specDir, "teams"), []string{"qa", "ghost"})

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", result.Warnings)
	}
}

func TestSync_WarnsOnMalformedAgentFileWithoutAborting(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	agentsDir := filepath.Join(specDir, "agents")
	writeAgent(t, agentsDir, "qa", "Quality checks")
	if err := os.WriteFile(filepath.Join(agentsDir, "broken.md"), []byte("not valid frontmatter"), 0o600); err != nil {
		t.Fatalf("write broken agent: %v", err)
	}

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("Sync should not abort on a malformed file: %v", err)
	}
	if result.AgentsRegistered != 1 {
		t.Errorf("AgentsRegistered = %d, want 1 (the valid agent, skipping the broken one)", result.AgentsRegistered)
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Warnings = %v, want exactly 1 for the broken file", result.Warnings)
	}
}

func TestSync_PopulatesIdentityFields(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)

	writeAgent(t, filepath.Join(specDir, "agents"), "qa", "Quality checks")
	writeTeam(t, filepath.Join(specDir, "teams"), []string{"qa"})
	if _, err := Sync(ctx, client, []string{specDir}); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	agent, err := client.AgentDefinition.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query agent: %v", err)
	}
	if want := "xrn:agentplexus:agent-definition:" + agent.ID.String(); agent.RegistryXrn != want {
		t.Errorf("agent RegistryXrn = %q, want %q", agent.RegistryXrn, want)
	}
	if !strings.HasSuffix(agent.SourceRef, "/agents/qa") {
		t.Errorf("agent SourceRef = %q, want suffix /agents/qa", agent.SourceRef)
	}

	team, err := client.TeamDefinition.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query team: %v", err)
	}
	if want := "xrn:agentplexus:team-definition:" + team.ID.String(); team.RegistryXrn != want {
		t.Errorf("team RegistryXrn = %q, want %q", team.RegistryXrn, want)
	}
	if !strings.HasSuffix(team.SourceRef, "/teams/fixture-team") {
		t.Errorf("team SourceRef = %q, want suffix /teams/fixture-team", team.SourceRef)
	}
}

// TestSync_IdentityStableAcrossFileRename proves the registry keys on
// (repo, namespace, name), not file path: moving a spec to a new filename
// (same frontmatter name) reuses the existing row and its registry identity.
func TestSync_IdentityStableAcrossFileRename(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	specDir := setupSpecDir(t)
	agentsDir := filepath.Join(specDir, "agents")

	writeAgent(t, agentsDir, "qa", "Quality checks")
	if _, err := Sync(ctx, client, []string{specDir}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	before, err := client.AgentDefinition.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query agent: %v", err)
	}

	// Move the spec to a different filename, keeping identical content
	// (same frontmatter name "qa" — writeAgent's exact bytes).
	if err := os.Remove(filepath.Join(agentsDir, "qa.md")); err != nil {
		t.Fatalf("remove qa.md: %v", err)
	}
	content := "---\nname: qa\ndescription: Quality checks\nmodel: sonnet\ntools: [Read]\n---\n\nInstructions for qa.\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "quality.md"), []byte(content), 0o600); err != nil {
		t.Fatalf("write moved agent: %v", err)
	}

	result, err := Sync(ctx, client, []string{specDir})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if result.AgentsRegistered != 0 {
		t.Errorf("a renamed-but-same-identity spec should not register a new agent, got %+v", result)
	}
	after, err := client.AgentDefinition.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query agent after move: %v", err)
	}
	if after.ID != before.ID {
		t.Errorf("agent identity changed across file rename: %s -> %s", before.ID, after.ID)
	}
	if after.RegistryXrn != before.RegistryXrn {
		t.Errorf("registry_xrn changed across file rename: %s -> %s", before.RegistryXrn, after.RegistryXrn)
	}
}

func TestSync_MissingSpecSubdirsAreNotErrors(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	emptyDir := t.TempDir()

	result, err := Sync(ctx, client, []string{emptyDir})
	if err != nil {
		t.Fatalf("Sync on a repo with no agents/teams dirs should not error: %v", err)
	}
	if result.AgentsRegistered != 0 || result.TeamsRegistered != 0 {
		t.Errorf("expected no-op result, got %+v", result)
	}
}
