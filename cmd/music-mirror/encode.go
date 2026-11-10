package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	errNoFFmpeg = errors.New("ffmpeg not found on PATH (sudo apt install ffmpeg)")
	errNoLame   = errors.New("ffmpeg was built without the libmp3lame encoder")
)

// ffmpegError is returned when ffmpeg exits non-zero.
// Stderr holds ffmpeg's own explanation, which is usually the useful part
type ffmpegError struct {
	Src    string
	Stderr string
	Err    error
}

func (e *ffmpegError) Error() string {
	return fmt.Sprintf("ffmpeg failed on %s: %v: %s", e.Src, e.Err, e.Stderr)
}

func (e *ffmpegError) Unwrap() error { return e.Err }

// checkFFmpeg confirms ffmpeg is installed and can encode mp3
func checkFFmpeg() (err error) {
	var out []byte
	if out, err = exec.Command("ffmpeg", "-hide_banner", "-encoders").Output(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errNoFFmpeg
		}
		return err
	}
	if !bytes.Contains(out, []byte("libmp3lame")) {
		return errNoLame
	}
	return nil
}

// ffmpeg builds the command line that turns one FLAC into an untagged CBR mp3.
// Tags are written seperately afterwards
func ffmpegArgs(src, out, bitrate string) (args []string) {
	return []string{
		"-nostdin", // never wait for keyboard input
		"-hide_banner",
		"-loglevel", "error", // only print problems
		"-y", // overwrite out if a previous run left one
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
	var stderr strings.Builder
	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs(src, out, bitrate)...)
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return &ffmpegError{Src: src, Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return nil
}
