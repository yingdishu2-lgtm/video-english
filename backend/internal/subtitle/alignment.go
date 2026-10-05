package subtitle

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"git.gdys01.cn/gdys01/English-media-learner/backend/ent"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/subtitlealignment"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/subtitlecue"
)

const alignmentVersion = 1

type LearningCue struct {
	ID       int    `json:"id"`
	Sequence int    `json:"sequence"`
	StartMs  int64  `json:"startMs"`
	EndMs    int64  `json:"endMs"`
	English  string `json:"en"`
	Chinese  string `json:"zh,omitempty"`
	EnCueIDs []int  `json:"enCueIds"`
	ZhCueIDs []int  `json:"zhCueIds"`
}

type timedCue struct {
	ID        int
	Sequence  int
	StartMs   int64
	EndMs     int64
	CleanText string
}

func (s *Service) LearningCues(ctx context.Context, mediaID, englishTrackID, chineseTrackID int) ([]LearningCue, error) {
	if englishTrackID == chineseTrackID {
		return nil, fmt.Errorf("english and chinese tracks must be different")
	}
	if _, err := s.Process(ctx, englishTrackID); err != nil {
		return nil, err
	}
	if _, err := s.Process(ctx, chineseTrackID); err != nil {
		return nil, err
	}
	enTrack, err := s.client.SubtitleTrack.Get(ctx, englishTrackID)
	if err != nil {
		return nil, err
	}
	zhTrack, err := s.client.SubtitleTrack.Get(ctx, chineseTrackID)
	if err != nil {
		return nil, err
	}
	if enTrack.MediaItemID != mediaID || zhTrack.MediaItemID != mediaID {
		return nil, fmt.Errorf("subtitle track does not belong to media")
	}

	enEntities, err := s.client.SubtitleCue.Query().Where(subtitlecue.SubtitleTrackIDEQ(englishTrackID)).Order(ent.Asc(subtitlecue.FieldSequence)).All(ctx)
	if err != nil {
		return nil, err
	}
	zhEntities, err := s.client.SubtitleCue.Query().Where(subtitlecue.SubtitleTrackIDEQ(chineseTrackID)).Order(ent.Asc(subtitlecue.FieldStartMs)).All(ctx)
	if err != nil {
		return nil, err
	}
	cacheKey := hashString(fmt.Sprintf("v%d:%s:%d:%s:%d", alignmentVersion, enTrack.SourceHash, enTrack.OffsetMs, zhTrack.SourceHash, zhTrack.OffsetMs))
	cached, err := s.client.SubtitleAlignment.Query().Where(
		subtitlealignment.EnglishTrackIDEQ(englishTrackID),
		subtitlealignment.ChineseTrackIDEQ(chineseTrackID),
	).Order(ent.Asc(subtitlealignment.FieldSequence)).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(cached) == len(enEntities) && allAlignmentKeys(cached, cacheKey) {
		return alignmentEntities(cached)
	}

	english := make([]timedCue, 0, len(enEntities))
	for _, cue := range enEntities {
		english = append(english, adjustedCue(cue.ID, cue.Sequence, cue.StartMs, cue.EndMs, cue.CleanText, enTrack.OffsetMs))
	}
	chinese := make([]timedCue, 0, len(zhEntities))
	for _, cue := range zhEntities {
		chinese = append(chinese, adjustedCue(cue.ID, cue.Sequence, cue.StartMs, cue.EndMs, cue.CleanText, zhTrack.OffsetMs))
	}
	rows := alignCues(english, chinese)

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	rollback := func(cause error) ([]LearningCue, error) {
		_ = tx.Rollback()
		return nil, cause
	}
	if _, err := tx.SubtitleAlignment.Delete().Where(
		subtitlealignment.EnglishTrackIDEQ(englishTrackID),
		subtitlealignment.ChineseTrackIDEQ(chineseTrackID),
	).Exec(ctx); err != nil {
		return rollback(err)
	}
	builders := make([]*ent.SubtitleAlignmentCreate, 0, len(rows))
	for _, row := range rows {
		enIDs, _ := json.Marshal(row.EnCueIDs)
		zhIDs, _ := json.Marshal(row.ZhCueIDs)
		builders = append(builders, tx.SubtitleAlignment.Create().
			SetMediaItemID(mediaID).SetEnglishTrackID(englishTrackID).SetChineseTrackID(chineseTrackID).
			SetSequence(row.Sequence).SetStartMs(row.StartMs).SetEndMs(row.EndMs).
			SetTextEn(row.English).SetTextZh(row.Chinese).SetEnCueIdsJSON(string(enIDs)).SetZhCueIdsJSON(string(zhIDs)).
			SetCacheKey(cacheKey).SetVersion(alignmentVersion))
	}
	if len(builders) > 0 {
		entities, err := tx.SubtitleAlignment.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return rollback(err)
		}
		for index, entity := range entities {
			rows[index].ID = entity.ID
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return rows, nil
}

func adjustedCue(id, sequence int, startMs, endMs int64, text string, offset int64) timedCue {
	startMs += offset
	endMs += offset
	if startMs < 0 {
		startMs = 0
	}
	if endMs <= startMs {
		endMs = startMs + 1
	}
	return timedCue{ID: id, Sequence: sequence, StartMs: startMs, EndMs: endMs, CleanText: text}
}

func alignCues(english, chinese []timedCue) []LearningCue {
	const minimumOverlap int64 = 150
	const maximumCenterDistance int64 = 1200
	rows := make([]LearningCue, 0, len(english))
	left := 0
	for _, en := range english {
		for left < len(chinese) && chinese[left].EndMs < en.StartMs-maximumCenterDistance {
			left++
		}
		matches := make([]timedCue, 0, 2)
		var nearest *timedCue
		nearestDistance := maximumCenterDistance + 1
		for index := left; index < len(chinese); index++ {
			zh := chinese[index]
			if zh.StartMs > en.EndMs+maximumCenterDistance {
				break
			}
			distance := absolute(center(en) - center(zh))
			if distance <= maximumCenterDistance && distance < nearestDistance {
				copy := zh
				nearest = &copy
				nearestDistance = distance
			}
			if overlap(en, zh) >= minimumOverlap {
				matches = append(matches, zh)
			}
		}
		if len(matches) == 0 && nearest != nil {
			matches = append(matches, *nearest)
		}
		zhIDs := make([]int, 0, len(matches))
		zhText := make([]string, 0, len(matches))
		for _, match := range matches {
			zhIDs = append(zhIDs, match.ID)
			if match.CleanText != "" && (len(zhText) == 0 || zhText[len(zhText)-1] != match.CleanText) {
				zhText = append(zhText, match.CleanText)
			}
		}
		rows = append(rows, LearningCue{Sequence: en.Sequence, StartMs: en.StartMs, EndMs: en.EndMs,
			English: en.CleanText, Chinese: strings.Join(zhText, " "), EnCueIDs: []int{en.ID}, ZhCueIDs: zhIDs})
	}
	return rows
}

func center(cue timedCue) int64 { return cue.StartMs + (cue.EndMs-cue.StartMs)/2 }

func overlap(a, b timedCue) int64 {
	start := max(a.StartMs, b.StartMs)
	end := min(a.EndMs, b.EndMs)
	if end <= start {
		return 0
	}
	return end - start
}

func absolute(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func allAlignmentKeys(rows []*ent.SubtitleAlignment, key string) bool {
	for _, row := range rows {
		if row.CacheKey != key || row.Version != alignmentVersion {
			return false
		}
	}
	return true
}

func alignmentEntities(entities []*ent.SubtitleAlignment) ([]LearningCue, error) {
	rows := make([]LearningCue, 0, len(entities))
	for _, entity := range entities {
		var enIDs, zhIDs []int
		if err := json.Unmarshal([]byte(entity.EnCueIdsJSON), &enIDs); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(entity.ZhCueIdsJSON), &zhIDs); err != nil {
			return nil, err
		}
		rows = append(rows, LearningCue{ID: entity.ID, Sequence: entity.Sequence, StartMs: entity.StartMs, EndMs: entity.EndMs,
			English: entity.TextEn, Chinese: entity.TextZh, EnCueIDs: enIDs, ZhCueIDs: zhIDs})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Sequence < rows[j].Sequence })
	return rows, nil
}
