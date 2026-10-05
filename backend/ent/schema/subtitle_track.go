package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubtitleTrack is an embedded or external subtitle source.
type SubtitleTrack struct {
	ent.Schema
}

func (SubtitleTrack) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_item_id").Positive(),
		field.String("source_type"),
		field.Int("stream_index").Optional().Nillable().NonNegative(),
		field.String("format"),
		field.String("language").Optional(),
		field.String("title").Optional(),
		field.String("external_path").Optional(),
		field.Int64("offset_ms").Default(0),
		field.Bool("supported").Default(true),
		field.Bool("processed").Default(false),
		field.String("source_hash").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (SubtitleTrack) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_item_id"),
		index.Fields("media_item_id", "source_type", "stream_index"),
		index.Fields("media_item_id", "external_path"),
	}
}
