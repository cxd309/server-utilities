package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// chapter is one chapter of the source file.
// Times are in seconds
type chapter struct {
	start, end float64
	title      string
}

// book is what m4b-split needs to know about one source file
type book struct {
	title    string // album, e.g. "A Wizard of Earthsea"
	author   string
	narrator string
	date     string
	chapters []chapter
}

// bookFromInfo builds a book from ffprobe's report.
// fallbackTitle is used when the file has no album or title tag
func bookFromInfo(info ffmpeg.Info, fallbackTitle string) (b book, err error) {
	if !hasAudio(info.Streams) {
		return book{}, errNoAudio
	}

	tags := lowerKeys(info.Format.Tags)
	tag := func(keys ...string) (value string) {
		for _, k := range keys {
			if value = tidy(tags[k]); value != "" {
				return value
			}
		}
		return ""
	}

	b = book{
		title:    tag("album", "title"),
		author:   tag("album_artist", "artist"),
		narrator: tag("composer"),
		date:     tag("date"),
	}
	if b.title == "" {
		b.title = fallbackTitle
	}

	for _, c := range info.Chapters {
		var ch chapter
		if ch, err = parseChapter(c, len(b.chapters)+1); err != nil {
			return book{}, err
		}
		if ch.end <= ch.start {
			continue // zero-length marker: nothing to put in a file
		}
		b.chapters = append(b.chapters, ch)
	}
	if len(b.chapters) == 0 {
		return book{}, errNoChapters
	}
	return b, nil
}

// parseChapter converts on ffprobe chapter.
// n is position among the kept chapters, used to name chapters that have no title
func parseChapter(c ffmpeg.Chapter, n int) (ch chapter, err error) {
	if ch.start, err = strconv.ParseFloat(c.StartTime, 64); err != nil {
		return chapter{}, fmt.Errorf("chapter %d start: %w", n, err)
	}
	if ch.end, err = strconv.ParseFloat(c.EndTime, 64); err != nil {
		return chapter{}, fmt.Errorf("chapter %d end: %w", n, err)
	}
	if ch.title = tidy(lowerKeys(c.Tags)["title"]); ch.title == "" {
		ch.title = fmt.Sprintf("Chapter %d", n)
	}
	return ch, nil
}

// hasAudio report whether any stream is audio
func hasAudio(streams []ffmpeg.Stream) (found bool) {
	for _, s := range streams {
		if s.CodecType == "audio" {
			return true
		}
	}
	return false
}

// probeBook asks ffprobe about the file at path and turns the answer into a book.
// The file name stands in for the title if there are no tags.
func probeBook(ctx context.Context, path string) (b book, err error) {
	var info ffmpeg.Info
	if info, err = ffmpeg.Probe(ctx, path); err != nil {
		return book{}, err
	}
	fallback := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return bookFromInfo(info, fallback)
}
