package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MediaItem describes a video found below MEDIA_ROOT.
type MediaItem struct {
	ent.Schema
}

func (MediaItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("path").Unique(),
		field.String("filename"),
		field.String("title"),
		field.String("container"),
		field.Int64("duration_ms").NonNegative(),
		field.String("video_codec").Optional(),
		field.Int("width").Optional().NonNegative(),
		field.Int("height").Optional().NonNegative(),
		field.Int64("file_size").NonNegative(),
		field.Time("file_mtime"),
		field.String("probe_hash"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}
