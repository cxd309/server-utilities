package main

import (
	"math"
	"regexp"
	"strings"
)

// piece is a stretch of one chapter, measured in frames from the chapter's start.
// end is exclusive, so a whole chapter is {chapter, 0, frames}
type piece struct {
	chapter    int // index into the book's chapters, from 0
	start, end int64
}

// framesFor is how many AAC frames a chapter of the given length holds.
// It rounds up, because a part that's one frame too big is worse than one that's one frame short
func framesFor(seconds float64, sampleRate int) (frames int64) {
	frames = int64(math.Ceil(seconds * float64(sampleRate) / samplesPerFrame))
	return frames
}

// cutChapter splits one chapter into the fewest equal pieces that each fit within limit.
// When the frames don't divide evenly, the first pieces get one frame more.
// limit must be above zero; planParts will check that
func cutChapter(chapter int, frames, limit int64) (pieces []piece) {
	if frames <= limit {
		return []piece{{chapter: chapter, start: 0, end: frames}}
	}

	n := (frames + limit - 1) / limit // ceiling division: the fewest pieces that fit
	size, extra := frames/n, frames%n

	var start int64
	for i := range n {
		end := start + size
		if i < extra {
			end++
		}
		pieces = append(pieces, piece{chapter: chapter, start: start, end: end})
		start = end
	}
	return pieces
}

// planParts packs a book's chapters, given as frame counts, into parts of at most limit frames.
// Whole chapters are added to a part until the next one won't fit.
// A chapter too long for any part is first cut into equal pieces
func planParts(chapterFrames []int64, limit int64) (parts [][]piece, err error) {
	if limit <= 0 {
		return nil, errBadLimit
	}

	var current []piece
	var used int64 // frames in current so far
	for i, frames := range chapterFrames {
		for _, p := range cutChapter(i, frames, limit) {
			size := p.end - p.start
			if len(current) > 0 && used+size > limit {
				parts = append(parts, current)
				current, used = nil, 0
			}
			current = append(current, p)
			used += size
		}
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts, nil
}

// planBook splits a book into parts of at most limit frames.
// Disc numbers are ignored: a CD's breaks are just where the disc ran out.
// But a named medium (see meaningfulSection), like "Animal Farm" in a collection,
// always starts a new part, so each section gets parts of its own
func planBook(chapters []chapterFile, limit int64) (parts [][]piece, err error) {
	for start := 0; start < len(chapters); {
		end := start + 1
		for end < len(chapters) && chapters[end].section == chapters[start].section {
			end++
		}

		frames := make([]int64, end-start)
		for i, c := range chapters[start:end] {
			frames[i] = c.frames
		}
		var group [][]piece
		if group, err = planParts(frames, limit); err != nil {
			return nil, err
		}
		for _, part := range group {
			for i := range part {
				part[i].chapter += start // planParts counted from this section's first chapter
			}
			parts = append(parts, part)
		}
		start = end
	}
	return parts, nil
}

// numberedSection matches medium names that only say where a disc comes in the set,
// like "Disc 1", "CD2", "Disk 3 of 4", "Part II", "Vol. 2", "Side A" or just "1".
// A bare number is at most two digits, so a title like "1984" isn't caught, and
// Roman numerals and letters need a word in front, so neither is a title like "Civil"
var numberedSection = regexp.MustCompile(
	`(?i)^(?:(?:disc|disk|cd|part|volume|vol\.?|side|tape)\s*(?:\d+|[ivxlc]+|[a-z])|\d{1,2})(?:\s*(?:of|/)\s*(?:\d+|[ivxlc]+))?$`)

// meaningfulSection returns a medium's name if it says what's on the medium,
// or "" for no name or one that only numbers the disc
func meaningfulSection(name string) (section string) {
	section = strings.TrimSpace(name)
	if numberedSection.MatchString(section) {
		return ""
	}
	return section
}

// outputBytesPerFrame estimates how big a book's parts will be, per frame, for partLimit.
// Copied AAC keeps its size. FLAC becomes ALAC a little bigger than the FLAC.
// AAC that's decoded to ALAC is taken as uncompressed PCM, the most it could be.
// It's 0 (unknown) for a book with no chapters or no frames
func outputBytesPerFrame(d download) (bpf float64) {
	var bytes, frames int64
	for _, c := range d.chapters {
		bytes += c.bytes
		frames += c.frames
	}
	if frames == 0 {
		return 0
	}
	codec := d.chapters[0].codec // a book never mixes codecs
	switch {
	case canCopy(codec, d.sampleRate):
		bpf = float64(bytes) / float64(frames)
	case codec == "flac":
		bpf = alacOverFLAC * float64(bytes) / float64(frames)
	default:
		bpf = pcmBytesPerFrame
	}
	return bpf
}

// partLimit is the most AAC frames one part may hold:
// the smallest of maxPartSeconds, maxPartSamples and maxPartBytes, in frames.
// bytesPerFrame is the expected size of the output, per frame;
// at or below zero means unknown, and the size limit is skipped.
// It rounds down, so a part never runs over
func partLimit(sampleRate int, bytesPerFrame float64) (limit int64) {
	limit = int64(maxPartSeconds) * int64(sampleRate) / samplesPerFrame
	if bySamples := int64(maxPartSamples) / samplesPerFrame; bySamples < limit {
		limit = bySamples
	}
	if bytesPerFrame > 0 {
		if bySize := int64(maxPartBytes / bytesPerFrame); bySize < limit {
			limit = bySize
		}
	}
	return limit
}
