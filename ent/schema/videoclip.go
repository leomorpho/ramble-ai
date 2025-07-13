package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Word represents a word with timestamps for transcription
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Highlight represents a highlighted text region with timestamps
type Highlight struct {
	ID      string  `json:"id"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	ColorID int     `json:"colorId"`
}

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
		field.JSON("transcription_words", []Word{}).
			Optional().
			Comment("Word-level transcription with timestamps"),
		field.String("transcription_language").
			Optional().
			Comment("Detected language of transcription"),
		field.Float("transcription_duration").
			Optional().
			Comment("Duration of transcribed audio in seconds"),
		field.JSON("highlights", []Highlight{}).
			Optional().
			Comment("Highlighted text regions with timestamps"),
		field.JSON("suggested_highlights", []Highlight{}).
			Optional().
			Comment("AI-suggested highlights pending user confirmation"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
		field.JSON("highlights_history", [][]Highlight{}).
			Optional().
			Comment("FIFO history of highlight states (last 20 states)"),
		field.Int("highlights_history_index").
			Optional().
			Default(-1).
			Comment("Current position in highlights history (-1 = no history)"),
		field.String("transcription_state").
			Optional().
			Default("idle").
			Comment("Current state of transcription: idle, checking, transcribing, completed, error"),
		field.String("transcription_error").
			Optional().
			Comment("Error message if transcription failed"),
		field.Time("transcription_started_at").
			Optional().
			Comment("When transcription was started"),
		field.Time("transcription_completed_at").
			Optional().
			Comment("When transcription was completed"),
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
