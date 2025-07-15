package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Settings holds the schema definition for the Settings entity.
type Settings struct {
	ent.Schema
}

// Fields of the Settings.
func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty().
			Unique().
			Comment("Setting key identifier"),
		field.Text("value").
			Optional().
			Comment("Setting value"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Edges of the Settings.
func (Settings) Edges() []ent.Edge {
	return nil
}
