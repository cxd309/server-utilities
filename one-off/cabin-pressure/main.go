// Command cabin-pressure is a one-off: it turns the ripped box set
// "The Complete Cabin Pressure: From A to Z"
// (14 CDs, each episode split into "Abu Dhabi, TR 1" … "TR 9" tracks)
// into one ALAC file per episode, tagged as a podcast so iTunes files it under Podcasts.
//
// Titles, release dates and descriptions come from episodes.csv, in A–Z order.
// Run it once on the server, from the repo root:
//
//	go run ./one-off/cabin-pressure --dry-run
//	go run ./one-off/cabin-pressure
//
// It skips episodes that already exist, so it's safe to run again after a failure.
package main

import (
	"context"
	_ "embed"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/cxd309/server-utilities/internal/podcast"
)

const (
	src    = "/mnt/prudence/audiobooks/John Finnemore/The Complete Cabin Pressure From A to Z (Audio Drama)"
	dest   = "/mnt/prudence/podcasts/Cabin Pressure"
	show   = "Cabin Pressure"
	artist = "John Finnemore"
)

//go:embed episodes.csv
var episodesCSV string

var (
	errNoTrackMark   = errors.New(`file name is not "NN-NN Episode, TR n"`)
	errEpisodeCount  = errors.New("the CDs and episodes.csv have a different number of episodes")
	errMissingColumn = errors.New("episodes.csv is missing a column")
)

// mismatchError is returned when a CD's episode name isn't in the matching CSV title,
// which would mean the two lists have drifted out of step
type mismatchError struct {
	N           int
	Disc, Title string
}

func (e *mismatchError) Error() string {
	return fmt.Sprintf("episode %d: CD says %q but episodes.csv says %q", e.N, e.Disc, e.Title)
}

// episode is one row of episodes.csv
type episode struct {
	title, releaseDate, description string
}

// group is one episode's tracks on the CDs, in order
type group struct {
	name   string
	tracks []string
}

// trackName matches Picard's names for these tracks, e.g. "13-10 Zurich, Part 1, TR 1".
// (.+) is greedy, so it runs up to the last ", TR n" and keeps names like "Zurich, Part 1" whole
var trackName = regexp.MustCompile(`^\d+-\d+ (.+), TR \d+$`)

// episodeName gets the episode from a track's file name:
// "13-10 Zurich, Part 1, TR 1.flac" -> "Zurich, Part 1"
func episodeName(path string) (name string, err error) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := trackName.FindStringSubmatch(base)
	if m == nil {
		return "", fmt.Errorf("%s: %w", base, errNoTrackMark)
	}
	name = m[1]
	return name, nil
}

// groupTracks gathers consecutive tracks that share an episode name
func groupTracks(paths []string) (groups []group, err error) {
	for _, p := range paths {
		var name string
		if name, err = episodeName(p); err != nil {
			return nil, err
		}
		if len(groups) == 0 || groups[len(groups)-1].name != name {
			groups = append(groups, group{name: name})
		}
		last := &groups[len(groups)-1]
		last.tracks = append(last.tracks, p)
	}
	return groups, nil
}

// readEpisodes parses episodes.csv by column name
func readEpisodes(r io.Reader) (episodes []episode, err error) {
	var rows [][]string
	if rows, err = csv.NewReader(r).ReadAll(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errMissingColumn
	}
	col := make(map[string]int)
	for i, h := range rows[0] {
		col[h] = i
	}
	for _, name := range []string{"title", "release_date", "description"} {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("%s: %w", name, errMissingColumn)
		}
	}
	for _, row := range rows[1:] {
		episodes = append(episodes, episode{
			title:       row[col["title"]],
			releaseDate: row[col["release_date"]],
			description: row[col["description"]],
		})
	}
	return episodes, nil
}

// pair checks the CD groups and CSV rows line up, one to one and in order
func pair(groups []group, episodes []episode) (err error) {
	if len(groups) != len(episodes) {
		return fmt.Errorf("%d on the CDs, %d in the CSV: %w", len(groups), len(episodes), errEpisodeCount)
	}
	for i, g := range groups {
		if !strings.Contains(strings.ToLower(episodes[i].title), strings.ToLower(g.name)) {
			return &mismatchError{N: i + 1, Disc: g.name, Title: episodes[i].title}
		}
	}
	return nil
}

