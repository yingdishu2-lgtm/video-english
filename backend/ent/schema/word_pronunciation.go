package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WordPronunciation stores a normalized dictionary lookup result.
type WordPronunciation struct {
	ent.Schema
}

func (WordPronunciation) Fields() []ent.Field {
	return []ent.Field{
		field.String("word"),
		field.String("phonetic").Optional(),
		field.String("audio_url").Optional(),
		field.String("source"),
		field.Bool("found").Default(false),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (WordPronunciation) Indexes() []ent.Index {
	return []ent.Index{index.Fields("word").Unique()}
}
