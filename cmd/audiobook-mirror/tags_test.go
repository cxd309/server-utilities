package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartTitles(t *testing.T) {
	tests := []struct {
		name     string
		sections []string
		want     []string
	}{
		{"one part", []string{""}, []string{"Mort"}},
		{"several parts", []string{"", "", ""}, []string{"Mort, Part 1 of 3", "Mort, Part 2 of 3", "Mort, Part 3 of 3"}},
		{
			"named sections, one needing two parts",
			[]string{"1984", "1984", "Animal Farm"},
			[]string{"1984, Part 1 of 2", "1984, Part 2 of 2", "Animal Farm"},
		},
		{
			"unnamed parts count across the whole book",
			[]string{"", "Appendix"},
			[]string{"Mort, Part 1 of 2", "Appendix"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, partTitles("Mort", tt.sections))
		})
	}
}

func TestNumbered(t *testing.T) {
	assert.Equal(t, "Mort", numbered("Mort", 1, 1))
	assert.Equal(t, "Mort, Part 2 of 4", numbered("Mort", 2, 4))
	assert.Equal(t, "Mort, Part 02 of 12", numbered("Mort", 2, 12))
}

func TestTagLines(t *testing.T) {
	mort := bookTags{
		title:    "Mort",
		author:   "Terry Pratchett read by Tony Robinson",
		narrator: "Tony Robinson",
		date:     "2004",
	}
	tests := []struct {
		name     string
		book     bookTags
		title    string
		n, total int
		want     string
	}{
		{
			"part of several", mort, "Mort, Part 2 of 4", 2, 4,
			"title=Mort, Part 2 of 4\n" +
				"album=Mort\n" +
				"artist=Terry Pratchett read by Tony Robinson\n" +
				"album_artist=Terry Pratchett read by Tony Robinson\n" +
				"composer=Tony Robinson\n" +
				"date=2004\n" +
				"track=2/4\n" +
				"genre=Audiobook\n" +
				"media_type=2\n",
		},
		{
			"named section, no narrator or date",
			bookTags{title: "Orwell", author: "George Orwell"}, "Animal Farm", 3, 12,
			"title=Animal Farm\n" +
				"album=Orwell\n" +
				"artist=George Orwell\n" +
				"album_artist=George Orwell\n" +
				"track=3/12\n" +
				"genre=Audiobook\n" +
				"media_type=2\n",
		},
		{
			"escaped",
			bookTags{title: "A=B; #1", author: "X"}, "A=B; #1", 1, 1,
			`title=A\=B\; \#1` + "\n" +
				`album=A\=B\; \#1` + "\n" +
				"artist=X\n" +
				"album_artist=X\n" +
				"track=1/1\n" +
				"genre=Audiobook\n" +
				"media_type=2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tagLines(tt.book, tt.title, tt.n, tt.total))
		})
	}
}
