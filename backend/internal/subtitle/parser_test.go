package subtitle

import "testing"

func TestParseSRT(t *testing.T) {
	input := []byte("1\r\n00:00:01,250 --> 00:00:03,500\r\n<i>Hello</i>, world!\r\n\r\n2\r\n00:01:02.000 --> 00:01:04.125\r\nSecond line\r\n")
	cues, err := Parse("srt", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 {
		t.Fatalf("cue count = %d", len(cues))
	}
	if cues[0].StartMs != 1250 || cues[0].EndMs != 3500 || cues[0].CleanText != "Hello, world!" {
		t.Fatalf("unexpected first cue: %#v", cues[0])
	}
	if cues[1].StartMs != 62000 || cues[1].EndMs != 64125 {
		t.Fatalf("unexpected second cue: %#v", cues[1])
	}
}

func TestParseASS(t *testing.T) {
	input := []byte(`[Script Info]
Title: Example

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:02.10,0:00:04.25,Default,,0,0,0,,{\i1}Hello{\i0}\Nworld
`)
	cues, err := Parse("ass", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 1 || cues[0].StartMs != 2100 || cues[0].EndMs != 4250 || cues[0].CleanText != "Hello world" {
		t.Fatalf("unexpected cues: %#v", cues)
	}
}
