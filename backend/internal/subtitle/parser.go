package subtitle

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

type ParsedCue struct {
	Sequence  int
	StartMs   int64
	EndMs     int64
	Text      string
	CleanText string
}

var (
	timeLinePattern = regexp.MustCompile(`(?m)^\s*([^\s]+)\s*-->\s*([^\s]+)(?:\s+.*)?$`)
	htmlTagPattern  = regexp.MustCompile(`<[^>]+>`)
	assTagPattern   = regexp.MustCompile(`\{[^}]*\}`)
)

func Parse(format string, data []byte) ([]ParsedCue, error) {
	content := strings.TrimPrefix(string(data), "\ufeff")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	switch strings.ToLower(format) {
	case "srt", "vtt", "webvtt":
		return parseTimedText(content)
	case "ass", "ssa":
		return parseASS(content)
	default:
		return nil, fmt.Errorf("unsupported subtitle format %q", format)
	}
}

func parseTimedText(content string) ([]ParsedCue, error) {
	blocks := regexp.MustCompile(`\n{2,}`).Split(strings.TrimSpace(content), -1)
	cues := make([]ParsedCue, 0, len(blocks))
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		timingIndex := -1
		for index, line := range lines {
			if timeLinePattern.MatchString(line) {
				timingIndex = index
				break
			}
		}
		if timingIndex < 0 || timingIndex+1 >= len(lines) {
			continue
		}
		match := timeLinePattern.FindStringSubmatch(lines[timingIndex])
		start, err := parseClock(match[1])
		if err != nil {
			return nil, err
		}
		end, err := parseClock(match[2])
		if err != nil {
			return nil, err
		}
		text := strings.TrimSpace(strings.Join(lines[timingIndex+1:], "\n"))
		if text == "" || end <= start {
			continue
		}
		cues = append(cues, ParsedCue{Sequence: len(cues), StartMs: start, EndMs: end, Text: text, CleanText: cleanText(text)})
	}
	return cues, nil
}

func parseASS(content string) ([]ParsedCue, error) {
	lines := strings.Split(content, "\n")
	inEvents := false
	formatFields := []string{"layer", "start", "end", "style", "name", "marginl", "marginr", "marginv", "effect", "text"}
	var cues []ParsedCue
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, "[") {
			inEvents = strings.EqualFold(line, "[Events]")
			continue
		}
		if !inEvents {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "format:") {
			parts := strings.Split(strings.TrimSpace(line[len("format:"):]), ",")
			formatFields = formatFields[:0]
			for _, part := range parts {
				formatFields = append(formatFields, strings.ToLower(strings.TrimSpace(part)))
			}
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "dialogue:") {
			continue
		}
		values := strings.SplitN(strings.TrimSpace(line[len("dialogue:"):]), ",", len(formatFields))
		if len(values) != len(formatFields) {
			continue
		}
		fields := make(map[string]string, len(values))
		for index, name := range formatFields {
			fields[name] = strings.TrimSpace(values[index])
		}
		start, err := parseClock(fields["start"])
		if err != nil {
			return nil, err
		}
		end, err := parseClock(fields["end"])
		if err != nil {
			return nil, err
		}
		text := assTagPattern.ReplaceAllString(fields["text"], "")
		text = strings.ReplaceAll(text, `\N`, "\n")
		text = strings.ReplaceAll(text, `\n`, "\n")
		text = strings.ReplaceAll(text, `\h`, " ")
		text = strings.TrimSpace(text)
		if text == "" || end <= start {
			continue
		}
		cues = append(cues, ParsedCue{Sequence: len(cues), StartMs: start, EndMs: end, Text: text, CleanText: cleanText(text)})
	}
	return cues, nil
}

func parseClock(value string) (int64, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	parts := strings.Split(value, ":")
	if len(parts) != 2 && len(parts) != 3 {
		return 0, fmt.Errorf("invalid subtitle timestamp %q", value)
	}
	var hours int64
	if len(parts) == 3 {
		parsed, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid subtitle timestamp %q", value)
		}
		hours = parsed
		parts = parts[1:]
	}
	minutes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid subtitle timestamp %q", value)
	}
	secondsParts := strings.SplitN(parts[1], ".", 2)
	seconds, err := strconv.ParseInt(secondsParts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid subtitle timestamp %q", value)
	}
	milliseconds := int64(0)
	if len(secondsParts) == 2 {
		fraction := secondsParts[1]
		if len(fraction) > 3 {
			fraction = fraction[:3]
		}
		fraction += strings.Repeat("0", 3-len(fraction))
		milliseconds, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid subtitle timestamp %q", value)
		}
	}
	return ((hours*60+minutes)*60+seconds)*1000 + milliseconds, nil
}

func cleanText(value string) string {
	value = html.UnescapeString(htmlTagPattern.ReplaceAllString(value, ""))
	value = strings.Join(strings.Fields(value), " ")
	return strings.TrimSpace(value)
}
