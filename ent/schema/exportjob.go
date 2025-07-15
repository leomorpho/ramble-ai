package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ExportJob holds the schema definition for the ExportJob entity.
type ExportJob struct {
	ent.Schema
}

// Fields of the ExportJob.
func (ExportJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("job_id").
			Unique().
			NotEmpty().
			Comment("Unique job identifier"),
		field.String("export_type").
			NotEmpty().
			Comment("Type of export: 'stitched' or 'individual'"),
		field.String("output_path").
			NotEmpty().
			Comment("Export destination folder path"),
		field.String("stage").
			Default("starting").
			Comment("Current stage of export"),
		field.Float("progress").
			Default(0.0).
			Comment("Progress percentage (0.0 to 1.0)"),
		field.String("current_file").
			Optional().
			Comment("Currently processing file name"),
		field.Int("total_files").
			Default(0).
			Comment("Total number of files to process"),
		field.Int("processed_files").
			Default(0).
			Comment("Number of files processed"),
		field.Bool("is_complete").
			Default(false).
			Comment("Whether the job is complete"),
		field.Bool("has_error").
			Default(false).
			Comment("Whether the job has an error"),
		field.Text("error_message").
			Optional().
			Comment("Error message if job failed"),
		field.Bool("is_cancelled").
			Default(false).
			Comment("Whether the job was cancelled"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Job creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
		field.Time("completed_at").
			Optional().
			Comment("Job completion timestamp"),
	}
}

// Edges of the ExportJob.
func (ExportJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("export_jobs").
			Unique().
			Comment("Project this export job belongs to"),
	}
}
