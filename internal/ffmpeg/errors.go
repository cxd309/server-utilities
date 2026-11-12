package ffmpeg

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotInstalled is returned when ffmpeg or ffprobe isn't on the PATH
var ErrNotInstalled = errors.New("not found on PATH")

// Error is returned when ffmpeg or ffprobe exists nonzero.
// Stderr holds the tool's own explanation, which is usually the useful part.
type Error struct {
	Tool   string // "ffmpeg" or "ffprobe"
	Stderr string
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s failed: %v: %s", e.Tool, e.Err, e.Stderr)
}

func (e *Error) Unwrap() error { return e.Err }

func returnErrNotInstalled(tool string) (err error) {
	err = fmt.Errorf("%s: %w", tool, ErrNotInstalled)
	return
}

func newError(tool string, stderr string, err error) (e error) {
	return &Error{Tool: tool, Stderr: strings.TrimSpace(stderr), Err: err}
}
