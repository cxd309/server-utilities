// Package musicbrainz is a minimal client for the MusicBrainz web service.
package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const DefaultBaseURL = "https://musicbrainz.org/ws/2"

// Client talks to MusicBrainz, keeping at least MinInterval between requests
// (MusicBrainz allows about one request per second).
type Client struct {
	BaseURL     string
	UserAgent   string
	MinInterval time.Duration
	Retries     int           // extra attempts when MusicBrainz says it's busy (503)
	RetryWait   time.Duration // wait before the first retry; doubles each time
	HTTP        *http.Client

	mu   sync.Mutex
	last time.Time
}

func New(userAgent string) *Client {
	return &Client{
		BaseURL:     DefaultBaseURL,
		UserAgent:   userAgent,
		MinInterval: time.Second,
		Retries:     3,
		RetryWait:   2 * time.Second,
		HTTP:        &http.Client{Timeout: 20 * time.Second},
	}
}

// get fetches path with query q into v, retrying while MusicBrainz is busy.
// It returns found=false for 404.
func (c *Client) get(ctx context.Context, path string, q url.Values, v any) (found bool, err error) {
	wait := c.RetryWait
	for attempt := 0; ; attempt++ {
		found, busy, err := c.getOnce(ctx, path, q, v)
		if !busy || attempt >= c.Retries {
			return found, err
		}
		select {
		case <-time.After(wait):
			wait *= 2
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
}

// getOnce makes a single request. busy reports a 503 from MusicBrainz.
func (c *Client) getOnce(ctx context.Context, path string, q url.Values, v any) (found, busy bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.MinInterval - time.Since(c.last); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return false, false, ctx.Err()
		}
	}
	defer func() { c.last = time.Now() }()

	params := url.Values{}
	for k, vs := range q {
		params[k] = vs
	}
	params.Set("fmt", "json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return false, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false, false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, false, json.NewDecoder(resp.Body).Decode(v)
	case http.StatusNotFound:
		return false, false, nil
	case http.StatusServiceUnavailable:
		return false, true, fmt.Errorf("musicbrainz: GET %s: server busy", path)
	default:
		return false, false, fmt.Errorf("musicbrainz: GET %s: %s", path, resp.Status)
	}
}

// ReleasesByBarcode returns releases with the barcode. It also tries a
// leading 0, as 12-digit UPCs are sometimes stored as 13-digit EANs.
func (c *Client) ReleasesByBarcode(ctx context.Context, barcode string) ([]Release, error) {
	var res struct {
		Releases []Release `json:"releases"`
	}
	q := url.Values{"query": {fmt.Sprintf("barcode:%s OR barcode:0%s", barcode, barcode)}}
	if _, err := c.get(ctx, "/release/", q, &res); err != nil {
		return nil, err
	}
	return res.Releases, nil
}

// ReleaseIDsByDiscID returns the IDs of releases containing the disc.
// An unknown disc ID gives an empty result, not an error.
func (c *Client) ReleaseIDsByDiscID(ctx context.Context, discID string) ([]string, error) {
	var res struct {
		Releases []struct {
			ID string `json:"id"`
		} `json:"releases"`
	}
	found, err := c.get(ctx, "/discid/"+url.PathEscape(discID), nil, &res)
	if err != nil || !found {
		return nil, err
	}
	ids := make([]string, 0, len(res.Releases))
	for _, r := range res.Releases {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
