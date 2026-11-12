package ffmpeg

import (
	"context"
	"encoding/json"
)

// Info is what ffprobe reports about a file.
// Only the fields the tools use are decoded.
// ffprobe prints times as decimal seconds in strings.
type Info struct {
	Format   Format    `json:"format"`
	Streams  []Stream  `json:"streams"`
	Chapters []Chapter `json:"chapters"`
}

// Format describes the file as a whole
type Format struct {
	Duration string            `json:"duration"`
	Tags     map[string]string `json:"tags"`
}

// Stream is one stream in the file:
// audio, cover art ("video") or the chapter text track ("data")
type Stream struct {
	CodecType  string `json:"codec_type"`
	CodecName  string `json:"codec_name"`
	Profile    string `json:"profile"`
	SampleRate string `json:"sample_rate"`
	Channels   int    `json:"channels"`
}

// Chapter is a chapter marker
type Chapter struct {
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Tags      map[string]string `json:"tags"`
}

// Probe asks ffprobe about the file at path
func Probe(ctx context.Context, path string) (info Info, err error) {
	var out []byte
	if out, err = run(
		ctx,
		"ffprobe",
		"-v", "error",
		"-of", "json",
		"-show_format",
		"-show_streams",
		"-show_format",
		"-show_streams",
		"-show_chapters",
		path); err != nil {
		return Info{}, err
	}
	if err = json.Unmarshal(out, &info); err != nil {
		return Info{}, err
	}
	return info, nil
}
