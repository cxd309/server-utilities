package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// listChapters returns the chapter files of a downloaded book, in name order.
// m4b-split names them "001 Title.m4a", so name order is chapter order
func listChapters(dir string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(dir); err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || hidden(e.Name()) || strings.ToLower(filepath.Ext(e.Name())) != ".m4a" {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	return paths, nil
}

// chapterFromInfo turns ffprobe's report on one chapter file into a chapterFile,
// plus the file's sample rate.
// A file with no title tag is named after the file
func chapterFromInfo(path string, info ffmpeg.Info) (c chapterFile, sampleRate int, err error) {
	var audio *ffmpeg.Stream
	for i := range info.Streams {
		if info.Streams[i].CodecType == "audio" {
			audio = &info.Streams[i]
			break
		}
	}
	if audio == nil {
		return chapterFile{}, 0, errNoAudio
	}

	if sampleRate, err = strconv.Atoi(audio.SampleRate); err != nil {
		return chapterFile{}, 0, &probeFieldError{Field: "sample rate", Err: err}
	}
	var seconds float64
	if seconds, err = strconv.ParseFloat(info.Format.Duration, 64); err != nil {
		return chapterFile{}, 0, &probeFieldError{Field: "duration", Err: err}
	}

	c = chapterFile{
		path:   path,
		title:  strings.TrimSpace(info.Format.Tags["title"]),
		frames: framesFor(seconds, sampleRate),
	}
	if c.title == "" {
		c.title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return c, sampleRate, nil
}

// tagsFromInfo reads the book's tags from ffprobe's report on one of its files.
// Picard writes the same album tags on every file, so any file will do
func tagsFromInfo(info ffmpeg.Info) (tags bookTags) {
	t := info.Format.Tags
	tags = bookTags{
		title:    strings.TrimSpace(t["album"]),
		author:   strings.TrimSpace(t["album_artist"]),
		narrator: strings.TrimSpace(t["composer"]),
		date:     strings.TrimSpace(t["date"]),
	}
	if tags.author == "" {
		tags.author = strings.TrimSpace(t["artist"])
	}
	return tags
}

// download is everything needed to build the parts of a downloaded book
type download struct {
	chapters   []chapterFile
	sampleRate int
	tags       bookTags
	cover      string // "" when the book has no cover.jpg
}

// readDownload probes every chapter file in dir.
// The book's tags come from the first chapter
func readDownload(ctx context.Context, dir string) (d download, err error) {
	var paths []string
	if paths, err = listChapters(dir); err != nil {
		return download{}, err
	}
	if len(paths) == 0 {
		return download{}, errNoChapters
	}

	for i, path := range paths {
		var info ffmpeg.Info
		var c chapterFile
		var rate int
		if info, err = ffmpeg.Probe(ctx, path); err != nil {
			return download{}, err
		}
		if c, rate, err = chapterFromInfo(path, info); err != nil {
			return download{}, err
		}
		if i == 0 {
			d.sampleRate = rate
			d.tags = tagsFromInfo(info)
		} else if rate != d.sampleRate {
			return download{}, &sampleRateError{Path: path, Want: d.sampleRate, Got: rate}
		}
		d.chapters = append(d.chapters, c)
	}

	d.cover = findCover(dir)
	return d, nil
}

// findCover returns the path of dir/cover.jpg, or "" if there isn't one
func findCover(dir string) (path string) {
	path = filepath.Join(dir, "cover.jpg")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}
