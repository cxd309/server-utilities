package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// buildDir is the hidden folder a book is built in, beside where it ends up.
// Being hidden keeps it out of scans and the Books sync while it's incomplete
func buildDir(dst string) (dir string) {
	dir = filepath.Join(filepath.Dir(dst), "."+filepath.Base(dst)+".building")
	return dir
}

// buildBook plans a book into parts with planBook,
// builds them all in the hidden build folder, then swaps that folder into place at b.dst.
// The album is the book's folder name, e.g. "Mort (Tony Robinson)", not its album tag,
// so two readings of the same book stay two albums on the iPod.
// It returns how many parts it wrote
func buildBook(ctx context.Context, b book, d download, limit int64) (parts int, err error) {
	var plan [][]piece
	if plan, err = planBook(d.chapters, limit); err != nil {
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

	// A cover taken straight from a chapter file made the server's ffmpeg time the chapter track
	// in 1/4,410,000ths of a second, which overflows 32 bits after 16 minutes, and the iPod showed
	// no chapters at all (Eric read by Colin Morgan). Saving the picture on its own first avoids it
	cover := d.cover
	if cover != "" && !strings.EqualFold(filepath.Ext(cover), ".jpg") {
		if cover, err = savePicture(ctx, d.cover); err != nil {
			return 0, fmt.Errorf("cover: %w", err)
		}
		defer func() { _ = os.RemoveAll(filepath.Dir(cover)) }()
	}

	d.tags.title = filepath.Base(b.dst)
	sections := make([]string, len(plan))
	for i, part := range plan {
		sections[i] = d.chapters[part[0].chapter].section // planBook never mixes sections in a part
	}
	titles := partTitles(d.tags.title, sections)

	for i, part := range plan {
		n := i + 1
		out := filepath.Join(tmp, fmt.Sprintf("Part %02d.m4b", n))
		tags := tagLines(d.tags, titles[i], n, len(plan))
		if err = joinPart(ctx, part, d.chapters, d.sampleRate, tags, cover, out); err != nil {
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
