// Package discid computes MusicBrainz disc IDs from a CD's table of contents.
package discid

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

// LeadIn is the standard 2-second (150-frame) gap before track 1.
// MusicBrainz offsets include it.
const LeadIn = 150

// TOC is a disc's table of contents.
// All values are frame offsets (1/75 s) including the LeadIn.
type TOC struct {
	First   int   // first track number
	Last    int   // last track number
	Leadout int   // offset of the lead-out (end of the last track)
	Offsets []int // start offset of each track, First...Last
}

// DiscID returns the MusicBrainz discID.
// See https://musicbrainz.org/doc/Disc_ID_Calculation
func (t TOC) DiscID() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%02X%02X%08X", t.First, t.Last, t.Leadout)
	for i := 0; i < 99; i++ {
		off := 0
		if i < len(t.Offsets) {
			off = t.Offsets[i]
		}
		fmt.Fprintf(&b, "%08X", off)
	}

	sum := sha1.Sum([]byte(b.String()))
	id := base64.StdEncoding.EncodeToString(sum[:])
	return strings.NewReplacer("+", ".", "/", "_", "=", "-").Replace(id)
}
