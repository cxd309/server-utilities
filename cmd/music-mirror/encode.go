package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

var (
	errNoLame = errors.New("ffmpeg was built without the libmp3lame encoder")
)

// checkFFmpeg confirms ffmpeg is installed and can encode mp3
func checkFFmpeg() (err error) {
	var out []byte
	if out, err = exec.Command("ffmpeg", "-hide_banner", "-encoders").Output(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return ffmpeg.ErrNotInstalled
		}
		return err
	}
	if !bytes.Contains(out, []byte("libmp3lame")) {
		return errNoLame
	}
	return nil
}

// ffmpegArgs builds the command line that turns one FLAC into an untagged CBR mp3.
// Tags are written separately afterwards
func ffmpegArgs(src, out, bitrate string) (args []string) {
	return []string{
		"-i", src,
		"-map", "0:a:0", // first audio stream only, which drops embedded cover art
		"-map_metadata", "-1", // copy no tags across
		"-c:a", "libmp3lame",
		"-b:a", bitrate,
		"-compression_level", "2", // LAME -q 2: high quality at a sensible speed
		"-id3v2_version", "0", // write no ID3 tag of our own...
		"-write_id3v1", "0", // ...or an old-style v1 one
		"-f", "mp3", // out will end in .part, so name the format explicitly
		out,
	}
}

// encode converts src to an untagged mp3 at out
func encode(ctx context.Context, src, out, bitrate string) (err error) {
	return ffmpeg.Run(ctx, ffmpegArgs(src, out, bitrate)...)
}
