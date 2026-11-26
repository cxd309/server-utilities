package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// episode is one row of a manifest: one source file that becomes one podcast episode
type episode struct {
	source      string // relative to libraryRoot
	show        string // the podcast, and its folder under podcastRoot
	artist      string
	title       string
	releaseDate string // UTC, as iTunes wants it: "1978-03-08T12:00:00Z"
	description string

	// start and end cut the episode from part of the source, in seconds.
	// An end of 0 means the end of the source
	start, end float64
}

// manifestColumns are the columns a manifest must have, found by name in its header
var manifestColumns = []string{"source", "show", "artist", "title", "release_date", "description"}

// readManifest parses a manifest CSV. Rows are episodes in order, and a show's episodes are
// numbered in the order they appear. release_date is RFC 3339, e.g. "1978-03-08T22:30:00Z".
// Every field but description must be filled in.
// Optional start and end columns, in seconds, take an episode from part of its source,
// for a file that holds several episodes with no chapter marks between them
func readManifest(r io.Reader) (episodes []episode, err error) {
	var rows [][]string
	if rows, err = csv.NewReader(r).ReadAll(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errMissingColumn
	}
	col := make(map[string]int)
	for i, h := range rows[0] {
		col[strings.TrimSpace(h)] = i
	}
	for _, name := range manifestColumns {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("%s: %w", name, errMissingColumn)
		}
	}

	for i, row := range rows[1:] {
		field := func(name string) (value string) {
			return strings.TrimSpace(row[col[name]])
		}
		ep := episode{
			source:      field("source"),
			show:        field("show"),
			artist:      field("artist"),
			title:       field("title"),
			description: field("description"),
		}
		for _, name := range []string{"source", "show", "artist", "title", "release_date"} {
			if field(name) == "" {
				return nil, &rowError{Row: i + 2, Err: fmt.Errorf("%s: %w", name, errEmptyField)}
			}
		}
		var t time.Time
		if t, err = time.Parse(time.RFC3339, field("release_date")); err != nil {
			return nil, &rowError{Row: i + 2, Err: err}
		}
		ep.releaseDate = t.UTC().Format("2006-01-02T15:04:05Z")
		if ep.start, ep.end, err = cutFromRow(row, col); err != nil {
			return nil, &rowError{Row: i + 2, Err: err}
		}
		episodes = append(episodes, ep)
	}
	return episodes, nil
}

// cutFromRow reads a row's optional start and end columns, in seconds.
// Empty or missing, start is the beginning and end is the end of the source
func cutFromRow(row []string, col map[string]int) (start, end float64, err error) {
	for _, c := range []struct {
		name  string
		value *float64
	}{{"start", &start}, {"end", &end}} {
		i, ok := col[c.name]
		if !ok || strings.TrimSpace(row[i]) == "" {
			continue
		}
		if *c.value, err = strconv.ParseFloat(strings.TrimSpace(row[i]), 64); err != nil {
			return 0, 0, fmt.Errorf("%s: %w", c.name, err)
		}
	}
	if start < 0 || (end != 0 && end <= start) {
		return 0, 0, fmt.Errorf("start %g, end %g: %w", start, end, errBadCut)
	}
	return start, end, nil
}
