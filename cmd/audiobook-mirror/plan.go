package main

import "math"

const (
	// samplesPerFrame is the size of one AAC-LC frame
	samplesPerFrame = 1024

	// maxPartSeconds is the longest a part may play for: 79 minutes, so it fits on an 80-minute CD-R.
	// It also keeps every part inside what's been tested to play on the iPod mini:
	// files up to 2 h and ~310k AAC frames played, and 3½ h, ~543k frames crashed.
	// Size isn't the cause: a 245 MB ALAC file played. Whether it's time or frames is unknown,
	// and 79 min stays under both
	maxPartSeconds = 79 * 60
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

// partLimit is how many AAC frames maxPartSeconds holds at sampleRate.
// It rounds down, so a part never runs over
func partLimit(sampleRate int) (limit int64) {
	limit = int64(maxPartSeconds) * int64(sampleRate) / samplesPerFrame
	return limit
}
