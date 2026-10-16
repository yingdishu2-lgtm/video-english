package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type RemuxRequest struct {
	Input       string
	Output      string
	AudioStream int
	CopyAudio   bool
}

type Remuxer interface {
	Remux(ctx context.Context, request RemuxRequest) error
}

type SubtitleExtractor interface {
	ExtractSubtitle(ctx context.Context, input string, streamIndex int, output string) error
}

type FrameRequest struct {
	Input  string
	Output string
	Seek   time.Duration
	Width  int
	Height int
}

type FrameExtractor interface {
	ExtractFrame(ctx context.Context, request FrameRequest) error
}

type FFmpeg struct {
	Path string
}

func (f FFmpeg) Remux(ctx context.Context, request RemuxRequest) error {
	audioCodec := "aac"
	if request.CopyAudio {
		audioCodec = "copy"
	}
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", request.Input,
		"-map", "0:v:0",
		"-map", "0:" + strconv.Itoa(request.AudioStream),
		"-c:v", "copy",
		"-c:a", audioCodec,
		"-movflags", "+faststart",
		request.Output,
	}
	cmd := exec.CommandContext(ctx, f.Path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("ffmpeg remux failed: %s", message)
	}
	return nil
}

func (f FFmpeg) ExtractSubtitle(ctx context.Context, input string, streamIndex int, output string) error {
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", input,
		"-map", "0:" + strconv.Itoa(streamIndex),
		"-c:s", "srt",
		output,
	}
	cmd := exec.CommandContext(ctx, f.Path, args...)
	result, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(result))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("ffmpeg subtitle extraction failed: %s", message)
	}
	return nil
}

func (f FFmpeg) ExtractFrame(ctx context.Context, request FrameRequest) error {
	filter := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
		request.Width, request.Height, request.Width, request.Height,
	)
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-ss", strconv.FormatFloat(request.Seek.Seconds(), 'f', 3, 64),
		"-i", request.Input,
		"-map", "0:v:0",
		"-frames:v", "1",
		"-vf", filter,
		"-q:v", "3",
		request.Output,
	}
	cmd := exec.CommandContext(ctx, f.Path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("ffmpeg frame extraction failed: %s", message)
	}
	return nil
}
