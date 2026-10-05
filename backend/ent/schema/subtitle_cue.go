package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubtitleCue is the format-independent subtitle timeline unit.
type SubtitleCue struct {
	ent.Schema
}

func (SubtitleCue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("subtitle_track_id").Positive(),
		field.Int("sequence").NonNegative(),
		field.Int64("start_ms").NonNegative(),
		field.Int64("end_ms").Positive(),
		field.String("text"),
		field.String("clean_text"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (SubtitleCue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subtitle_track_id", "sequence").Unique(),
		index.Fields("subtitle_track_id", "start_ms"),
	}
}
