// Package podcast finishes the tagging of podcast episodes that ffmpeg can't do.
package podcast

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"os/exec"
	"strings"
)

//go:embed flag.py
var flagScript string

var (
	ErrNoPython  = errors.New("python3 not found on PATH")
	ErrNoMutagen = errors.New("the Python mutagen module is missing: sudo apt install python3-mutagen")
)

// Error is returned when the helper fails for any other reason, with Python's message
type Error struct {
	Stderr string
	Err    error
}

func (e *Error) Error() string {
	if e.Stderr != "" {
		return "podcast flag: " + e.Stderr
	}
	return "podcast flag: " + e.Err.Error()
}

func (e *Error) Unwrap() error { return e.Err }

// Flag marks each MP4 file in paths as a podcast, editing its tags in place.
// The file's media kind should already be Podcast (ffmpeg's media_type=21)
func Flag(ctx context.Context, paths ...string) (err error) {
	if len(paths) == 0 {
		return nil
	}
	cmd := exec.CommandContext(ctx, "python3", append([]string{"-c", flagScript}, paths...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return ErrNoPython
	}
	msg := strings.TrimSpace(stderr.String())
	if strings.Contains(msg, "No module named 'mutagen'") {
		return ErrNoMutagen
	}
	return &Error{Stderr: msg, Err: err}
}
