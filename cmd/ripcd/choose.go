package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/cxd309/server-utilities/internal/musicbrainz"
)

// autoPick returns the release to use without asking: the only release that
// contains this disc, or the only release there is.
func autoPick(rels []musicbrainz.Release, onDisc map[string]bool) (int, bool) {
	match := -1
	for i, r := range rels {
		if onDisc[r.ID] {
			if match >= 0 {
				return 0, false // several editions contain this disc
			}
			match = i
		}
	}
	if match >= 0 {
		return match, true
	}
	if len(rels) == 1 {
		return 0, true
	}
	return 0, false
}

// printReleases writes an aligned table of releases.
func printReleases(w io.Writer, rels []musicbrainz.Release, onDisc map[string]bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tTITLE\tARTIST\tCOUNTRY\tDATE\tLABEL / CAT NO.\tFORMAT\tTRACKS\t")
	for i, r := range rels {
		title := r.Title
		if r.Disambiguation != "" {
			title += " (" + r.Disambiguation + ")"
		}
		mark := ""
		if onDisc[r.ID] {
			mark = "✓ this disc"
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			i+1, title, dash(r.Artist()), dash(r.Country), dash(r.Date),
			dash(strings.Join(r.Labels(), ", ")), dash(strings.Join(r.Formats(), " + ")),
			r.TrackCount, mark)
	}
	tw.Flush()
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// promptPick asks for a number 0..n and returns the chosen index, or -1 for 0 (none).
func promptPick(in io.Reader, out io.Writer, n int) (int, error) {
	sc := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, "Pick 1-%d (0 = none, use the disc ID only): ", n)
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return 0, err
			}
			return 0, errors.New("no choice made")
		}
		k, err := strconv.Atoi(strings.TrimSpace(sc.Text()))
		if err == nil && k >= 0 && k <= n {
			return k - 1, nil
		}
	}
}
