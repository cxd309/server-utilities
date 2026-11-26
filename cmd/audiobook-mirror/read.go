package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// listChapters returns the chapter files of a book, in name order.
// m4b-split names them "001 Title.m4a" and Picard names CD tracks "1-01 Title.flac",
// so name order is chapter order
func listChapters(dir string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(dir); err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || hidden(e.Name()) || !isAudio(e.Name()) {
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

	tags := lowerKeys(info.Format.Tags)
	c = chapterFile{
		path:   path,
		title:  strings.TrimSpace(tags["title"]),
		frames: framesFor(seconds, sampleRate),
		codec:  audio.CodecName,
	}
	if c.title == "" {
		c.title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return c, sampleRate, nil
}

// lowerKeys returns a copy of tags with every key in lower case.
// ffprobe reports m4a tags as "album" but FLAC tags as "ALBUM"
func lowerKeys(tags map[string]string) (lower map[string]string) {
	lower = make(map[string]string, len(tags))
	for k, v := range tags {
		lower[strings.ToLower(k)] = v
	}
	return lower
}

// listItem returns entry i of a Picard multi-value tag such as "Terry Pratchett;Tony Robinson",
// or "" when there are not that many entries
func listItem(list string, i int) (item string) {
	items := strings.Split(list, ";")
	if i < len(items) {
		item = strings.TrimSpace(items[i])
	}
	return item
}

// tagsFromInfo reads the book's tags from ffprobe's report on one of its files.
// Picard writes the same album tags on every file, so any file will do.
// The author is the first of albumartists, as in the Picard folder names, and
// the narrator is the composer, or failing that the second of albumartists
func tagsFromInfo(info ffmpeg.Info) (tags bookTags) {
	t := lowerKeys(info.Format.Tags)
	tags = bookTags{
		title:    strings.TrimSpace(t["album"]),
		author:   listItem(t["albumartists"], 0),
		narrator: strings.TrimSpace(t["composer"]),
		date:     strings.TrimSpace(t["date"]),
	}
	if tags.author == "" {
		tags.author = strings.TrimSpace(t["album_artist"])
	}
	if tags.author == "" {
		tags.author = strings.TrimSpace(t["artist"])
	}
	if tags.narrator == "" {
		tags.narrator = listItem(t["albumartists"], 1)
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

	var firstHasPicture bool
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
		var stat os.FileInfo
		if stat, err = os.Stat(path); err != nil {
			return download{}, err
		}
		c.bytes = stat.Size()
		if i == 0 {
			d.sampleRate = rate
			d.tags = tagsFromInfo(info)
			firstHasPicture = hasPicture(info)
		} else if rate != d.sampleRate {
			return download{}, &sampleRateError{Path: path, Want: d.sampleRate, Got: rate}
		}
		d.chapters = append(d.chapters, c)
	}

	var sections []string
	if sections, err = readSections(dir, paths); err != nil {
		return download{}, err
	}
	for i := range d.chapters {
		d.chapters[i].section = sections[i]
	}

	// Picard embeds the cover in every file but may not save a cover.jpg.
	// Then the first chapter file is the cover's source, and buildBook saves the picture from it
	if d.cover = findCover(dir); d.cover == "" && firstHasPicture {
		d.cover = paths[0]
	}
	return d, nil
}

// hasPicture reports whether ffprobe found an embedded picture, which it lists as a video stream
func hasPicture(info ffmpeg.Info) (yes bool) {
	for _, s := range info.Streams {
		if s.CodecType == "video" {
			return true
		}
	}
	return false
}

// findCover returns the path of dir/cover.jpg, or "" if there isn't one
func findCover(dir string) (path string) {
	path = filepath.Join(dir, "cover.jpg")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}
