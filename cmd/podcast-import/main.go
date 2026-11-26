// Command podcast-import turns audiobook-library files into podcast episodes, for radio series
// that read better as a podcast than as an audiobook (one show, one episode per broadcast).
// A manifest CSV lists the episodes, each from a source file or a stretch of one
// (see readManifest for its columns).
// Each episode goes to /mnt/prudence/podcasts/<show>/NN Title.m4a: AAC is copied as it is,
// FLAC becomes ALAC, the cover comes from the file or a cover.jpg beside it, and the file is
// tagged and flagged as a podcast so iTunes takes it.
//
//	podcast-import --dry-run podcasts/hitchhikers/episodes.csv
//	podcast-import podcasts/hitchhikers/episodes.csv
//
// Episodes that already exist are skipped, so it's safe to run again after a failure.
// To redo one, delete its file first, but remember iTunes's deleted-episode memory.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "show the plan, write nothing")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, flag.Arg(0), *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run reads the manifest, checks every source can be read, then makes each episode
func run(ctx context.Context, manifest string, dryRun bool) (err error) {
	if manifest == "" {
		return errNoManifest
	}
	var f *os.File
	if f, err = os.Open(manifest); err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	var episodes []episode
	if episodes, err = readManifest(f); err != nil {
		return err
	}

	// Probe everything first, so a typo in the manifest stops the run before anything is written
	totals := make(map[string]int) // episodes per show
	sources := make([]source, len(episodes))
	for i, ep := range episodes {
		totals[ep.show]++
		var path string
		if path, err = findFile(libraryRoot, ep.source); err != nil {
			return &rowError{Row: i + 2, Err: err}
		}
		if sources[i], err = probeSource(ctx, path); err != nil {
			return &rowError{Row: i + 2, Err: err}
		}
	}

	counts := make(map[string]int) // episodes per show so far
	for i, ep := range episodes {
		counts[ep.show]++
		n, total := counts[ep.show], totals[ep.show]
		dir := filepath.Join(podcastRoot, ep.show)
		out := filepath.Join(dir, fileName(n, ep.title))
		if _, statErr := os.Stat(out); statErr == nil {
			fmt.Printf("skip  %s/%s (already there)\n", ep.show, filepath.Base(out))
			continue
		}
		fmt.Printf("%s/%s  ← %s (%s), dated %s\n", ep.show, filepath.Base(out), ep.source, sources[i].codec, ep.releaseDate)
		if dryRun {
			continue
		}
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err = makeEpisode(ctx, sources[i], out, ep, n, total); err != nil {
			return fmt.Errorf("%s: %w", out, err)
		}
	}
	return nil
}
