package subtitle

import "testing"

func TestAlignCuesHandlesOneToManyAndNearestFallback(t *testing.T) {
	english := []timedCue{
		{ID: 1, Sequence: 0, StartMs: 1000, EndMs: 3000, CleanText: "One sentence."},
		{ID: 2, Sequence: 1, StartMs: 5000, EndMs: 5600, CleanText: "Next."},
	}
	chinese := []timedCue{
		{ID: 10, StartMs: 1100, EndMs: 1800, CleanText: "第一段"},
		{ID: 11, StartMs: 1850, EndMs: 2900, CleanText: "第二段"},
		{ID: 12, StartMs: 5700, EndMs: 6000, CleanText: "下一句"},
	}
	rows := alignCues(english, chinese)
	if got := rows[0].Chinese; got != "第一段 第二段" {
		t.Fatalf("merged Chinese = %q", got)
	}
	if len(rows[0].ZhCueIDs) != 2 {
		t.Fatalf("first matches = %#v", rows[0].ZhCueIDs)
	}
	if got := rows[1].Chinese; got != "下一句" {
		t.Fatalf("nearest fallback = %q", got)
	}
}

func TestAlignCuesAllowsManyEnglishToOneChinese(t *testing.T) {
	english := []timedCue{
		{ID: 1, Sequence: 0, StartMs: 1000, EndMs: 1800, CleanText: "First."},
		{ID: 2, Sequence: 1, StartMs: 1800, EndMs: 2600, CleanText: "Second."},
	}
	chinese := []timedCue{{ID: 10, StartMs: 1050, EndMs: 2550, CleanText: "合并翻译"}}
	rows := alignCues(english, chinese)
	if rows[0].Chinese != "合并翻译" || rows[1].Chinese != "合并翻译" {
		t.Fatalf("many-to-one rows = %#v", rows)
	}
}
