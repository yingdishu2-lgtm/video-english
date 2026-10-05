package dictionary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/database"
)

func TestLookupCachesIPA(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[{"word":"hello","phonetic":"/həˈləʊ/","phonetics":[{"audio":"//audio.example/hello.mp3"}]}]`))
	}))
	defer server.Close()
	client, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	service := NewService(client, WithEndpoint(server.URL+"/"))

	first, err := service.Lookup(context.Background(), "Hello")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if first.Phonetic != "/həˈləʊ/" || first.AudioURL != "https://audio.example/hello.mp3" || first.Cached {
		t.Fatalf("first result = %#v", first)
	}
	if !second.Cached || requests != 1 {
		t.Fatalf("second result = %#v, requests = %d", second, requests)
	}
}

func TestLookupRejectsNonWord(t *testing.T) {
	client, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := NewService(client).Lookup(context.Background(), "hello world"); err == nil {
		t.Fatal("expected invalid word error")
	}
}
