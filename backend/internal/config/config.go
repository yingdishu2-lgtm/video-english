package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Port         string
	MediaRoot    string
	DataRoot     string
	Database     string
	FrontendDist string
	FFmpegPath   string
	FFprobePath  string
}

func Load() (Config, error) {
	mediaRoot := envOr("MEDIA_ROOT", "../media")
	dataRoot := envOr("DATA_ROOT", "../data")
	database := envOr("DATABASE_PATH", filepath.Join(dataRoot, "app.db"))
	frontendDist := envOr("FRONTEND_DIST", "../frontend/dist")

	var err error
	mediaRoot, err = filepath.Abs(mediaRoot)
	if err != nil {
		return Config{}, fmt.Errorf("resolve media root: %w", err)
	}
	dataRoot, err = filepath.Abs(dataRoot)
	if err != nil {
		return Config{}, fmt.Errorf("resolve data root: %w", err)
	}
	database, err = filepath.Abs(database)
	if err != nil {
		return Config{}, fmt.Errorf("resolve database path: %w", err)
	}
	frontendDist, err = filepath.Abs(frontendDist)
	if err != nil {
		return Config{}, fmt.Errorf("resolve frontend dist: %w", err)
	}

	return Config{
		Port:         envOr("APP_PORT", "8080"),
		MediaRoot:    mediaRoot,
		DataRoot:     dataRoot,
		Database:     database,
		FrontendDist: frontendDist,
		FFmpegPath:   envOr("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:  envOr("FFPROBE_PATH", "ffprobe"),
	}, nil
}

func (c Config) EnsureDirectories() error {
	if err := os.MkdirAll(c.MediaRoot, 0o755); err != nil {
		return fmt.Errorf("create media root: %w", err)
	}
	if err := os.MkdirAll(c.DataRoot, 0o755); err != nil {
		return fmt.Errorf("create data root: %w", err)
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
