package discid

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
)

// ErrNoTracks means the output contained no track rows (no disc, or not an audio CD)
var ErrNoTracks = errors.New("discid: no tracks in cd-paranoia output")

// Matches cd-paranoia -Q track rows, e.g.
//
//  1. 13962 [03:06.12]        0 [00:00.00]    no   no  2
//
// capturing track number, length and begin (both in frames, without LeadIn).
var paranoiaRow = regexp.MustCompile(`(?m)^\s*(\d+)\.\s+(\d+)\s+\[[^\]]*\]\s+(\d+)\s+\[`)

// ParseParanoiaTOC builds a TOC from `cd-paranoia -Q` output (audio tracks only).
func ParseParanoiaTOC(out string) (TOC, error) {
	rows := paranoiaRow.FindAllStringSubmatch(out, -1)
	if len(rows) == 0 {
		return TOC{}, ErrNoTracks
	}

	var toc TOC
	var lastBegin, lastLength int
	for i, row := range rows {
		track, _ := strconv.Atoi(row[1]) // regex guarantees digits
		length, _ := strconv.Atoi(row[2])
		begin, _ := strconv.Atoi(row[3])

		if i == 0 {
			toc.First = track
		}
		toc.Last = track
		toc.Offsets = append(toc.Offsets, begin+LeadIn)
		lastBegin, lastLength = begin, length
	}
	toc.Leadout = lastBegin + lastLength + LeadIn
	return toc, nil
}

// ReadTOC reads the table of contents of the disc in drive using cd-paranoia.
func ReadTOC(drive string) (TOC, error) {
	// cd-paranoia prints the TOC on stderr and its exit code isn't reliable,
	// so judge success by whether the output parses.
	out, _ := exec.Command("cd-paranoia", "-d", drive, "-Q").CombinedOutput()
	return ParseParanoiaTOC(string(out))
}
