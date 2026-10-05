// Package registry syncs multi-agent-spec team and agent definitions found
// on disk into AgentPlexus's persistent registry (the ent AgentDefinition
// and TeamDefinition entities — the Definition plane of the control plane).
// It provides a stable, queryable inventory of the agents already defined in
// agent-team-* repositories, keyed by (repo, namespace, name) rather than by
// file path so identity survives file moves.
package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/plexusone/agentplexus/internal/ent"
	entagentdef "github.com/plexusone/agentplexus/internal/ent/agentdefinition"
	entteamdef "github.com/plexusone/agentplexus/internal/ent/teamdefinition"
)

// syncMu serializes Sync calls. The watcher-triggered resync and the
// POST /api/registry/sync handler can otherwise run concurrently: both
// can observe ent.IsNotFound for the same not-yet-registered agent or
// team and both attempt Create, and the loser of that race fails on the
// unique index and aborts its whole Sync call.
var syncMu sync.Mutex

// Result summarizes one Sync call.
type Result struct {
	AgentsRegistered int
	AgentsUpdated    int
	TeamsRegistered  int
	TeamsUpdated     int
	// Warnings holds non-fatal problems, e.g. a team referencing an agent
	// whose spec file could not be found or parsed.
	Warnings []string
}

func (r Result) String() string {
	return fmt.Sprintf("agents: %d new, %d updated · teams: %d new, %d updated · %d warning(s)",
		r.AgentsRegistered, r.AgentsUpdated, r.TeamsRegistered, r.TeamsUpdated, len(r.Warnings))
}

// Sync scans each specDir for multi-agent-spec agent and team files and
// upserts them into the registry. specDir is expected to contain
// "agents/*.md" and "teams/*.json", matching the layout the AgentPlexus
// viewer already reads. Sync is idempotent: re-running it against
// unchanged files only refreshes last_seen_at.
func Sync(ctx context.Context, client *ent.Client, specDirs []string) (Result, error) {
	syncMu.Lock()
	defer syncMu.Unlock()

	var result Result

	for _, specDir := range specDirs {
		repoName := repoNameFromSpecDir(specDir)

		agentIDs, agentStats, err := syncAgents(ctx, client, repoName, specDir)
		if err != nil {
			return result, fmt.Errorf("sync agents for %s: %w", repoName, err)
		}
		result.AgentsRegistered += agentStats.registered
		result.AgentsUpdated += agentStats.updated
		result.Warnings = append(result.Warnings, agentStats.warnings...)

		teamStats, err := syncTeams(ctx, client, repoName, specDir, agentIDs)
		if err != nil {
			return result, fmt.Errorf("sync teams for %s: %w", repoName, err)
		}
		result.TeamsRegistered += teamStats.registered
		result.TeamsUpdated += teamStats.updated
		result.Warnings = append(result.Warnings, teamStats.warnings...)
	}

	return result, nil
}

type syncStats struct {
	registered int
	updated    int
	warnings   []string
}

// agentKey identifies an agent within a single repo. Team specs reference
// agents by bare name (no namespace), so lookups use namespace "".
type agentKey struct {
	namespace string
	name      string
}

