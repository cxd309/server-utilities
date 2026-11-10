package discid

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParanoiaTOC(t *testing.T) {
	out, err := os.ReadFile("testdata/paranoia_q.txt")
	require.NoError(t, err)

	toc, err := ParseParanoiaTOC(string(out))
	require.NoError(t, err)

	assert.Equal(t, 1, toc.First)
	assert.Equal(t, 14, toc.Last)
	assert.Equal(t, 203342, toc.Leadout) // 189110 + 14082 + 150
	assert.Len(t, toc.Offsets, 14)
	assert.Equal(t, 150, toc.Offsets[0])
	assert.Equal(t, 189260, toc.Offsets[13])
	assert.Equal(t, "1QYnhv0RV6aSgltrgNxa22bDGgA-", toc.DiscID())
}

func TestParseParanoiaTOCNoTracks(t *testing.T) {
	_, err := ParseParanoiaTOC("cdparanoia III release 10.2\n\nUnable to open disc.\n")
	assert.ErrorIs(t, err, ErrNoTracks)
}
