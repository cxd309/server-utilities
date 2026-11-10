package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/cxd309/server-utilities/internal/pathutil"
)

type options struct {
	src     string
	dest    string
	bitrate string
	prune   bool
	dryRun  bool
	jobs    int
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("finding home folder: %v", err)
	}

	var o options
	flag.StringVar(&o.src, "src", "/Volumes/music", "source library root (mounted share)")
	flag.StringVar(&o.dest, "dest", filepath.Join(home, "Music", "music-mirror"), "mirror root")
	flag.StringVar(&o.bitrate, "bitrate", "320k", "MP3 bitrate passed to ffmpeg")
	flag.BoolVar(&o.prune, "prune", true, "delete mirror files where source has been removed")
	flag.BoolVar(&o.dryRun, "dry-run", false, "show what would happen, change nothing")
	flag.IntVar(&o.jobs, "jobs", runtime.NumCPU(), "number of parallel encodes")
	flag.Parse()

	if o.jobs < 1 {
		o.jobs = 1
	}

	src, dest, err := checkPaths(o.src, o.dest)
	if err != nil {
		fatalf("%v (--src %s, --dest %s)", err, o.src, o.dest)
	}
	o.src, o.dest = src, dest

	if err = checkFFmpeg(); err != nil {
		fatalf("%v", err)
	}

	jobs, err := scan(o.src, o.dest)
	if err != nil {
		fatalf("%v", err)
	}
	if len(jobs) == 0 {
		fatalf("no audio files found in %s", o.src)
	}

	var todo []job
	for _, j := range jobs {
		done, err := upToDate(j)
		if err != nil {
			fatalf("%v", err)
		}
		if !done {
			todo = append(todo, j)
		}
	}
	fmt.Printf("%d tracks, %d to update\n", len(jobs), len(todo))

	// Ctrl-C or a service stop cancels ctx:
	// no new jobs start, running encodes are killed, and finished tracks are kept.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	failed := 0
	if o.dryRun {
		for _, j := range todo {
			fmt.Printf("%-6s %s\n", j.kind, o.rel(j.dst))
		}
	} else {
		done := 0
		runJobs(ctx, todo, o.jobs, o.bitrate, func(j job, err error) {
			done++
			switch {
			case errors.Is(err, context.Canceled):
				// interrupted, reported once below
			case err != nil:
				failed++
				fmt.Fprintf(os.Stderr, "[%d/%d] FAILED %s: %v\n", done, len(todo), o.rel(j.src), err)
			default:
				fmt.Printf("[%d/%d] %-6s %s\n", done, len(todo), j.kind, o.rel(j.dst))
			}
		})
	}

	if ctx.Err() != nil {
		fatalf("interrupted: finished tracks are kept, the rest will be done next run")
	}

	if o.prune {
		removed, err := prune(o.dest, jobs, o.dryRun)
		for _, p := range removed {
			fmt.Printf("prune  %s\n", o.rel(p))
		}
		if err != nil {
			fatalf("pruning: %v", err)
		}
	}

	if failed > 0 {
		fatalf("%d track(s) failed, see above", failed)
	}
}

// rel shortens a path for display, relative to the source or mirror root.
func (o options) rel(path string) (short string) {
	for _, root := range []string{o.dest, o.src} {
		if r, err := filepath.Rel(root, path); err == nil && pathutil.Within(root, path) {
			return r
		}
	}
	return path
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