func syncAgents(ctx context.Context, client *ent.Client, repoName, specDir string) (map[agentKey]uuid.UUID, syncStats, error) {
	var stats syncStats
	ids := make(map[agentKey]uuid.UUID)
	moduleRef := moduleRefFromSpecDir(specDir)

	agentsDir := filepath.Join(specDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ids, stats, nil
		}
		return ids, stats, fmt.Errorf("read %s: %w", agentsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		sourcePath := filepath.Join(agentsDir, entry.Name())
		// A malformed spec file shouldn't abort registration for the rest
		// of the repo (or workspace); record it and move on.
		spec, err := multiagentspec.LoadAgentFromFile(sourcePath)
		if err != nil {
			stats.warnings = append(stats.warnings, fmt.Sprintf("%s: parse %s: %v", repoName, entry.Name(), err))
			continue
		}

		hash, err := hashFile(sourcePath)
		if err != nil {
			stats.warnings = append(stats.warnings, fmt.Sprintf("%s: hash %s: %v", repoName, spec.Name, err))
			continue
		}

		existing, err := client.AgentDefinition.Query().
			Where(
				entagentdef.RepoName(repoName),
				entagentdef.Namespace(spec.Namespace),
				entagentdef.Name(spec.Name),
			).
			Only(ctx)

		switch {
		case ent.IsNotFound(err):
			id := uuid.New()
			created, cerr := client.AgentDefinition.Create().
				SetID(id).
				SetRegistryXrn("xrn:agentplexus:agent-definition:" + id.String()).
				SetSourceRef(moduleRef + "/agents/" + spec.QualifiedName()).
				SetRepoName(repoName).
				SetNamespace(spec.Namespace).
				SetName(spec.Name).
				SetQualifiedName(spec.QualifiedName()).
				SetDescription(spec.Description).
				SetModel(string(spec.Model)).
				SetIcon(spec.Icon).
				SetRole(spec.Role).
				SetTools(spec.Tools).
				SetAllowedTools(spec.AllowedTools).
				SetSkills(spec.Skills).
				SetDependencies(spec.Dependencies).
				SetRequires(spec.Requires).
				SetSourcePath(sourcePath).
				SetContentHash(hash).
				Save(ctx)
			if cerr != nil {
				return ids, stats, fmt.Errorf("create agent %s/%s: %w", repoName, spec.Name, cerr)
			}
			ids[agentKey{spec.Namespace, spec.Name}] = created.ID
			stats.registered++

		case err != nil:
			return ids, stats, fmt.Errorf("query agent %s/%s: %w", repoName, spec.Name, err)

		default:
			update := existing.Update().SetLastSeenAt(time.Now())
			if existing.ContentHash != hash {
				update = update.
					SetSourceRef(moduleRef + "/agents/" + spec.QualifiedName()).
					SetQualifiedName(spec.QualifiedName()).
					SetDescription(spec.Description).
					SetModel(string(spec.Model)).
					SetIcon(spec.Icon).
					SetRole(spec.Role).
					SetTools(spec.Tools).
					SetAllowedTools(spec.AllowedTools).
					SetSkills(spec.Skills).
					SetDependencies(spec.Dependencies).
					SetRequires(spec.Requires).
					SetSourcePath(sourcePath).
					SetContentHash(hash)
				stats.updated++
			}
			saved, uerr := update.Save(ctx)
			if uerr != nil {
				return ids, stats, fmt.Errorf("update agent %s/%s: %w", repoName, spec.Name, uerr)
			}
			ids[agentKey{spec.Namespace, spec.Name}] = saved.ID
		}
	}

	return ids, stats, nil
}

