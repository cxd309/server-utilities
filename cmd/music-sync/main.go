package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("finding home folder: %v", err)
	}

	var mirror string
	var dryRun bool
	flag.StringVar(&mirror, "mirror", filepath.Join(home, "Music", "music-mirror"), "folder written by music-mirror")
	flag.BoolVar(&dryRun, "dry-run", false, "show what would change, change nothing")
	flag.Parse()

	if mirror, err = filepath.Abs(mirror); err != nil {
		fatalf("%v", err)
	}

	// Recorded before looking at anything,
	// so a file changed while run is newer than this and gets refreshed next time.
	started := time.Now()

	files, err := listMirror(mirror)
	if err != nil {
		fatalf("reading mirror: %v", err)
	}
	if len(files) == 0 {
		// Without this, an empty or missing mirror would remove everything.
		fatalf("no MP3s in %s (run music-mirror first)", mirror)
	}

	lastSync, err := readLastSync(mirror)
	if err != nil {
		fatalf("reading last sync time: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracks, err := listTracks(ctx)
	if err != nil {
		fatalf("asking the Music app for its library: %v", err)
	}

	p := plan(mirror, tracks, files, lastSync)
	fmt.Printf("%d in mirror, %d in library: %d to add, %d to refresh, %d to remove\n",
		len(files), len(tracks), len(p.add), len(p.refresh), len(p.remove))

	// Refresh and remove work on IDs, show paths so the output means something.
	pathOf := make(map[string]string, len(tracks))
	for _, t := range tracks {
		pathOf[t.id] = t.path
	}
	show := func(id string) (label string) {
		if pathOf[id] == "" {
			return "(file missing) " + id
		}
		return pathOf[id]
	}

	if dryRun {
		for _, path := range p.add {
			fmt.Println("add     ", path)
		}
		for _, id := range p.refresh {
			fmt.Println("refresh ", show(id))
		}
		for _, id := range p.remove {
			fmt.Println("remove  ", show(id))
		}
		return
	}

	if err = inBatches(ctx, removeScript, p.remove, nil); err != nil {
		fatalf("removing tracks: %v", err)
	}
	err = inBatches(ctx, addScript, p.add, func(done int) {
		fmt.Printf("added %d/%d\n", done, len(p.add))
	})
	if err != nil {
		fatalf("adding tracks: %v", err)
	}
	if err = inBatches(ctx, refreshScript, p.refresh, nil); err != nil {
		fatalf("refreshing tracks: %v", err)
	}

	if err = writeLastSync(mirror, started); err != nil {
		fatalf("saving sync time: %v", err)
	}
	fmt.Println("library is up to date: plug in the iPod and sync it in Finder")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