// isoUTC turns a CSV date such as "2008-07-02T11:30:00+01:00" into "2008-07-02T10:30:00Z",
// the form iTunes expects
func isoUTC(date string) (utc string, err error) {
	var t time.Time
	if t, err = time.Parse(time.RFC3339, date); err != nil {
		return "", err
	}
	utc = t.UTC().Format("2006-01-02T15:04:05Z")
	return utc, nil
}

// concatList is the ffmpeg concat list for one episode's tracks
func concatList(tracks []string) (list string) {
	var b strings.Builder
	for _, t := range tracks {
		fmt.Fprintf(&b, "file '%s'\n", strings.ReplaceAll(t, "'", `'\''`))
	}
	list = b.String()
	return list
}

// joinArgs joins one episode's FLAC tracks into a lossless ALAC file tagged as a podcast
func joinArgs(listPath, cover, out string, ep episode, date string, n, total int) (args []string) {
	args = []string{
		"-f", "concat", "-safe", "0", "-i", listPath,
		"-i", cover,
		"-map", "0:a:0", "-map", "1:v:0",
		"-map_metadata", "-1",
		"-c:a", "alac", "-sample_fmt", "s16p",
		"-c:v", "copy", "-disposition:v", "attached_pic",
		"-metadata", "title=" + ep.title,
		"-metadata", "album=" + show,
		"-metadata", "artist=" + artist,
		"-metadata", "album_artist=" + artist,
		"-metadata", "date=" + date,
		"-metadata", fmt.Sprintf("track=%d/%d", n, total),
		"-metadata", "genre=Podcast",
		"-metadata", "description=" + ep.description,
		"-metadata", "media_type=21", // iTunes "Podcast"
		"-f", "ipod", out,
	}
	return args
}

// fileName is the output name for episode n, e.g. "07 Gdansk.m4a"
func fileName(n int, title string) (name string) {
	name = fmt.Sprintf("%02d %s.m4a", n, strings.NewReplacer("/", "-", ":", "").Replace(title))
	return name
}

func main() {
	dryRun := flag.Bool("dry-run", false, "show the plan, write nothing")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dryRun bool) (err error) {
	var tracks []string
	if tracks, err = filepath.Glob(filepath.Join(src, "*.flac")); err != nil {
		return err
	}
	var groups []group
	if groups, err = groupTracks(tracks); err != nil { // Glob sorts, so disc order
		return err
	}
	var episodes []episode
	if episodes, err = readEpisodes(strings.NewReader(episodesCSV)); err != nil {
		return err
	}
	if err = pair(groups, episodes); err != nil {
		return err
	}
	if !dryRun {
		if err = os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
	}

	for i, g := range groups {
		n, ep := i+1, episodes[i]
		out := filepath.Join(dest, fileName(n, ep.title))
		var date string
		if date, err = isoUTC(ep.releaseDate); err != nil {
			return fmt.Errorf("episode %d release_date: %w", n, err)
		}
		if _, statErr := os.Stat(out); statErr == nil {
			fmt.Printf("skip  %s (already there)\n", filepath.Base(out))
			continue
		}
		fmt.Printf("%s  <- %d tracks of %q, dated %s\n", filepath.Base(out), len(g.tracks), g.name, date)
		if dryRun {
			continue
		}
		if err = joinEpisode(ctx, g, ep, date, n, len(groups), out); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(out), err)
		}
	}
	return nil
}

// joinEpisode writes one episode via a .part file, renamed into place when complete
func joinEpisode(ctx context.Context, g group, ep episode, date string, n, total int, out string) (err error) {
	var dir string
	if dir, err = os.MkdirTemp("", "cabin-pressure-"); err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	listPath := filepath.Join(dir, "list.txt")
	if err = os.WriteFile(listPath, []byte(concatList(g.tracks)), 0o600); err != nil {
		return err
	}
	part := out + ".part"
	defer func() {
		if err != nil {
			_ = os.Remove(part)
		}
	}()
	if err = ffmpeg.Run(ctx, joinArgs(listPath, filepath.Join(src, "cover.jpg"), part, ep, date, n, total)...); err != nil {
		return err
	}
	// Without this flag iTunes silently refuses the file, even though its media kind is Podcast
	if err = podcast.Flag(ctx, part); err != nil {
		return err
	}
	return os.Rename(part, out)
}
