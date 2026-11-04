package musicbrainz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMB serves fixture files by path and records the last request.
func fakeMB(t *testing.T, routes map[string]string) (*Client, *http.Request) {
	t.Helper()
	var last http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last = *r
		file, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, err := os.ReadFile(file)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	c := New("ripcd-test/0")
	c.BaseURL = srv.URL
	c.MinInterval = 0
	return c, &last
}

func TestReleasesByBarcode(t *testing.T) {
	c, req := fakeMB(t, map[string]string{"/release/": "testdata/barcode_634904074029.json"})

	rels, err := c.ReleasesByBarcode(context.Background(), "634904074029")
	require.NoError(t, err)

	assert.Equal(t, "barcode:634904074029 OR barcode:0634904074029", req.URL.Query().Get("query"))
	assert.Equal(t, "json", req.URL.Query().Get("fmt"))
	assert.Equal(t, "ripcd-test/0", req.Header.Get("User-Agent"))

	ids := make([]string, 0, len(rels))
	for _, r := range rels {
		ids = append(ids, r.ID)
	}
	assert.Subset(t, ids, []string{
		"2b4c1e9c-1904-4155-9df6-d1681ddef130", // AU
		"0a536975-66b0-447d-8877-42cbc079e10c", // no country
		"a4f190d5-c752-4ebd-8d2f-925abfb52c40", // GB
	})

	for _, r := range rels {
		if r.ID == "a4f190d5-c752-4ebd-8d2f-925abfb52c40" {
			assert.Equal(t, "25", r.Title)
			assert.Equal(t, "Adele", r.Artist())
			assert.Equal(t, "GB", r.Country)
			assert.Equal(t, 11, r.TrackCount)
			assert.Equal(t, []string{"CD"}, r.Formats())
		}
	}
}

func TestReleaseIDsByDiscID(t *testing.T) {
	c, req := fakeMB(t, map[string]string{
		"/discid/KvJxY_MY82qy2MVu2GhP2r5NAm4-": "testdata/discid_loussier.json",
	})

	ids, err := c.ReleaseIDsByDiscID(context.Background(), "KvJxY_MY82qy2MVu2GhP2r5NAm4-")
	require.NoError(t, err)
	assert.NotEmpty(t, ids)
	assert.Equal(t, "/discid/KvJxY_MY82qy2MVu2GhP2r5NAm4-", req.URL.Path)
}

func TestReleaseIDsByDiscIDUnknown(t *testing.T) {
	c, _ := fakeMB(t, nil) // everything 404s

	ids, err := c.ReleaseIDsByDiscID(context.Background(), "nope")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRateLimit(t *testing.T) {
	c, _ := fakeMB(t, nil)
	c.MinInterval = 50 * time.Millisecond

	start := time.Now()
	_, _ = c.ReleaseIDsByDiscID(context.Background(), "a")
	_, _ = c.ReleaseIDsByDiscID(context.Background(), "b")
	assert.GreaterOrEqual(t, time.Since(start), 50*time.Millisecond)
}

func TestRetriesWhenBusy(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		body, err := os.ReadFile("testdata/discid_loussier.json")
		require.NoError(t, err)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	c := New("ripcd-test/0")
	c.BaseURL, c.MinInterval, c.RetryWait = srv.URL, 0, time.Millisecond

	ids, err := c.ReleaseIDsByDiscID(context.Background(), "KvJxY_MY82qy2MVu2GhP2r5NAm4-")
	require.NoError(t, err)
	assert.NotEmpty(t, ids)
	assert.Equal(t, 3, calls)
}

func TestGivesUpWhenStillBusy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c := New("ripcd-test/0")
	c.BaseURL, c.MinInterval, c.RetryWait, c.Retries = srv.URL, 0, time.Millisecond, 2

	_, err := c.ReleaseIDsByDiscID(context.Background(), "x")
	assert.ErrorContains(t, err, "server busy")
}
