package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Position represents x,y coordinates for a node.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// View holds the schema definition for the View entity.
type View struct {
	ent.Schema
}

// Fields of the View.
func (View) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.JSON("teams", []string{}).
			Comment("Team names in display order"),
		field.JSON("node_positions", map[string]map[string]Position{}).
			Comment("teamName -> nodeId -> position"),
		field.Int("left_width").
			Default(240),
		field.Int("right_width").
			Default(480),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the View.
func (View) Edges() []ent.Edge {
	return nil
}
