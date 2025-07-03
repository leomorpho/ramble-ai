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
		field.String("ai_model").
			Optional().
			Default("anthropic/claude-3.5-haiku-20241022").
			Comment("Preferred OpenRouter AI model for this project"),
		field.Text("ai_prompt").
			Optional().
			Comment("Custom AI prompt for segment reordering"),
		field.JSON("ai_suggestion_order", []string{}).
			Optional().
			Comment("Cached AI-suggested highlight order (array of highlight IDs)"),
		field.String("ai_suggestion_model").
			Optional().
			Comment("AI model used for the cached suggestion"),
		field.Time("ai_suggestion_created_at").
			Optional().
			Comment("When the AI suggestion was created"),
		field.String("ai_highlight_model").
			Optional().
			Default("anthropic/claude-sonnet-4").
			Comment("Preferred OpenRouter AI model for highlight suggestions"),
		field.Text("ai_highlight_prompt").
			Optional().
			Comment("Custom AI prompt for highlight suggestions"),
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
