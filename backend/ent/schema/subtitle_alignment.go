package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubtitleAlignment caches one English-led bilingual learning cue.
type SubtitleAlignment struct {
	ent.Schema
}

func (SubtitleAlignment) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_item_id").Positive(),
		field.Int("english_track_id").Positive(),
		field.Int("chinese_track_id").Positive(),
		field.Int("sequence").NonNegative(),
		field.Int64("start_ms").NonNegative(),
		field.Int64("end_ms").Positive(),
		field.String("text_en"),
		field.String("text_zh").Optional(),
		field.String("en_cue_ids_json"),
		field.String("zh_cue_ids_json"),
		field.String("cache_key"),
		field.Int("version").Default(1),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (SubtitleAlignment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("english_track_id", "chinese_track_id", "sequence").Unique(),
		index.Fields("english_track_id", "chinese_track_id", "start_ms"),
		index.Fields("cache_key"),
	}
}
