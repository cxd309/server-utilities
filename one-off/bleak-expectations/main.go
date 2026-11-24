// Command bleak-expectations is a one-off: it turns the Audible download of
// "Bleak Expectations: The Complete BBC Radio 4 Series"
// (30 chapters, one per episode, named "Chapter 1" … "Chapter 30")
// into 30 podcast episodes with their broadcast titles,
// so iTunes files them under Podcasts.
//
// The audio is copied, not re-encoded. Titles, release dates and descriptions come
// from episodes.csv. Run it once on the server, from the repo root:
//
//	go run ./one-off/bleak-expectations --dry-run
//	go run ./one-off/bleak-expectations
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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/cxd309/server-utilities/internal/podcast"
)

const (
	src    = "/mnt/prudence/audiobooks/Mark Evans/Bleak Expectations The Complete BBC Radio 4 Series (Audio Drama)"
	dest   = "/mnt/prudence/podcasts/Bleak Expectations"
	show   = "Bleak Expectations"
	artist = "Mark Evans"

	// cover is optional artwork, relative to the repo root. The download has none embedded
	cover = "one-off/bleak-expectations/cover.jpg"
)

//go:embed episodes.csv
var episodesCSV string

var (
	errNotChapter    = errors.New(`file name is not "NN Chapter N"`)
	errEpisodeCount  = errors.New("the download and episodes.csv have a different number of episodes")
	errMissingColumn = errors.New("episodes.csv is missing a column")
)

// orderError is returned when the files aren't Chapter 1, 2, 3 … in order,
// which would pair them with the wrong CSV rows
type orderError struct {
	Path      string
	Want, Got int
}

func (e *orderError) Error() string {
	return fmt.Sprintf("%s: expected chapter %d, found chapter %d", e.Path, e.Want, e.Got)
}

// episode is one row of episodes.csv
type episode struct {
	title, releaseDate, description string
}

// chapterFile matches Picard's names for the download's files, e.g. "07 Chapter 7"
var chapterFile = regexp.MustCompile(`^\d+ Chapter (\d+)$`)

// chapterNumber gets N from a file named "NN Chapter N.m4a"
func chapterNumber(path string) (n int, err error) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := chapterFile.FindStringSubmatch(base)
	if m == nil {
		return 0, fmt.Errorf("%s: %w", base, errNotChapter)
	}
	n, err = strconv.Atoi(m[1])
	return n, err
}

// checkOrder makes sure the files are chapters 1 … len(paths), in order
func checkOrder(paths []string) (err error) {
	for i, p := range paths {
		var n int
		if n, err = chapterNumber(p); err != nil {
			return err
		}
		if n != i+1 {
			return &orderError{Path: filepath.Base(p), Want: i + 1, Got: n}
		}
	}
	return nil
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

// isoUTC turns a CSV date such as "2007-08-15T11:30:00+01:00" into "2007-08-15T10:30:00Z",
// the form iTunes expects
func isoUTC(date string) (utc string, err error) {
	var t time.Time
	if t, err = time.Parse(time.RFC3339, date); err != nil {
		return "", err
	}
	utc = t.UTC().Format("2006-01-02T15:04:05Z")
	return utc, nil
}

// copyArgs copies one chapter's audio unchanged into a file tagged as a podcast.
// coverPath is "" for no artwork
func copyArgs(in, coverPath, out string, ep episode, date string, n, total int) (args []string) {
	args = []string{"-i", in}
	if coverPath != "" {
		args = append(args, "-i", coverPath, "-map", "0:a:0", "-map", "1:v:0")
	} else {
		args = append(args, "-map", "0:a:0")
	}
	args = append(args,
		"-map_metadata", "-1",
		"-c", "copy", // the AAC is copied as it is
		"-disposition:v", "attached_pic",
		"-metadata", "title="+ep.title,
		"-metadata", "album="+show,
		"-metadata", "artist="+artist,
		"-metadata", "album_artist="+artist,
		"-metadata", "date="+date,
		"-metadata", fmt.Sprintf("track=%d/%d", n, total),
		"-metadata", "genre=Podcast",
		"-metadata", "description="+ep.description,
		"-metadata", "media_type=21", // iTunes "Podcast"
		"-f", "ipod", out,
	)
	return args
}

// fileName is the output name for episode n, e.g. "01 A Childhood Cruelly Kippered.m4a"
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
	var files []string
	if files, err = filepath.Glob(filepath.Join(src, "*.m4a")); err != nil {
		return err
	}
	if err = checkOrder(files); err != nil {
		return err
	}
	var episodes []episode
	if episodes, err = readEpisodes(strings.NewReader(episodesCSV)); err != nil {
		return err
	}
	if len(files) != len(episodes) {
		return fmt.Errorf("%d files, %d in the CSV: %w", len(files), len(episodes), errEpisodeCount)
	}

	coverPath := cover
	if _, statErr := os.Stat(coverPath); statErr != nil {
		fmt.Printf("note: no %s, so the episodes will have no artwork\n", cover)
		coverPath = ""
	}
	if !dryRun {
		if err = os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
	}

	for i, f := range files {
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
		fmt.Printf("%s  ← %s, dated %s\n", filepath.Base(out), filepath.Base(f), date)
		if dryRun {
			continue
		}
		if err = copyEpisode(ctx, f, coverPath, out, ep, date, n, len(files)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(out), err)
		}
	}
	return nil
}

// copyEpisode writes one episode via a .part file, renamed into place when complete
func copyEpisode(ctx context.Context, in, coverPath, out string, ep episode, date string, n, total int) (err error) {
	part := out + ".part"
	defer func() {
		if err != nil {
			_ = os.Remove(part)
		}
	}()
	if err = ffmpeg.Run(ctx, copyArgs(in, coverPath, part, ep, date, n, total)...); err != nil {
		return err
	}
	// Without this flag iTunes silently refuses the file, even though its media kind is Podcast
	if err = podcast.Flag(ctx, part); err != nil {
		return err
	}
	return os.Rename(part, out)
}
