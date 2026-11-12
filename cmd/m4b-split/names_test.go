package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSafeName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Chapter 7. The Hawk’s Flight", "Chapter 7. The Hawk’s Flight"},
		{"Chapter 2: Gary", "Chapter 2 Gary"},
		{"Part One : The Road", "Part One The Road"},
		{`AC/DC "Live" <1992>?`, `AC_DC _Live_ _1992__`},
		{"...And Justice for All", "And Justice for All"},
		{"Ends with dots...", "Ends with dots"},
		{"Chapter  2.", "Chapter 2"},
		{"bell\x07char", "bell_char"},
		{"tab\tinside", "tab inside"},
		{"", "_"},
		{"..", "_"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, safeName(tt.in), "safeName(%q)", tt.in)
	}
}

func TestSafeNameLong(t *testing.T) {
	long := strings.Repeat("é", 150) // 300 bytes: é is two bytes
	got := safeName(long)
	assert.LessOrEqual(t, len(got), maxNameBytes)
	assert.True(t, utf8.ValidString(got), "must not cut a character in half")
}

func TestChapterFileName(t *testing.T) {
	assert.Equal(t, "01 Opening Credits.m4a", chapterFileName(1, 12, "Opening Credits"))
	assert.Equal(t, "12 End Credits.m4a", chapterFileName(12, 12, "End Credits"))
	assert.Equal(t, "007 Bond.m4a", chapterFileName(7, 120, "Bond"))
	assert.Equal(t, "01 Only.m4a", chapterFileName(1, 1, "Only"))
}
