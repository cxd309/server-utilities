package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		n, total int
		want     string
	}{
		{
			"part of several", mort, 2, 4,
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
			"only part, no narrator or date",
			bookTags{title: "Thud!", author: "Terry Pratchett"}, 1, 1,
			"title=Thud!\n" +
				"album=Thud!\n" +
				"artist=Terry Pratchett\n" +
				"album_artist=Terry Pratchett\n" +
				"track=1/1\n" +
				"genre=Audiobook\n" +
				"media_type=2\n",
		},
		{
			"ten or more parts are zero padded",
			bookTags{title: "Mort", author: "X"}, 2, 12,
			"title=Mort, Part 02 of 12\n" +
				"album=Mort\n" +
				"artist=X\n" +
				"album_artist=X\n" +
				"track=2/12\n" +
				"genre=Audiobook\n" +
				"media_type=2\n",
		},
		{
			"escaped",
			bookTags{title: "A=B; #1", author: "X"}, 1, 1,
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
			assert.Equal(t, tt.want, tagLines(tt.book, tt.n, tt.total))
		})
	}
}
