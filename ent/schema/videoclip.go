package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// VideoClip holds the schema definition for the VideoClip entity.
type VideoClip struct {
	ent.Schema
}

// Fields of the VideoClip.
func (VideoClip) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Video clip name"),
		field.Text("description").
			Optional().
			Comment("Video clip description"),
		field.String("file_path").
			NotEmpty().
			Comment("Video clip file path"),
		field.Float("duration").
			Optional().
			Comment("Video duration in seconds"),
		field.String("format").
			Optional().
			Comment("Video format (mp4, mov, etc.)"),
		field.Int("width").
			Optional().
			Comment("Video width in pixels"),
		field.Int("height").
			Optional().
			Comment("Video height in pixels"),
		field.Int64("file_size").
			Optional().
			Comment("File size in bytes"),
		field.Text("transcription").
			Optional().
			Comment("Video transcription text"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Edges of the VideoClip.
func (VideoClip) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("video_clips").
			Unique().
			Comment("Project this video clip belongs to"),
	}
}
