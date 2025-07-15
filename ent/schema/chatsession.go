package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ChatSession holds the schema definition for the ChatSession entity.
type ChatSession struct {
	ent.Schema
}

// Fields of the ChatSession.
func (ChatSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").
			NotEmpty().
			Comment("Unique session identifier"),
		field.Int("project_id").
			Comment("ID of the project this session belongs to"),
		field.String("endpoint_id").
			NotEmpty().
			Comment("Chatbot endpoint identifier (e.g., 'highlight_ordering')"),
		field.Time("created_at").
			Default(time.Now).
			Comment("When the session was created"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("When the session was last updated"),
		field.String("selected_model").
			Optional().
			Comment("The AI model selected for this chat session"),
	}
}

// Edges of the ChatSession.
func (ChatSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("chat_sessions").
			Field("project_id").
			Required().
			Unique(),
		edge.To("messages", ChatMessage.Type),
	}
}

// Indexes of the ChatSession.
func (ChatSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id").Unique(),
		index.Fields("project_id", "endpoint_id").Unique(),
		index.Fields("project_id"),
		index.Fields("endpoint_id"),
	}
}
