package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ChatMessage holds the schema definition for the ChatMessage entity.
type ChatMessage struct {
	ent.Schema
}

// Fields of the ChatMessage.
func (ChatMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("message_id").
			NotEmpty().
			Comment("Unique message identifier"),
		field.Int("session_id").
			Comment("ID of the chat session this message belongs to"),
		field.Enum("role").
			Values("user", "assistant", "system", "error").
			Comment("Role of the message sender"),
		field.Text("content").
			NotEmpty().
			Comment("The message content"),
		field.Text("hidden_context").
			Optional().
			Comment("Hidden context not sent to frontend"),
		field.Time("timestamp").
			Default(time.Now).
			Comment("When the message was created"),
		field.String("model").
			Optional().
			Comment("LLM model used for assistant messages"),
	}
}

// Edges of the ChatMessage.
func (ChatMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", ChatSession.Type).
			Ref("messages").
			Field("session_id").
			Required().
			Unique(),
	}
}

// Indexes of the ChatMessage.
func (ChatMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id").Unique(),
		index.Fields("session_id"),
		index.Fields("session_id", "timestamp"),
		index.Fields("role"),
	}
}
