package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Project name"),
		field.Text("description").
			Optional().
			Comment("Project description"),
		field.String("path").
			NotEmpty().
			Comment("Project file path"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("video_clips", VideoClip.Type).
			Comment("Video clips in this project"),
		edge.To("export_jobs", ExportJob.Type).
			Comment("Export jobs for this project"),
	}
}
