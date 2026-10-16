package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	posterWidth             = 640
	posterHeight            = 360
	posterExtractionTimeout = 15 * time.Second
)

var ErrPosterUnavailable = errors.New("poster unavailable")

type Poster struct {
	Path string
	ETag string
}

func (s *Service) Poster(ctx context.Context, id int) (Poster, error) {
	if s.frameExtractor == nil || s.posterCacheRoot == "" {
		return Poster{}, ErrPosterUnavailable
	}

	s.posterMu.Lock()
	defer s.posterMu.Unlock()

	input, entity, err := s.MediaPath(ctx, id)
	if err != nil {
		return Poster{}, err
	}
	inputInfo, err := os.Stat(input)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Poster{}, os.ErrNotExist
		}
		return Poster{}, fmt.Errorf("stat poster source: %w", err)
	}
	cacheKey := posterCacheKey(inputInfo.Size(), inputInfo.ModTime())
	directory := filepath.Join(s.posterCacheRoot, "media", strconv.Itoa(id))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Poster{}, fmt.Errorf("create poster cache: %w", err)
	}
	output := filepath.Join(directory, "poster-"+cacheKey+".jpg")
	if info, err := os.Stat(output); err == nil && !info.IsDir() && info.Size() > 0 {
		return Poster{Path: output, ETag: `"` + cacheKey + `"`}, nil
	}

	temporary, err := os.CreateTemp(directory, "poster-*.jpg")
	if err != nil {
		return Poster{}, fmt.Errorf("create poster temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return Poster{}, fmt.Errorf("close poster temporary file: %w", err)
	}
	defer os.Remove(temporaryPath)

	request := FrameRequest{
		Input: input, Output: temporaryPath, Seek: posterSeek(entity.DurationMs),
		Width: posterWidth, Height: posterHeight,
	}
	extractionCtx, cancelExtraction := context.WithTimeout(ctx, posterExtractionTimeout)
	defer cancelExtraction()
	if err := s.frameExtractor.ExtractFrame(extractionCtx, request); err != nil {
		return Poster{}, fmt.Errorf("%w: extraction failed", ErrPosterUnavailable)
	}
	info, err := os.Stat(temporaryPath)
	if err != nil || info.Size() == 0 {
		return Poster{}, fmt.Errorf("%w: empty extraction", ErrPosterUnavailable)
	}
	if err := os.Rename(temporaryPath, output); err != nil {
		return Poster{}, fmt.Errorf("publish poster: %w", err)
	}
	removeStalePosters(directory, filepath.Base(output))
	return Poster{Path: output, ETag: `"` + cacheKey + `"`}, nil
}

func posterCacheKey(size int64, modified time.Time) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d\x00%d\x00%dx%d", size, modified.UnixNano(), posterWidth, posterHeight)))
	return hex.EncodeToString(sum[:8])
}

func posterSeek(durationMs int64) time.Duration {
	if durationMs <= 0 {
		return 0
	}
	seek := time.Duration(durationMs/10) * time.Millisecond
	if seek < time.Second {
		return time.Duration(durationMs/2) * time.Millisecond
	}
	if seek > 60*time.Second {
		return 60 * time.Second
	}
	return seek
}

func removeStalePosters(directory, keep string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == keep || !strings.HasPrefix(name, "poster-") || filepath.Ext(name) != ".jpg" {
			continue
		}
		_ = os.Remove(filepath.Join(directory, name))
	}
}
