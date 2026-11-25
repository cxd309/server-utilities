package main

import (
	"errors"
	"fmt"
)

var (
	errBadLimit   = errors.New("frame limit must be above zero")
	errNoAudio    = errors.New("file has no audio stream")
	errNoChapters = errors.New("book folder has no chapter files")
	errSrcNotDir  = errors.New("library is not a folder")
	errDestInSrc  = errors.New("mirror is inside the library, it would scan itself")
	errSrcInDest  = errors.New("library is inside the mirror, prune could delete it")

	errSectionNoName = errors.New("section has no name")
	errSectionStart  = errors.New("section starts at a chapter file that isn't there")
	errSectionOrder  = errors.New("section doesn't start after the one before")
)

// probeFieldError is returned when a value in ffprobe's report can't be parsed.
// Field says which one, e.g. "duration"
type probeFieldError struct {
	Field string
	Err   error
}

func (e *probeFieldError) Error() string {
	return "ffprobe " + e.Field + ": " + e.Err.Error()
}

func (e *probeFieldError) Unwrap() error { return e.Err }

// sampleRateError is returned when a book's chapters don't all share one sample rate.
// They can't be joined without re-encoding, so the book is skipped
type sampleRateError struct {
	Path      string
	Want, Got int
}

func (e *sampleRateError) Error() string {
	return fmt.Sprintf("%s: sample rate %d Hz, but the first chapter is %d Hz", e.Path, e.Got, e.Want)
}

// sectionLineError is returned for a bad line in sections.txt.
// Line counts from 1, and Text is the line as written
type sectionLineError struct {
	Line int
	Text string
	Err  error
}

func (e *sectionLineError) Error() string {
	return fmt.Sprintf("%s line %d %q: %v", sectionsFile, e.Line, e.Text, e.Err)
}

func (e *sectionLineError) Unwrap() error { return e.Err }
