package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// buildDir is the hidden folder a book is built in, beside where it ends up.
// Being hidden keeps it out of scans and the Books sync while it's incomplete
func buildDir(dst string) (dir string) {
	dir = filepath.Join(filepath.Dir(dst), "."+filepath.Base(dst)+".building")
	return dir
}

// buildBook plans a downloaded book into parts of at most limit frames,
// builds them all in the hidden build folder, then swaps that folder into place at b.dst.
// It returns how many parts it wrote
func buildBook(ctx context.Context, b book, d download, limit int64) (parts int, err error) {
	frames := make([]int64, len(d.chapters))
	for i, c := range d.chapters {
		frames[i] = c.frames
	}
	var plan [][]piece
	if plan, err = planParts(frames, limit); err != nil {
		return 0, err
	}

	tmp := buildDir(b.dst)
	if err = os.RemoveAll(tmp); err != nil { // left over from an interrupted run
		return 0, err
	}
	if err = os.MkdirAll(tmp, 0o755); err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tmp)
		}
	}()

	for i, part := range plan {
		n := i + 1
		out := filepath.Join(tmp, fmt.Sprintf("Part %02d.m4b", n))
		tags := tagLines(d.tags, n, len(plan))
		if err = joinPart(ctx, part, d.chapters, d.sampleRate, tags, d.cover, out); err != nil {
			return 0, err
		}
	}
	if err = swapDir(tmp, b.dst); err != nil {
		return 0, err
	}
	return len(plan), nil
}

// swapDir replaces the folder dst with the folder src.
// The old dst is moved aside first and only deleted once src is in place,
// so a failure part way leaves either the old book or the new one, never neither
func swapDir(src, dst string) (err error) {
	old := filepath.Join(filepath.Dir(dst), "."+filepath.Base(dst)+".old")
	if err = os.RemoveAll(old); err != nil {
		return err
	}
	if err = os.Rename(dst, old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err = os.Rename(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(old)
}
