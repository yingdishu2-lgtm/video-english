package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MediaStream stores one ffprobe stream entry.
type MediaStream struct {
	ent.Schema
}

func (MediaStream) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_item_id").Positive(),
		field.Int("stream_index").NonNegative(),
		field.String("stream_type"),
		field.String("codec"),
		field.String("language").Optional(),
		field.String("title").Optional(),
		field.Int("channels").Optional().Nillable().NonNegative(),
		field.Int("sample_rate").Optional().Nillable().NonNegative(),
		field.Int("width").Optional().Nillable().NonNegative(),
		field.Int("height").Optional().Nillable().NonNegative(),
		field.Bool("is_default").Default(false),
		field.Bool("is_forced").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (MediaStream) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_item_id", "stream_index").Unique(),
	}
}
