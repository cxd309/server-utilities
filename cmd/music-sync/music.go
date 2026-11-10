package main

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// The AppleScripts that talk to the Music app, built into the binary.
var (
	//go:embed scripts/list.applescript
	listScript string
	//go:embed scripts/add.applescript
	addScript string
	//go:embed scripts/refresh.applescript
	refreshScript string
	//go:embed scripts/remove.applescript
	removeScript string
)

// batchSize is how many tracks are sent to the Music app per osascript call.
const batchSize = 100

// osascriptError is returned when osascript exits non-zero.
// Stderr holds the AppleScript error, e.g. a permissions problem.
type osascriptError struct {
	Stderr string
	Err    error
}

func (e *osascriptError) Error() string {
	return fmt.Sprintf("osascript failed: %v: %s", e.Err, e.Stderr)
}

func (e *osascriptError) Unwrap() error { return e.Err }

// listingError is returned when list.applescript prints a line we can't parse.
type listingError struct {
	Line string
}

func (e *listingError) Error() string {
	return fmt.Sprintf("unexpected line from the Music app: %q", e.Line)
}

// osascript runs script, passing args to its "on run argv" handler,
// and returns what the script returned.
func osascript(ctx context.Context, script string, args ...string) (out string, err error) {
	cmd := exec.CommandContext(ctx, "osascript", append([]string{"-e", script}, args...)...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err = cmd.Run(); err != nil {
		return "", &osascriptError{Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// listTracks asks the Music app for every file track in its library.
func listTracks(ctx context.Context) (tracks []libraryTrack, err error) {
	var out string
	if out, err = osascript(ctx, listScript); err != nil {
		return nil, err
	}
	return parseTracks(out)
}

// parseTracks reads list.applescript's output: one "id<tab>path" per line.
func parseTracks(out string) (tracks []libraryTrack, err error) {
	for line := range strings.Lines(out) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		id, path, ok := strings.Cut(line, "\t")
		if !ok || id == "" {
			return nil, &listingError{Line: line}
		}
		tracks = append(tracks, libraryTrack{id: id, path: path})
	}
	return tracks, nil
}

// inBatches runs script once per batch of items,
// calling onBatch (if not nil) with the running total after each batch.
func inBatches(ctx context.Context, script string, items []string, onBatch func(done int)) (err error) {
	done := 0
	for batch := range slices.Chunk(items, batchSize) {
		if _, err = osascript(ctx, script, batch...); err != nil {
			return err
		}
		done += len(batch)
		if onBatch != nil {
			onBatch(done)
		}
	}
	return nil
}
