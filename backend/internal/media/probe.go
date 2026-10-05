package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Stream struct {
	Index      int
	Type       string
	Codec      string
	Language   string
	Title      string
	Channels   *int
	SampleRate *int
	Width      *int
	Height     *int
	IsDefault  bool
	IsForced   bool
}

type ProbeResult struct {
	Container  string
	DurationMs int64
	Streams    []Stream
}

type Prober interface {
	Probe(ctx context.Context, path string) (ProbeResult, error)
}

type FFprobe struct {
	Path string
}

type rawProbe struct {
	Streams []struct {
		Index       int               `json:"index"`
		CodecName   string            `json:"codec_name"`
		CodecType   string            `json:"codec_type"`
		Width       int               `json:"width"`
		Height      int               `json:"height"`
		Channels    int               `json:"channels"`
		SampleRate  string            `json:"sample_rate"`
		Tags        map[string]string `json:"tags"`
		Disposition struct {
			Default int `json:"default"`
			Forced  int `json:"forced"`
		} `json:"disposition"`
	} `json:"streams"`
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
	} `json:"format"`
}

func (p FFprobe) Probe(ctx context.Context, path string) (ProbeResult, error) {
	cmd := exec.CommandContext(ctx, p.Path, "-v", "error", "-show_format", "-show_streams", "-of", "json", path)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return ProbeResult{}, fmt.Errorf("ffprobe failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return ProbeResult{}, fmt.Errorf("run ffprobe: %w", err)
	}
	return parseProbe(output)
}

func parseProbe(data []byte) (ProbeResult, error) {
	var raw rawProbe
	if err := json.Unmarshal(data, &raw); err != nil {
		return ProbeResult{}, fmt.Errorf("decode ffprobe output: %w", err)
	}
	duration, err := strconv.ParseFloat(raw.Format.Duration, 64)
	if err != nil {
		return ProbeResult{}, fmt.Errorf("parse duration %q: %w", raw.Format.Duration, err)
	}
	result := ProbeResult{
		Container:  strings.Split(raw.Format.FormatName, ",")[0],
		DurationMs: int64(duration*1000 + 0.5),
		Streams:    make([]Stream, 0, len(raw.Streams)),
	}
	for _, item := range raw.Streams {
		stream := Stream{
			Index: item.Index, Type: item.CodecType, Codec: item.CodecName,
			Language: item.Tags["language"], Title: item.Tags["title"],
			IsDefault: item.Disposition.Default == 1, IsForced: item.Disposition.Forced == 1,
		}
		if item.Width > 0 {
			stream.Width = &item.Width
		}
		if item.Height > 0 {
			stream.Height = &item.Height
		}
		if item.Channels > 0 {
			stream.Channels = &item.Channels
		}
		if value, err := strconv.Atoi(item.SampleRate); err == nil && value > 0 {
			stream.SampleRate = &value
		}
		result.Streams = append(result.Streams, stream)
	}
	return result, nil
}
