package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git.gdys01.cn/gdys01/English-media-learner/backend/ent"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/mediaitem"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/mediastream"
)

type Service struct {
	client    *ent.Client
	mediaRoot string
	prober    Prober
}

type ScanResult struct{ Added, Updated, Unchanged, Failed int }

type Item struct {
	ID         int          `json:"id"`
	Path       string       `json:"path"`
	Filename   string       `json:"filename"`
	Title      string       `json:"title"`
	Container  string       `json:"container"`
	DurationMs int64        `json:"durationMs"`
	VideoCodec string       `json:"videoCodec"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	FileSize   int64        `json:"fileSize"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	Streams    []StreamItem `json:"streams"`
}

type StreamItem struct {
	ID         int    `json:"id"`
	Index      int    `json:"index"`
	Type       string `json:"type"`
	Codec      string `json:"codec"`
	Language   string `json:"language,omitempty"`
	Title      string `json:"title,omitempty"`
	Channels   *int   `json:"channels,omitempty"`
	SampleRate *int   `json:"sampleRate,omitempty"`
	Width      *int   `json:"width,omitempty"`
	Height     *int   `json:"height,omitempty"`
	IsDefault  bool   `json:"isDefault"`
	IsForced   bool   `json:"isForced"`
}

func NewService(client *ent.Client, mediaRoot string, prober Prober) *Service {
	return &Service{client: client, mediaRoot: mediaRoot, prober: prober}
}

func (s *Service) MediaRoot() string { return s.mediaRoot }

func (s *Service) Scan(ctx context.Context) (ScanResult, error) {
	var result ScanResult
	err := filepath.WalkDir(s.mediaRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isSupportedVideo(path) {
			return nil
		}
		relative, err := filepath.Rel(s.mediaRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fingerprint := fingerprint(relative, info.Size(), info.ModTime())
		existing, err := s.client.MediaItem.Query().Where(mediaitem.PathEQ(relative)).Only(ctx)
		if err == nil && existing.ProbeHash == fingerprint {
			result.Unchanged++
			return nil
		}
		if err != nil && !ent.IsNotFound(err) {
			return err
		}

		probe, err := s.prober.Probe(ctx, path)
		if err != nil {
			result.Failed++
			return nil
		}
		if existing == nil {
			if _, err := s.createItem(ctx, relative, info, fingerprint, probe); err != nil {
				return err
			}
			result.Added++
		} else {
			if _, err := s.updateItem(ctx, existing, info, fingerprint, probe); err != nil {
				return err
			}
			result.Updated++
		}
		return nil
	})
	return result, err
}

func (s *Service) List(ctx context.Context) ([]Item, error) {
	entities, err := s.client.MediaItem.Query().Order(ent.Asc(mediaitem.FieldPath)).All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(entities))
	for _, entity := range entities {
		item, err := s.toItem(ctx, entity)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, id int) (Item, error) {
	entity, err := s.client.MediaItem.Get(ctx, id)
	if err != nil {
		return Item{}, err
	}
	return s.toItem(ctx, entity)
}

func (s *Service) MediaPath(ctx context.Context, id int) (string, *ent.MediaItem, error) {
	entity, err := s.client.MediaItem.Get(ctx, id)
	if err != nil {
		return "", nil, err
	}
	path, err := ResolveInside(s.mediaRoot, entity.Path)
	return path, entity, err
}

func (s *Service) Reprobe(ctx context.Context, id int) (Item, error) {
	path, entity, err := s.MediaPath(ctx, id)
	if err != nil {
		return Item{}, err
	}
	return s.reprobePath(ctx, entity, path)
}

func (s *Service) reprobePath(ctx context.Context, entity *ent.MediaItem, path string) (Item, error) {
	info, err := statFile(path)
	if err != nil {
		return Item{}, err
	}
	probe, err := s.prober.Probe(ctx, path)
	if err != nil {
		return Item{}, err
	}
	updated, err := s.updateItem(ctx, entity, info, fingerprint(entity.Path, info.Size(), info.ModTime()), probe)
	if err != nil {
		return Item{}, err
	}
	return s.toItem(ctx, updated)
}

