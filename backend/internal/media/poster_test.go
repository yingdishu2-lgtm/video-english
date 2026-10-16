package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/database"
)

type posterTestExtractor struct {
	mu        sync.Mutex
	calls     int
	requests  []FrameRequest
	deadlines []time.Time
}

func (e *posterTestExtractor) ExtractFrame(ctx context.Context, request FrameRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	e.requests = append(e.requests, request)
	if deadline, ok := ctx.Deadline(); ok {
		e.deadlines = append(e.deadlines, deadline)
	}
	return os.WriteFile(request.Output, []byte("jpeg-contents"), 0o644)
}

func (e *posterTestExtractor) callCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.calls
}

func TestPosterCachesDeduplicatesAndInvalidates(t *testing.T) {
	ctx := context.Background()
	mediaRoot := t.TempDir()
	videoPath := filepath.Join(mediaRoot, "lesson.mp4")
	if err := os.WriteFile(videoPath, []byte("source-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	client, err := database.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	extractor := &posterTestExtractor{}
	service := NewService(client, mediaRoot, fakePosterProber{}, WithPoster(t.TempDir(), extractor))
	if result, err := service.Scan(ctx); err != nil || result.Added != 1 {
		t.Fatalf("scan result = %#v, err = %v", result, err)
	}

	const callers = 8
	posters := make(chan Poster, callers)
	errors := make(chan error, callers)
	var group sync.WaitGroup
	for range callers {
		group.Add(1)
		go func() {
			defer group.Done()
			poster, err := service.Poster(ctx, 1)
			posters <- poster
			errors <- err
		}()
	}
	group.Wait()
	close(posters)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first Poster
	for poster := range posters {
		if first.Path == "" {
			first = poster
		}
		if poster.Path != first.Path || poster.ETag != first.ETag {
			t.Fatalf("concurrent poster mismatch: %#v vs %#v", poster, first)
		}
	}
	if extractor.callCount() != 1 {
		t.Fatalf("extractor calls = %d, want 1", extractor.callCount())
	}
	if info, err := os.Stat(first.Path); err != nil || info.Size() == 0 {
		t.Fatalf("cached poster invalid: info = %#v, err = %v", info, err)
	}
	if got := extractor.requests[0]; got.Width != 640 || got.Height != 360 || got.Seek != 6*time.Second {
		t.Fatalf("frame request = %#v", got)
	}
	if len(extractor.deadlines) != 1 {
		t.Fatalf("extractor deadlines = %d, want 1", len(extractor.deadlines))
	}
	remaining := time.Until(extractor.deadlines[0])
	if remaining < 14*time.Second || remaining > posterExtractionTimeout {
		t.Fatalf("extractor deadline remaining = %v, want about %v", remaining, posterExtractionTimeout)
	}

	changed := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(videoPath, changed, changed); err != nil {
		t.Fatal(err)
	}
	second, err := service.Poster(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if extractor.callCount() != 2 {
		t.Fatalf("extractor calls after source change = %d, want 2", extractor.callCount())
	}
	if second.Path == first.Path || second.ETag == first.ETag {
		t.Fatalf("poster did not invalidate: first = %#v, second = %#v", first, second)
	}
	if _, err := os.Stat(first.Path); !os.IsNotExist(err) {
		t.Fatalf("stale poster was not removed: %v", err)
	}
}

type fakePosterProber struct{}

func (fakePosterProber) Probe(context.Context, string) (ProbeResult, error) {
	width, height := 1920, 1080
	return ProbeResult{
		Container: "mov", DurationMs: 60_000,
		Streams: []Stream{{Index: 0, Type: "video", Codec: "h264", Width: &width, Height: &height}},
	}, nil
}

func TestPosterSeek(t *testing.T) {
	tests := []struct {
		durationMs int64
		want       time.Duration
	}{
		{durationMs: 0, want: 0},
		{durationMs: 500, want: 250 * time.Millisecond},
		{durationMs: 30_000, want: 3 * time.Second},
		{durationMs: 900_000, want: 60 * time.Second},
	}
	for _, test := range tests {
		if got := posterSeek(test.durationMs); got != test.want {
			t.Errorf("posterSeek(%d) = %v, want %v", test.durationMs, got, test.want)
		}
	}
}

type cancelTestExtractor struct {
	started chan struct{}
	stopped chan error
}

func (e cancelTestExtractor) ExtractFrame(ctx context.Context, _ FrameRequest) error {
	close(e.started)
	<-ctx.Done()
	e.stopped <- ctx.Err()
	return ctx.Err()
}

func TestPosterExtractionFollowsRequestCancellation(t *testing.T) {
	mediaRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(mediaRoot, "lesson.mp4"), []byte("source"), 0o644); err != nil {
		t.Fatal(err)
	}
	client, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	extractor := cancelTestExtractor{started: make(chan struct{}), stopped: make(chan error, 1)}
	service := NewService(client, mediaRoot, fakePosterProber{}, WithPoster(t.TempDir(), extractor))
	if result, err := service.Scan(context.Background()); err != nil || result.Added != 1 {
		t.Fatalf("scan result = %#v, err = %v", result, err)
	}

	requestCtx, cancelRequest := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := service.Poster(requestCtx, 1)
		result <- err
	}()
	<-extractor.started
	cancelRequest()

	select {
	case err := <-result:
		if !errors.Is(err, ErrPosterUnavailable) {
			t.Fatalf("Poster error = %v, want ErrPosterUnavailable", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Poster did not stop after request cancellation")
	}
	if err := <-extractor.stopped; !errors.Is(err, context.Canceled) {
		t.Fatalf("extractor context error = %v, want context.Canceled", err)
	}
}
