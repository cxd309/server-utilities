package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// tagArgs builds the ffmpeg command line that copies piece to out with tags for chapter n of b.
// The audio is copied, not re-encoded
func tagArgs(piece, out string, b book, n int) (args []string) {
	tags := []struct{ key, value string }{
		{"title", b.chapters[n-1].title},
		{"album", b.title},
		{"artist", b.author},
		{"album_artist", b.author},
		{"composer", b.narrator},
		{"date", b.date},
		{"track", fmt.Sprintf("%d/%d", n, len(b.chapters))},
	}

	args = []string{"-i", piece, "-c", "copy", "-map_metadata", "-1"}
	for _, t := range tags {
		if t.value != "" {
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", t.key, t.value))
		}
	}
	return append(args, "-f", "ipod", out)
}

// tagPiece tags the nth untagged piece in dir, gives it its proper name
// and removes the untagged copy
func tagPiece(ctx context.Context, dir string, b book, n int) (err error) {
	piece := filepath.Join(dir, fmt.Sprintf(pieceName, n-1)) // pieces count from 0
	out := filepath.Join(dir, chapterFileName(n, len(b.chapters), b.chapters[n-1].title))
	if err = ffmpeg.Run(ctx, tagArgs(piece, out, b, n)...); err != nil {
		return err
	}
	return os.Remove(piece)
}