func (s *Service) createItem(ctx context.Context, relative string, info fs.FileInfo, hash string, probe ProbeResult) (*ent.MediaItem, error) {
	video := primaryVideo(probe.Streams)
	entity, err := s.client.MediaItem.Create().
		SetPath(relative).SetFilename(filepath.Base(relative)).SetTitle(strings.TrimSuffix(filepath.Base(relative), filepath.Ext(relative))).
		SetContainer(probe.Container).SetDurationMs(probe.DurationMs).SetVideoCodec(video.Codec).
		SetWidth(valueOrZero(video.Width)).SetHeight(valueOrZero(video.Height)).SetFileSize(info.Size()).SetFileMtime(info.ModTime()).SetProbeHash(hash).Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.replaceStreams(ctx, entity.ID, probe.Streams); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) updateItem(ctx context.Context, entity *ent.MediaItem, info fs.FileInfo, hash string, probe ProbeResult) (*ent.MediaItem, error) {
	video := primaryVideo(probe.Streams)
	updated, err := entity.Update().SetContainer(probe.Container).SetDurationMs(probe.DurationMs).SetVideoCodec(video.Codec).
		SetWidth(valueOrZero(video.Width)).SetHeight(valueOrZero(video.Height)).SetFileSize(info.Size()).SetFileMtime(info.ModTime()).SetProbeHash(hash).Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.replaceStreams(ctx, entity.ID, probe.Streams); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) replaceStreams(ctx context.Context, mediaID int, streams []Stream) error {
	if _, err := s.client.MediaStream.Delete().Where(mediastream.MediaItemIDEQ(mediaID)).Exec(ctx); err != nil {
		return err
	}
	for _, stream := range streams {
		create := s.client.MediaStream.Create().SetMediaItemID(mediaID).SetStreamIndex(stream.Index).SetStreamType(stream.Type).
			SetCodec(stream.Codec).SetLanguage(stream.Language).SetTitle(stream.Title).SetIsDefault(stream.IsDefault).SetIsForced(stream.IsForced)
		if stream.Channels != nil {
			create.SetChannels(*stream.Channels)
		}
		if stream.SampleRate != nil {
			create.SetSampleRate(*stream.SampleRate)
		}
		if stream.Width != nil {
			create.SetWidth(*stream.Width)
		}
		if stream.Height != nil {
			create.SetHeight(*stream.Height)
		}
		if _, err := create.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) toItem(ctx context.Context, entity *ent.MediaItem) (Item, error) {
	streams, err := s.client.MediaStream.Query().Where(mediastream.MediaItemIDEQ(entity.ID)).All(ctx)
	if err != nil {
		return Item{}, err
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].StreamIndex < streams[j].StreamIndex })
	result := Item{ID: entity.ID, Path: entity.Path, Filename: entity.Filename, Title: entity.Title, Container: entity.Container,
		DurationMs: entity.DurationMs, VideoCodec: entity.VideoCodec, Width: entity.Width, Height: entity.Height,
		FileSize: entity.FileSize, UpdatedAt: entity.UpdatedAt, Streams: make([]StreamItem, 0, len(streams))}
	for _, stream := range streams {
		result.Streams = append(result.Streams, StreamItem{ID: stream.ID, Index: stream.StreamIndex, Type: stream.StreamType, Codec: stream.Codec,
			Language: stream.Language, Title: stream.Title, Channels: stream.Channels, SampleRate: stream.SampleRate,
			Width: stream.Width, Height: stream.Height, IsDefault: stream.IsDefault, IsForced: stream.IsForced})
	}
	return result, nil
}

func isSupportedVideo(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mkv", ".mp4", ".m4v", ".mov", ".webm":
		return true
	default:
		return false
	}
}

func fingerprint(relative string, size int64, mtime time.Time) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%d", relative, size, mtime.UnixNano())))
	return hex.EncodeToString(sum[:])
}

func primaryVideo(streams []Stream) Stream {
	for _, stream := range streams {
		if stream.Type == "video" {
			return stream
		}
	}
	return Stream{}
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
