package subtitle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"git.gdys01.cn/gdys01/English-media-learner/backend/ent"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/subtitlecue"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/subtitletrack"
	"git.gdys01.cn/gdys01/English-media-learner/backend/internal/media"
)

type Service struct {
	client    *ent.Client
	media     *media.Service
	cacheRoot string
	extractor media.SubtitleExtractor
	mu        sync.Mutex
}

type Track struct {
	ID           int    `json:"id"`
	MediaID      int    `json:"mediaId"`
	SourceType   string `json:"sourceType"`
	StreamIndex  *int   `json:"streamIndex,omitempty"`
	Format       string `json:"format"`
	Language     string `json:"language,omitempty"`
	Title        string `json:"title,omitempty"`
	ExternalPath string `json:"externalPath,omitempty"`
	OffsetMs     int64  `json:"offsetMs"`
	Supported    bool   `json:"supported"`
	Processed    bool   `json:"processed"`
}

type Cue struct {
	ID        int    `json:"id"`
	TrackID   int    `json:"trackId"`
	Sequence  int    `json:"sequence"`
	StartMs   int64  `json:"startMs"`
	EndMs     int64  `json:"endMs"`
	Text      string `json:"text"`
	CleanText string `json:"cleanText"`
}

type ProcessResult struct {
	TrackID int  `json:"trackId"`
	Cues    int  `json:"cues"`
	Cached  bool `json:"cached"`
}

func NewService(client *ent.Client, mediaService *media.Service, cacheRoot string, extractor media.SubtitleExtractor) *Service {
	return &Service{client: client, media: mediaService, cacheRoot: cacheRoot, extractor: extractor}
}

func (s *Service) ListTracks(ctx context.Context, mediaID int) ([]Track, error) {
	if err := s.syncTracks(ctx, mediaID); err != nil {
		return nil, err
	}
	entities, err := s.client.SubtitleTrack.Query().Where(subtitletrack.MediaItemIDEQ(mediaID)).All(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].SourceType != entities[j].SourceType {
			return entities[i].SourceType < entities[j].SourceType
		}
		return pointerValue(entities[i].StreamIndex) < pointerValue(entities[j].StreamIndex)
	})
	tracks := make([]Track, 0, len(entities))
	for _, entity := range entities {
		tracks = append(tracks, toTrack(entity))
	}
	return tracks, nil
}

func (s *Service) Process(ctx context.Context, trackID int) (ProcessResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	track, err := s.client.SubtitleTrack.Get(ctx, trackID)
	if err != nil {
		return ProcessResult{}, err
	}
	if !track.Supported {
		return ProcessResult{}, fmt.Errorf("subtitle format %s is not supported", track.Format)
	}
	input, format, fingerprint, err := s.trackInput(ctx, track)
	if err != nil {
		return ProcessResult{}, err
	}
	if track.Processed && track.SourceHash == fingerprint {
		count, err := s.client.SubtitleCue.Query().Where(subtitlecue.SubtitleTrackIDEQ(track.ID)).Count(ctx)
		if err == nil && count > 0 {
			return ProcessResult{TrackID: track.ID, Cues: count, Cached: true}, nil
		}
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("read subtitle: %w", err)
	}
	parsed, err := Parse(format, data)
	if err != nil {
		return ProcessResult{}, err
	}
	if len(parsed) == 0 {
		return ProcessResult{}, fmt.Errorf("subtitle contains no usable cues")
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return ProcessResult{}, err
	}
	rollback := func(cause error) (ProcessResult, error) {
		_ = tx.Rollback()
		return ProcessResult{}, cause
	}
	if _, err := tx.SubtitleCue.Delete().Where(subtitlecue.SubtitleTrackIDEQ(track.ID)).Exec(ctx); err != nil {
		return rollback(err)
	}
	builders := make([]*ent.SubtitleCueCreate, 0, len(parsed))
	for _, cue := range parsed {
		builders = append(builders, tx.SubtitleCue.Create().SetSubtitleTrackID(track.ID).SetSequence(cue.Sequence).
			SetStartMs(cue.StartMs).SetEndMs(cue.EndMs).SetText(cue.Text).SetCleanText(cue.CleanText))
	}
	if _, err := tx.SubtitleCue.CreateBulk(builders...).Save(ctx); err != nil {
		return rollback(err)
	}
	if _, err := tx.SubtitleTrack.UpdateOneID(track.ID).SetProcessed(true).SetSourceHash(fingerprint).Save(ctx); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return ProcessResult{}, err
	}
	return ProcessResult{TrackID: track.ID, Cues: len(parsed), Cached: false}, nil
}

func (s *Service) Cues(ctx context.Context, trackID int) ([]Cue, error) {
	if _, err := s.Process(ctx, trackID); err != nil {
		return nil, err
	}
	entities, err := s.client.SubtitleCue.Query().Where(subtitlecue.SubtitleTrackIDEQ(trackID)).Order(ent.Asc(subtitlecue.FieldSequence)).All(ctx)
	if err != nil {
		return nil, err
	}
	cues := make([]Cue, 0, len(entities))
	for _, entity := range entities {
		cues = append(cues, Cue{ID: entity.ID, TrackID: entity.SubtitleTrackID, Sequence: entity.Sequence,
			StartMs: entity.StartMs, EndMs: entity.EndMs, Text: entity.Text, CleanText: entity.CleanText})
	}
	return cues, nil
}

