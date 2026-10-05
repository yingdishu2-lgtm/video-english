package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProbe(t *testing.T) {
	data := []byte(`{
		"streams": [
			{"index":0,"codec_name":"h264","codec_type":"video","width":1920,"height":1080,"tags":{"language":"und"},"disposition":{"default":1,"forced":0}},
			{"index":1,"codec_name":"aac","codec_type":"audio","channels":2,"sample_rate":"48000","tags":{"language":"eng","title":"English"},"disposition":{"default":1,"forced":0}}
		],
		"format":{"format_name":"mov,mp4,m4a,3gp,3g2,mj2","duration":"12.345000"}
	}`)
	result, err := parseProbe(data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Container != "mov" {
		t.Fatalf("container = %q", result.Container)
	}
	if result.DurationMs != 12345 {
		t.Fatalf("duration = %d", result.DurationMs)
	}
	if len(result.Streams) != 2 {
		t.Fatalf("streams = %d", len(result.Streams))
	}
	if result.Streams[1].Language != "eng" || result.Streams[1].SampleRate == nil || *result.Streams[1].SampleRate != 48000 {
		t.Fatalf("unexpected audio stream: %#v", result.Streams[1])
	}
}

func TestResolveInside(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "show"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "show", "episode.mp4"), []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveInside(root, "show/episode.mp4"); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveInside(root, "../secret.txt"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}
