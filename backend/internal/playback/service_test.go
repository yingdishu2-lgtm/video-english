package playback

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/database"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
)

type testProber struct{}

func (testProber) Probe(context.Context, string) (media.ProbeResult, error) {
	width, height, channels := 1920, 1080, 2
	return media.ProbeResult{
		Container: "matroska", DurationMs: 60_000,
		Streams: []media.Stream{
			{Index: 0, Type: "video", Codec: "h264", Width: &width, Height: &height},
			{Index: 1, Type: "audio", Codec: "aac", Language: "eng", Channels: &channels},
		},
	}, nil
}

type testRemuxer struct{ calls int }

func (r *testRemuxer) Remux(_ context.Context, request media.RemuxRequest) error {
	r.calls++
	return os.WriteFile(request.Output, []byte("derived-mp4"), 0o644)
}

func TestMKVUsesAndReusesRemuxCache(t *testing.T) {
	ctx := context.Background()
	mediaRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(mediaRoot, "episode.mkv"), []byte("source"), 0o644); err != nil {
		t.Fatal(err)
	}
	client, err := database.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	mediaService := media.NewService(client, mediaRoot, testProber{})
	if result, err := mediaService.Scan(ctx); err != nil || result.Added != 1 {
		t.Fatalf("scan result = %#v, err = %v", result, err)
	}

	remuxer := &testRemuxer{}
	service := NewService(mediaService, WithRemux(t.TempDir(), remuxer))
	first, err := service.Get(ctx, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Mode != "remux" || first.Status != "ready" || first.URL != "/api/v1/media/1/playback/audio-1.mp4" {
		t.Fatalf("unexpected playback info: %#v", first)
	}
	if _, err := service.DerivedPath(1, "audio-1.mp4"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(ctx, 1, nil); err != nil {
		t.Fatal(err)
	}
	if remuxer.calls != 1 {
		t.Fatalf("remux calls = %d, want 1", remuxer.calls)
	}
}