func (s *Service) syncTracks(ctx context.Context, mediaID int) error {
	item, err := s.media.Get(ctx, mediaID)
	if err != nil {
		return err
	}
	for _, stream := range item.Streams {
		if stream.Type != "subtitle" {
			continue
		}
		format, supported := subtitleFormat(stream.Codec)
		existing, err := s.client.SubtitleTrack.Query().Where(
			subtitletrack.MediaItemIDEQ(mediaID),
			subtitletrack.SourceTypeEQ("embedded"),
			subtitletrack.StreamIndexEQ(stream.Index),
		).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if existing == nil {
			if _, err := s.client.SubtitleTrack.Create().SetMediaItemID(mediaID).SetSourceType("embedded").SetStreamIndex(stream.Index).
				SetFormat(format).SetLanguage(stream.Language).SetTitle(stream.Title).SetSupported(supported).Save(ctx); err != nil {
				return err
			}
		} else if _, err := existing.Update().SetFormat(format).SetLanguage(stream.Language).SetTitle(stream.Title).SetSupported(supported).Save(ctx); err != nil {
			return err
		}
	}
	return s.syncExternalTracks(ctx, item)
}

func (s *Service) syncExternalTracks(ctx context.Context, item media.Item) error {
	videoPath, _, err := s.media.MediaPath(ctx, item.ID)
	if err != nil {
		return err
	}
	directory := filepath.Dir(videoPath)
	videoStem := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(entry.Name())), ".")
		if extension != "srt" && extension != "ass" && extension != "ssa" && extension != "vtt" {
			continue
		}
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if stem != videoStem && !strings.HasPrefix(stem, videoStem+".") {
			continue
		}
		absolute := filepath.Join(directory, entry.Name())
		relative, err := filepath.Rel(s.media.MediaRoot(), absolute)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		language := languageFromFilename(stem)
		existing, err := s.client.SubtitleTrack.Query().Where(subtitletrack.MediaItemIDEQ(item.ID), subtitletrack.ExternalPathEQ(relative)).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if existing == nil {
			if _, err := s.client.SubtitleTrack.Create().SetMediaItemID(item.ID).SetSourceType("external").SetFormat(extension).
				SetLanguage(language).SetTitle(entry.Name()).SetExternalPath(relative).SetSupported(true).Save(ctx); err != nil {
				return err
			}
		} else if _, err := existing.Update().SetFormat(extension).SetLanguage(language).SetTitle(entry.Name()).SetSupported(true).Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) trackInput(ctx context.Context, track *ent.SubtitleTrack) (string, string, string, error) {
	if track.SourceType == "external" {
		path, err := media.ResolveInside(s.media.MediaRoot(), track.ExternalPath)
		if err != nil {
			return "", "", "", err
		}
		info, err := os.Stat(path)
		if err != nil {
			return "", "", "", err
		}
		return path, track.Format, hashString(fmt.Sprintf("%s:%d:%d", track.ExternalPath, info.Size(), info.ModTime().UnixNano())), nil
	}
	if track.StreamIndex == nil {
		return "", "", "", fmt.Errorf("embedded subtitle has no stream index")
	}
	input, mediaEntity, err := s.media.MediaPath(ctx, track.MediaItemID)
	if err != nil {
		return "", "", "", err
	}
	directory := filepath.Join(s.cacheRoot, "subtitles", fmt.Sprint(track.MediaItemID))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", "", "", err
	}
	output := filepath.Join(directory, fmt.Sprintf("track-%d.srt", *track.StreamIndex))
	fingerprint := hashString(fmt.Sprintf("%s:%d", mediaEntity.ProbeHash, *track.StreamIndex))
	if track.SourceHash != fingerprint {
		if err := s.extractor.ExtractSubtitle(ctx, input, *track.StreamIndex, output); err != nil {
			return "", "", "", err
		}
	} else if _, err := os.Stat(output); err != nil {
		if err := s.extractor.ExtractSubtitle(ctx, input, *track.StreamIndex, output); err != nil {
			return "", "", "", err
		}
	}
	return output, "srt", fingerprint, nil
}

func subtitleFormat(codec string) (string, bool) {
	switch strings.ToLower(codec) {
	case "subrip", "srt", "mov_text":
		return "srt", true
	case "ass":
		return "ass", true
	case "ssa":
		return "ssa", true
	case "webvtt":
		return "vtt", true
	default:
		return strings.ToLower(codec), false
	}
}

func languageFromFilename(stem string) string {
	parts := strings.Split(strings.ToLower(stem), ".")
	if len(parts) < 2 {
		return ""
	}
	switch parts[len(parts)-1] {
	case "en", "eng":
		return "eng"
	case "zh", "zho", "chi", "chs", "cht":
		return "chi"
	default:
		return ""
	}
}

func toTrack(entity *ent.SubtitleTrack) Track {
	return Track{ID: entity.ID, MediaID: entity.MediaItemID, SourceType: entity.SourceType, StreamIndex: entity.StreamIndex,
		Format: entity.Format, Language: entity.Language, Title: entity.Title, ExternalPath: entity.ExternalPath,
		OffsetMs: entity.OffsetMs, Supported: entity.Supported, Processed: entity.Processed}
}

func pointerValue(value *int) int {
	if value == nil {
		return 1 << 30
	}
	return *value
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
