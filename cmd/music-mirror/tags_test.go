package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinValues(t *testing.T) {
	assert.Equal(t, "Rock; Pop", joinValues([]string{"Rock", " Pop "}))
	assert.Equal(t, "Rock", joinValues([]string{"", "Rock", "  "}))
	assert.Equal(t, "", joinValues(nil))
}

func TestYearOf(t *testing.T) {
	assert.Equal(t, "1968", yearOf("1968-10-16", "2012-09-10"), "first usable date wins")
	assert.Equal(t, "2012", yearOf("", "2012-09-10"), "falls back when original is missing")
	assert.Equal(t, "1999", yearOf("1999"))
	assert.Equal(t, "", yearOf("n/a", ""))
}

func TestSplitCredit(t *testing.T) {
	tests := []struct {
		name, credit string
		artists      []string
		primary      string
		rest         string
	}{
		{"featured", "Calvin Harris feat. Rihanna", []string{"Calvin Harris", "Rihanna"}, "Calvin Harris", "feat. Rihanna"},
		{"duet", "Queen & David Bowie", []string{"Queen", "David Bowie"}, "Queen", "& David Bowie"},
		{"three artists", "A, B & C", []string{"A", "B", "C"}, "A", ", B & C"},
		{"one artist with ampersand", "Simon & Garfunkel", []string{"Simon & Garfunkel"}, "Simon & Garfunkel", ""},
		{"no ARTISTS tag", "Bob Dylan", nil, "Bob Dylan", ""},
		{"credit not starting with first artist", "Someone Else", []string{"A", "B"}, "A", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, rest := splitCredit(tt.credit, tt.artists)
			assert.Equal(t, tt.primary, primary)
			assert.Equal(t, tt.rest, rest)
		})
	}
}

func TestTitleWithGuests(t *testing.T) {
	tests := []struct {
		name, title, rest, want string
	}{
		{"no guests", "Song", "", "Song"},
		{"feat.", "Song", "feat. Rihanna", "Song (feat. Rihanna)"},
		{"ft. normalised", "Song", "ft. Rihanna", "Song (feat. Rihanna)"},
		{"ampersand", "Under Pressure", "& David Bowie", "Under Pressure (with David Bowie)"},
		{"comma list", "Song", ", B & C", "Song (with B & C)"},
		{"x collab", "Song", "x Stormzy", "Song (with Stormzy)"},
		{"title already has feat.", "Song (feat. Rihanna)", "feat. Rihanna", "Song (feat. Rihanna)"},
		{"join with nothing after", "Song", "&", "Song"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, titleWithGuests(tt.title, tt.rest))
		})
	}
}
