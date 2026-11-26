package main

import (
	"math"
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
// It uses as few parts as will fit, then evens them out (see evenCuts),
// so a book just over the limit becomes two halves rather than a full part and a stub.
// A chapter too long for any part is first cut into equal pieces
func planParts(chapterFrames []int64, limit int64) (parts [][]piece, err error) {
	if limit <= 0 {
		return nil, errBadLimit
	}

	var pieces []piece
	for i, frames := range chapterFrames {
		pieces = append(pieces, cutChapter(i, frames, limit)...)
	}

	// The fewest parts: fill each one until the next piece won't fit
	var n int
	var used int64 // frames in the part being filled
	for i, p := range pieces {
		size := p.end - p.start
		if i == 0 || used+size > limit {
			n++
			used = 0
		}
		used += size
	}

	for _, run := range evenCuts(pieces, n, limit) {
		parts = append(parts, pieces[run[0]:run[1]])
	}
	return parts, nil
}

// evenCuts splits pieces into n runs of at most limit frames each, as evenly as it can:
// of all the ways to do it, the one with the smallest sum of squared run lengths.
// Each run is returned as [from, to) indexes into pieces.
// planParts makes sure n runs can fit
func evenCuts(pieces []piece, n int, limit int64) (runs [][2]int) {
	m := len(pieces)
	sum := make([]int64, m+1) // sum[j] is the frames in pieces[:j]
	for i, p := range pieces {
		sum[i+1] = sum[i] + p.end - p.start
	}

	// cost[k][j] is the best score for the first j pieces in k runs, and
	// from[k][j] is where the last of those runs starts
	const impossible = math.MaxInt64
	cost := make([][]int64, n+1)
	from := make([][]int, n+1)
	for k := range cost {
		cost[k] = make([]int64, m+1)
		from[k] = make([]int, m+1)
		for j := range cost[k] {
			cost[k][j] = impossible
		}
	}
	cost[0][0] = 0
	for k := 1; k <= n; k++ {
		for j := k; j <= m; j++ {
			for i := j - 1; i >= k-1; i-- { // the last run is pieces[i:j], growing leftwards
				size := sum[j] - sum[i]
				if size > limit {
					break
				}
				if cost[k-1][i] == impossible {
					continue
				}
				if c := cost[k-1][i] + size*size; c < cost[k][j] {
					cost[k][j], from[k][j] = c, i
				}
			}
		}
	}

	runs = make([][2]int, n)
	for k, j := n, m; k > 0; k-- {
		runs[k-1] = [2]int{from[k][j], j}
		j = from[k][j]
	}
	return runs
}

// planBook splits a book into parts of at most limit frames.
// Disc numbers are ignored: a CD's breaks are just where the disc ran out.
// But a section from sections.txt, like "Animal Farm" in a collection,
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

// outputBytesPerFrame estimates how big a book's parts will be, per frame, for partLimit.
// Copied AAC keeps its size. FLAC becomes ALAC a little bigger than the FLAC.
// Anything else decoded to ALAC (MP3, say) is taken as uncompressed PCM, the most it could be.
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
	case canCopy(codec):
		bpf = float64(bytes) / float64(frames)
	case codec == "flac":
		bpf = alacOverFLAC * float64(bytes) / float64(frames)
	default:
		bpf = pcmBytesPerFrame
	}
	return bpf
}

// partLimit is the most AAC frames one part may hold:
// the smaller of maxPartSamples and maxPartBytes, in frames.
// bytesPerFrame is the expected size of the output, per frame;
// at or below zero means unknown, and the size limit is skipped.
// It rounds down, so a part never runs over
func partLimit(bytesPerFrame float64) (limit int64) {
	limit = int64(maxPartSamples) / samplesPerFrame
	if bytesPerFrame > 0 {
		if bySize := int64(maxPartBytes / bytesPerFrame); bySize < limit {
			limit = bySize
		}
	}
	return limit
}
