package playback

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
)

type Info struct {
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	URL         string `json:"url,omitempty"`
	AudioStream *int   `json:"audioStream,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type Service struct {
	media     *media.Service
	cacheRoot string
	remuxer   media.Remuxer
	mu        sync.Mutex
}

type Option func(*Service)

func WithRemux(cacheRoot string, remuxer media.Remuxer) Option {
	return func(service *Service) {
		service.cacheRoot = cacheRoot
		service.remuxer = remuxer
	}
}

func NewService(mediaService *media.Service, options ...Option) *Service {
	service := &Service{media: mediaService}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) Get(ctx context.Context, id int, requestedAudio *int) (Info, error) {
	item, err := s.media.Get(ctx, id)
	if err != nil {
		return Info{}, err
	}

	var defaultAudio *int
	for _, stream := range item.Streams {
		if stream.Type == "audio" && (stream.IsDefault || defaultAudio == nil) {
			value := stream.Index
			defaultAudio = &value
			if stream.IsDefault {
				break
			}
		}
	}
	selectedAudio := defaultAudio
	if requestedAudio != nil {
		selectedAudio = requestedAudio
	}
	var selectedStream *media.StreamItem
	if selectedAudio != nil {
		for index := range item.Streams {
			stream := &item.Streams[index]
			if stream.Type == "audio" && stream.Index == *selectedAudio {
				selectedStream = stream
				break
			}
		}
		if selectedStream == nil {
			return Info{}, fmt.Errorf("audio stream %d not found", *selectedAudio)
		}
	}

	extension := strings.ToLower(filepath.Ext(item.Filename))
	containerCompatible := extension == ".mp4" || extension == ".m4v" || extension == ".mov"
	videoCompatible := item.VideoCodec == "h264" || item.VideoCodec == "avc1"
	audioCompatible := selectedStream == nil || selectedStream.Codec == "aac" || selectedStream.Codec == "mp3"
	usesContainerDefault := requestedAudio == nil || (defaultAudio != nil && *requestedAudio == *defaultAudio)
	if containerCompatible && videoCompatible && audioCompatible && usesContainerDefault {
		streamURL := fmt.Sprintf("/api/v1/media/%d/stream", id)
		return Info{Mode: "direct", Status: "ready", URL: streamURL, AudioStream: selectedAudio}, nil
	}
	if videoCompatible && selectedAudio != nil && s.remuxer != nil && s.cacheRoot != "" {
		filename, err := s.ensureRemux(ctx, id, *selectedAudio, audioCompatible)
		if err != nil {
			return Info{}, err
		}
		playbackURL := fmt.Sprintf("/api/v1/media/%d/playback/%s", id, url.PathEscape(filename))
		return Info{Mode: "remux", Status: "ready", URL: playbackURL, AudioStream: selectedAudio}, nil
	}
	if !containerCompatible || !videoCompatible || !audioCompatible {
		return Info{Mode: "unsupported", Status: "unsupported", AudioStream: defaultAudio,
			Reason: fmt.Sprintf("direct play requires MP4/M4V/MOV with H.264 and AAC/MP3; found %s with %s", extension, item.VideoCodec)}, nil
	}
	return Info{Mode: "unsupported", Status: "unsupported", AudioStream: selectedAudio, Reason: "no compatible playback mode"}, nil
}

func (s *Service) ensureRemux(ctx context.Context, id, audioStream int, copyAudio bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	input, _, err := s.media.MediaPath(ctx, id)
	if err != nil {
		return "", err
	}
	inputInfo, err := os.Stat(input)
	if err != nil {
		return "", fmt.Errorf("stat remux input: %w", err)
	}
	directory := filepath.Join(s.cacheRoot, "media", strconv.Itoa(id))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("create playback cache: %w", err)
	}
	filename := fmt.Sprintf("audio-%d.mp4", audioStream)
	output := filepath.Join(directory, filename)
	if info, err := os.Stat(output); err == nil && info.Size() > 0 && !info.ModTime().Before(inputInfo.ModTime()) {
		return filename, nil
	}
	temporary, err := os.CreateTemp(directory, "remux-*.mp4")
	if err != nil {
		return "", fmt.Errorf("create remux temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return "", err
	}
	defer os.Remove(temporaryPath)
	if err := s.remuxer.Remux(ctx, media.RemuxRequest{Input: input, Output: temporaryPath, AudioStream: audioStream, CopyAudio: copyAudio}); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryPath, output); err != nil {
		return "", fmt.Errorf("publish remux output: %w", err)
	}
	return filename, nil
}

func (s *Service) DerivedPath(id int, filename string) (string, error) {
	if filename != filepath.Base(filename) || !strings.HasPrefix(filename, "audio-") || filepath.Ext(filename) != ".mp4" {
		return "", fmt.Errorf("invalid playback filename")
	}
	streamIndex := strings.TrimSuffix(strings.TrimPrefix(filename, "audio-"), ".mp4")
	if _, err := strconv.Atoi(streamIndex); err != nil {
		return "", fmt.Errorf("invalid playback filename")
	}
	path := filepath.Join(s.cacheRoot, "media", strconv.Itoa(id), filename)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", os.ErrNotExist
	}
	return path, nil
}
