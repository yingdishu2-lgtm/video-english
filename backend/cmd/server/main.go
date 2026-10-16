package main

import (
	"context"
	"log"
	"path/filepath"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/api"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/config"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/database"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/dictionary"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/playback"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/subtitle"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatal(err)
	}
	client, err := database.Open(context.Background(), cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	cacheRoot := filepath.Join(cfg.DataRoot, "cache")
	ffmpeg := media.FFmpeg{Path: cfg.FFmpegPath}
	mediaService := media.NewService(client, cfg.MediaRoot, media.FFprobe{Path: cfg.FFprobePath}, media.WithPoster(cacheRoot, ffmpeg))
	playbackService := playback.NewService(mediaService, playback.WithRemux(cacheRoot, ffmpeg))
	subtitleService := subtitle.NewService(client, mediaService, cacheRoot, ffmpeg)
	dictionaryService := dictionary.NewService(client)
	router := api.NewRouter(mediaService, playbackService, subtitleService, dictionaryService, cfg.FrontendDist)
	log.Printf("English Media Learner API listening on http://localhost:%s (media: %s)", cfg.Port, cfg.MediaRoot)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
