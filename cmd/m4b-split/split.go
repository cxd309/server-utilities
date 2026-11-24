package main

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// pieceName is the pattern ffmpeg numbers the untagged pieces with,
// starting from 0000.m4a.
// The next step renames and tags them
const pieceName = "%04d.m4a"

// segmentArgs builds the ffmpeg command line that cuts src into one file per chapter in dir,
// copying the audio without re-encoding
func segmentArgs(src, dir string, chapters []chapter) (args []string) {
	args = []string{
		"-i", src,
		"-map", "0:a:0", // first audio stream only: no cover art or chapter track
		"-c", "copy", // no re-encoding
		"-map_chapters", "-1", // each piece gets its own tags later
		"-map_metadata", "-1",
		"-f", "segment",
		"-segment_format", "ipod",
		"-reset_timestamps", "1", // each piece starts at 0:00
	}

	// Cut at the start of every chapter except the first
	var times []string
	for _, c := range chapters[1:] {
		times = append(times, strconv.FormatFloat(c.start, 'f', -1, 64))
	}
	if len(times) > 0 {
		args = append(args, "-segment_times", strings.Join(times, ","))
	} else {
		// One chapter: no cuts. Given no rule at all, the segment muxer cuts every 2 s
		args = append(args, "-segment_time", "1000000000")
	}

	return append(args, filepath.Join(dir, pieceName))
}

// splitAudio cuts src into one untagged file per chapter in dir
func splitAudio(ctx context.Context, src, dir string, chapters []chapter) (err error) {
	return ffmpeg.Run(ctx, segmentArgs(src, dir, chapters)...)
}
