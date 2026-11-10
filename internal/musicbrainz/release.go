package musicbrainz

import "strings"

type Release struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Disambiguation string `json:"disambiguation"`
	Country        string `json:"country"`
	Date           string `json:"date"`
	TrackCount     int    `json:"track-count"`
	ArtistCredit   []struct {
		Name string `json:"name"`
	} `json:"artist-credit"`
	LabelInfo []struct {
		CatalogNumber string `json:"catalog-number"`
		Label         *struct {
			Name string `json:"name"`
		} `json:"label"`
	} `json:"label-info"`
	Media []struct {
		Format string `json:"format"`
	} `json:"media"`
}

// Artist is the first credited artist, or "" if none.
func (r Release) Artist() string {
	if len(r.ArtistCredit) == 0 {
		return ""
	}
	return r.ArtistCredit[0].Name
}

// Labels lists unique "Label CATNO" entries.
func (r Release) Labels() []string {
	seen := map[string]bool{}
	var out []string
	for _, li := range r.LabelInfo {
		s := li.CatalogNumber
		if li.Label != nil && li.Label.Name != "" {
			s = strings.TrimSpace(li.Label.Name + " " + li.CatalogNumber)
		}
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// Formats lists each medium's format, e.g. ["CD", "CD"].
func (r Release) Formats() []string {
	out := make([]string, 0, len(r.Media))
	for _, m := range r.Media {
		f := m.Format
		if f == "" {
			f = "?"
		}
		out = append(out, f)
	}
	return out
}
