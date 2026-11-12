package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

// Run runs ffmpeg with args.
// It never reads the keyboard, overwrites existing outputs and prints only errors
func Run(ctx context.Context, args ...string) (err error) {
	base := []string{
		"-nostdin",
		"-hide_banner",
		"-loglevel",
		"error",
		"-y",
	}
	_, err = run(ctx, "ffmpeg", append(base, args...)...)
	return err
}

// run runs tool and returns what it printed to stdout
func run(ctx context.Context, tool string, args ...string) (stdout []byte, err error) {
	var out, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err = cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, returnErrNotInstalled(tool)
		}
		return nil, newError(tool, stderr.String(), err)
	}
	return out.Bytes(), nil
}
