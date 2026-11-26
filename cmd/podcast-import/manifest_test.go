package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const header = "source,show,artist,title,release_date,description\n"

func TestReadManifest(t *testing.T) {
	csv := header +
		`Douglas Adams/Primary/02 Fit the First.m4a,The Show,Douglas Adams,Fit the First,1978-03-08T22:30:00Z,"First broadcast 8 March 1978, Radio 4."` + "\n" +
		"x/2-01 Mushrooms.flac,Other,J.R.R. Tolkien,Mushrooms,2004-09-21T18:30:00+01:00,\n"
	episodes, err := readManifest(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, episodes, 2)
	assert.Equal(t, episode{
		source:      "Douglas Adams/Primary/02 Fit the First.m4a",
		show:        "The Show",
		artist:      "Douglas Adams",
		title:       "Fit the First",
		releaseDate: "1978-03-08T22:30:00Z",
		description: "First broadcast 8 March 1978, Radio 4.",
	}, episodes[0])
	assert.Equal(t, "2004-09-21T17:30:00Z", episodes[1].releaseDate, "turned into UTC")
	assert.Equal(t, "", episodes[1].description, "description may be empty")
}

func TestReadManifestErrors(t *testing.T) {
	tests := []struct {
		name    string
		csv     string
		wantErr error
		wantRow int
	}{
		{"empty", "", errMissingColumn, 0},
		{"no description column", "source,show,artist,title,release_date\n", errMissingColumn, 0},
		{"empty title", header + "a.m4a,S,A,,1978-03-08T22:30:00Z,d\n", errEmptyField, 2},
		{"empty date", header + "a.m4a,S,A,T,1978-03-08T22:30:00Z,d\nb.m4a,S,A,T,,d\n", errEmptyField, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readManifest(strings.NewReader(tt.csv))
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantRow > 0 {
				var re *rowError
				require.ErrorAs(t, err, &re)
				assert.Equal(t, tt.wantRow, re.Row)
			}
		})
	}
}

func TestReadManifestBadDate(t *testing.T) {
	_, err := readManifest(strings.NewReader(header + "a.m4a,S,A,T,8 March 1978,d\n"))
	var re *rowError
	require.ErrorAs(t, err, &re)
	assert.Equal(t, 2, re.Row)
}

func TestReadManifestCuts(t *testing.T) {
	csv := "source,show,artist,title,release_date,description,start,end\n" +
		"a.m4a,S,A,One,2018-03-08T18:30:00Z,,,2355.04\n" +
		"a.m4a,S,A,Two,2018-03-15T18:30:00Z,,2355.04,4543.94\n" +
		"a.m4a,S,A,Six,2018-04-12T18:30:00+01:00,,11060.48,\n"
	episodes, err := readManifest(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, episodes, 3)
	assert.Equal(t, [2]float64{0, 2355.04}, [2]float64{episodes[0].start, episodes[0].end})
	assert.Equal(t, [2]float64{2355.04, 4543.94}, [2]float64{episodes[1].start, episodes[1].end})
	assert.Equal(t, [2]float64{11060.48, 0}, [2]float64{episodes[2].start, episodes[2].end}, "to the end")

	for _, bad := range []string{",abc", "-1,", "10,5", "10,10"} {
		_, err = readManifest(strings.NewReader("source,show,artist,title,release_date,description,start,end\n" +
			"a.m4a,S,A,T,2018-03-08T18:30:00Z,," + bad + "\n"))
		var re *rowError
		assert.ErrorAs(t, err, &re, bad)
	}
}
