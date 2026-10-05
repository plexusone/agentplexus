package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentDefinition holds the schema definition for the agent-definition
// registry entity (the Definition plane of the AgentPlexus control plane).
//
// AgentDefinition records are upserted from multi-agent-spec agent markdown
// files (specs/agents/*.md) discovered across registered repositories. The
// (repo_name, namespace, name) triple is the stable identity key: re-running
// registration updates the existing row instead of creating a duplicate.
type AgentDefinition struct {
	ent.Schema
}

// Fields of the AgentDefinition.
func (AgentDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("repo_name").
			NotEmpty().
			Comment("Owning repository, e.g. agent-team-release"),
		field.String("namespace").
			Optional().
			Comment("Subdirectory namespace, if any"),
		field.String("name").
			NotEmpty().
			Comment("Agent name from spec frontmatter"),
		field.String("qualified_name").
			NotEmpty().
			Comment("repo_name/namespace/name, for display and lookup"),
		field.String("source_ref").
			Optional().
			Comment("User-facing Go-style source path, e.g. github.com/plexusone/agent-team-release/agents/release-reviewer"),
		field.String("registry_xrn").
			Optional().
			Immutable().
			Unique().
			Comment("Internal registry identity, xrn:agentplexus:agent-definition:<id>"),
		field.String("description").
			Optional(),
		field.String("model").
			Optional(),
		field.String("icon").
			Optional(),
		field.String("role").
			Optional(),
		field.JSON("tools", []string{}).
			Optional(),
		field.JSON("allowed_tools", []string{}).
			Optional(),
		field.JSON("skills", []string{}).
			Optional(),
		field.JSON("dependencies", []string{}).
			Optional(),
		field.JSON("requires", []string{}).
			Optional(),
		field.String("source_path").
			NotEmpty().
			Comment("Absolute path to the agent markdown file"),
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

// Edges of the AgentDefinition.
func (AgentDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("teams", TeamDefinition.Type).
			Ref("agents"),
	}
}

// Indexes of the AgentDefinition.
func (AgentDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_name", "namespace", "name").
			Unique(),
	}
}
