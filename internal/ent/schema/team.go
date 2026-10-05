package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TeamDefinition holds the schema definition for the team-definition
// registry entity (the Definition plane of the AgentPlexus control plane).
//
// TeamDefinition records are upserted from multi-agent-spec team JSON files
// (specs/teams/*.json) discovered across registered repositories. The
// (repo_name, name) pair is the stable identity key.
type TeamDefinition struct {
	ent.Schema
}

// Fields of the TeamDefinition.
func (TeamDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("repo_name").
			NotEmpty().
			Comment("Owning repository, e.g. agent-team-release"),
		field.String("name").
			NotEmpty().
			Comment("Team name from spec"),
		field.String("source_ref").
			Optional().
			Comment("User-facing Go-style source path, e.g. github.com/plexusone/agent-team-release/teams/release"),
		field.String("registry_xrn").
			Optional().
			Immutable().
			Unique().
			Comment("Internal registry identity, xrn:agentplexus:team-definition:<id>"),
		field.String("version").
			Optional(),
		field.String("description").
			Optional(),
		field.String("orchestrator").
			Optional().
			Comment("Agent name that orchestrates this team's workflow"),
		field.JSON("agent_names", []string{}).
			Optional().
			Comment("Agent names referenced by the team spec, in spec order"),
		field.String("source_path").
			NotEmpty().
			Comment("Absolute path to the team JSON file"),
		field.String("content_hash").
			NotEmpty().
			Comment("sha256 of the source file, for drift detection"),
		field.Time("first_seen_at").
			Default(time.Now).
			Immutable(),
		field.Time("last_seen_at").
			Default(time.Now).
			Comment("Updated on every successful registration sync"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the TeamDefinition.
func (TeamDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("agents", AgentDefinition.Type),
	}
}

// Indexes of the TeamDefinition.
func (TeamDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_name", "name").
			Unique(),
	}
}
