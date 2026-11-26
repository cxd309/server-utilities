package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
)

// options are the command-line settings
type options struct {
	src    string // library root, e.g. /Volumes/audiobooks
	dest   string // mirror root, e.g. ~/audiobooks/mirror
	prune  bool
	dryRun bool
}

// run mirrors every book under o.src into o.dest, then prunes,
// writing one line per book it acts on to out.
// A book that fails is reported and counted, and the run carries on with the next one.
// It only stops early if ctx is cancelled
func run(ctx context.Context, o options, out io.Writer) (failed int, err error) {
	var books []book
	var stray []string
	if books, stray, err = scanBooks(o.src, o.dest); err != nil {
		return 0, err
	}
	for _, s := range stray {
		fmt.Fprintf(out, "warning: %s is not in an Author/Title folder, skipping\n", o.rel(s))
	}

	for _, b := range books {
		var action string
		action, err = buildOne(ctx, b, o.dryRun)
		if ctx.Err() != nil {
			return failed, ctx.Err()
		}
		switch {
		case err != nil:
			failed++
			fmt.Fprintf(out, "FAILED %s: %v\n", o.rel(b.src), err)
		case action != "":
			fmt.Fprintf(out, "%s: %s\n", o.rel(b.dst), action)
		}
	}

	if o.prune {
		var removed []string
		removed, err = prune(o.dest, books, o.dryRun)
		for _, p := range removed {
			fmt.Fprintf(out, "prune %s\n", o.rel(p))
		}
		if err != nil {
			return failed, err
		}
	}
	return failed, nil
}

// buildOne brings one book's mirror up to date.
// action says what it did, or would do with dryRun, and is "" when the book was already up to date
func buildOne(ctx context.Context, b book, dryRun bool) (action string, err error) {
	var done bool
	if done, err = upToDate(b); err != nil || done {
		return "", err
	}

	var paths []string
	if paths, err = listChapters(b.src); err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "skipped, no chapter files", nil
	}
	if dryRun {
		return "would build", nil
	}

	var d download
	if d, err = readDownload(ctx, b.src); err != nil {
		return "", err
	}
	var parts int
	if parts, err = buildBook(ctx, b, d, partLimit(outputBytesPerFrame(d))); err != nil {
		return "", err
	}
	return fmt.Sprintf("built %d part(s)", parts), nil
}

// rel shortens path for display, relative to the mirror or library root
func (o options) rel(path string) (short string) {
	for _, root := range []string{o.dest, o.src} {
		if r, relErr := filepath.Rel(root, path); relErr == nil && filepath.IsLocal(r) {
			return r
		}
	}
	return path
}