func syncTeams(ctx context.Context, client *ent.Client, repoName, specDir string, agentIDs map[agentKey]uuid.UUID) (syncStats, error) {
	var stats syncStats
	moduleRef := moduleRefFromSpecDir(specDir)

	// Unqualified names available in this repo, for Team.ValidateAgentReferences.
	// Team specs in these repos reference agents by bare name, never namespaced.
	availableAgents := make([]string, 0, len(agentIDs))
	for key := range agentIDs {
		if key.namespace == "" {
			availableAgents = append(availableAgents, key.name)
		}
	}

	teamsDir := filepath.Join(specDir, "teams")
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return stats, fmt.Errorf("read %s: %w", teamsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		sourcePath := filepath.Join(teamsDir, entry.Name())
		team, err := multiagentspec.LoadTeamFromFile(sourcePath)
		if err != nil {
			stats.warnings = append(stats.warnings, fmt.Sprintf("%s: load team %s: %v", repoName, entry.Name(), err))
			continue
		}

		hash, err := hashFile(sourcePath)
		if err != nil {
			stats.warnings = append(stats.warnings, fmt.Sprintf("%s: hash team %s: %v", repoName, team.Name, err))
			continue
		}

		// Checks the roster, orchestrator, and agent-assigned workflow
		// steps in one pass — catches drift my own membership resolution
		// below can't see (e.g. a stale orchestrator or workflow step
		// name), not just a missing roster entry.
		if verr := team.ValidateAgentReferences(availableAgents); verr != nil {
			if joined, ok := verr.(interface{ Unwrap() []error }); ok {
				for _, sub := range joined.Unwrap() {
					stats.warnings = append(stats.warnings, fmt.Sprintf("%s: %v", repoName, sub))
				}
			} else {
				stats.warnings = append(stats.warnings, fmt.Sprintf("%s: %v", repoName, verr))
			}
		}

		var memberIDs []uuid.UUID
		for _, agentName := range team.Agents {
			if id, ok := agentIDs[agentKey{"", agentName}]; ok {
				memberIDs = append(memberIDs, id)
			}
		}

		existing, err := client.TeamDefinition.Query().
			Where(entteamdef.RepoName(repoName), entteamdef.Name(team.Name)).
			Only(ctx)

		switch {
		case ent.IsNotFound(err):
			teamID := uuid.New()
			if _, cerr := client.TeamDefinition.Create().
				SetID(teamID).
				SetRegistryXrn("xrn:agentplexus:team-definition:" + teamID.String()).
				SetSourceRef(moduleRef + "/teams/" + team.Name).
				SetRepoName(repoName).
				SetName(team.Name).
				SetVersion(team.Version).
				SetDescription(team.Description).
				SetOrchestrator(team.Orchestrator).
				SetAgentNames(team.Agents).
				SetSourcePath(sourcePath).
				SetContentHash(hash).
				AddAgentIDs(memberIDs...).
				Save(ctx); cerr != nil {
				return stats, fmt.Errorf("create team %s/%s: %w", repoName, team.Name, cerr)
			}
			stats.registered++

		case err != nil:
			return stats, fmt.Errorf("query team %s/%s: %w", repoName, team.Name, err)

		default:
			update := existing.Update().SetLastSeenAt(time.Now())
			if existing.ContentHash != hash {
				update = update.
					SetSourceRef(moduleRef + "/teams/" + team.Name).
					SetVersion(team.Version).
					SetDescription(team.Description).
					SetOrchestrator(team.Orchestrator).
					SetAgentNames(team.Agents).
					SetSourcePath(sourcePath).
					SetContentHash(hash)
				stats.updated++
			}
			// Membership resolution depends on the agent registry, not just
			// this team file's content, so it must be refreshed on every
			// sync: an agent referenced by the roster can be registered (or
			// re-identified) after this team file was last seen, and that
			// shouldn't require touching team.json to pick up the edge.
			update = update.ClearAgents().AddAgentIDs(memberIDs...)
			if _, uerr := update.Save(ctx); uerr != nil {
				return stats, fmt.Errorf("update team %s/%s: %w", repoName, team.Name, uerr)
			}
		}
	}

	return stats, nil
}

// repoNameFromSpecDir derives the owning repository name from a spec
// directory path, e.g. ".../agent-team-release/specs" -> "agent-team-release".
func repoNameFromSpecDir(specDir string) string {
	return filepath.Base(filepath.Dir(specDir))
}

// moduleRefFromSpecDir derives the Go-style module path of the owning
// repository from a spec directory, e.g.
// ".../github.com/plexusone/agent-team-release/specs" ->
// "github.com/plexusone/agent-team-release". It scans for a host-like
// component (one containing a ".") and returns that plus the next two path
// segments (org/repo). If no such segment is found — e.g. a spec dir outside
// the standard GOPATH layout — it falls back to the bare repository name so
// source_ref is always populated with something identifying.
func moduleRefFromSpecDir(specDir string) string {
	parts := strings.Split(filepath.ToSlash(filepath.Dir(specDir)), "/")
	for i := 0; i+2 < len(parts); i++ {
		if strings.Contains(parts[i], ".") {
			return strings.Join(parts[i:i+3], "/")
		}
	}
	return repoNameFromSpecDir(specDir)
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
