package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/database"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/playback"
	"github.com/gin-gonic/gin"
)

type fakeProber struct{}

func (fakeProber) Probe(context.Context, string) (media.ProbeResult, error) {
	width, height, channels, sampleRate := 1280, 720, 2, 48000
	return media.ProbeResult{
		Container: "mov", DurationMs: 5000,
		Streams: []media.Stream{
			{Index: 0, Type: "video", Codec: "h264", Width: &width, Height: &height, IsDefault: true},
			{Index: 1, Type: "audio", Codec: "aac", Language: "eng", Channels: &channels, SampleRate: &sampleRate, IsDefault: true},
		},
	}, nil
}

func TestScanPlaybackAndRangeStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	video := []byte("0123456789abcdef")
	if err := os.WriteFile(filepath.Join(root, "lesson.mp4"), video, 0o644); err != nil {
		t.Fatal(err)
	}
	client, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	mediaService := media.NewService(client, root, fakeProber{})
	router := NewRouter(mediaService, playback.NewService(mediaService), nil, nil, "")

	scan := httptest.NewRecorder()
	router.ServeHTTP(scan, httptest.NewRequest(http.MethodPost, "/api/v1/media/scan", nil))
	if scan.Code != http.StatusOK {
		t.Fatalf("scan status = %d, body = %s", scan.Code, scan.Body.String())
	}
	var scanBody map[string]int
	if err := json.Unmarshal(scan.Body.Bytes(), &scanBody); err != nil {
		t.Fatal(err)
	}
	if scanBody["added"] != 1 {
		t.Fatalf("scan result = %#v", scanBody)
	}

	play := httptest.NewRecorder()
	router.ServeHTTP(play, httptest.NewRequest(http.MethodGet, "/api/v1/media/1/playback", nil))
	if play.Code != http.StatusOK {
		t.Fatalf("playback status = %d, body = %s", play.Code, play.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/media/1/stream", nil)
	request.Header.Set("Range", "bytes=3-7")
	stream := httptest.NewRecorder()
	router.ServeHTTP(stream, request)
	if stream.Code != http.StatusPartialContent {
		t.Fatalf("stream status = %d", stream.Code)
	}
	if stream.Body.String() != "34567" {
		t.Fatalf("stream body = %q", stream.Body.String())
	}
	if value := stream.Header().Get("Content-Range"); value != "bytes 3-7/16" {
		t.Fatalf("Content-Range = %q", value)
	}
}

func TestServesFrontendAndSPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("<main>English Media Learner</main>"), 0o644); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(nil, nil, nil, nil, dist)

	for _, route := range []string{"/", "/watch/42"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
		if response.Code != http.StatusOK || response.Body.String() != "<main>English Media Learner</main>" {
			t.Fatalf("route %s: status = %d, body = %q", route, response.Code, response.Body.String())
		}
	}

	apiResponse := httptest.NewRecorder()
	router.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil))
	if apiResponse.Code != http.StatusNotFound {
		t.Fatalf("unknown API status = %d", apiResponse.Code)
	}
}
