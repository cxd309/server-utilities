package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sectionsIn writes text as sections.txt in a new folder and reads it against files
func sectionsIn(t *testing.T, text string, files ...string) (sections []string, err error) {
	t.Helper()
	dir := t.TempDir()
	if text != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, sectionsFile), []byte(text), 0o644))
	}
	var paths []string
	for _, f := range files {
		paths = append(paths, filepath.Join(dir, f))
	}
	return readSections(dir, paths)
}

func TestReadSections(t *testing.T) {
	files := []string{"1-01 Opening.flac", "1-02 Chapter 1.flac", "1-03 Chapter 2.flac", "2-01 Chapter 1.flac", "2-02 Chapter 2.flac"}
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"no sections.txt", "", []string{"", "", "", "", ""}},
		{
			"chapters before the first section have none",
			"1-02 1984\n2-01 Animal Farm\n",
			[]string{"", "1984", "1984", "Animal Farm", "Animal Farm"},
		},
		{
			"comments, blank lines, spaces and Windows line ends",
			"# Orwell\r\n\r\n1-01   Right Ho, Jeeves  \r\n",
			[]string{"Right Ho, Jeeves", "Right Ho, Jeeves", "Right Ho, Jeeves", "Right Ho, Jeeves", "Right Ho, Jeeves"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := sectionsIn(t, tt.text, files...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, sections)
		})
	}
}

func TestReadSectionsErrors(t *testing.T) {
	files := []string{"001 One.m4a", "002 Two.m4a", "003 Three.m4a"}
	tests := []struct {
		name     string
		text     string
		wantErr  error
		wantLine int
	}{
		{"no name", "001 A\n002\n", errSectionNoName, 2},
		{"no such chapter", "001 A\n004 B\n", errSectionStart, 2},
		{"typo in the number", "01 A\n", errSectionStart, 1},
		{"out of order", "# notes\n002 B\n001 A\n", errSectionOrder, 3},
		{"same start twice", "002 B\n002 C\n", errSectionOrder, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sectionsIn(t, tt.text, files...)
			require.ErrorIs(t, err, tt.wantErr)
			var lineErr *sectionLineError
			require.ErrorAs(t, err, &lineErr)
			assert.Equal(t, tt.wantLine, lineErr.Line)
		})
	}
}
