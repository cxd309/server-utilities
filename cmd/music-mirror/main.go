package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
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
	var o options
	flag.StringVar(&o.src, "src", "/mnt/prudence/music", "source library root")
	flag.StringVar(&o.dest, "dest", "/mnt/prudence/music-mirror", "mirror root")
	flag.StringVar(&o.bitrate, "bitrate", "320k", "MP3 bitrate passed to ffmpeg")
	flag.BoolVar(&o.prune, "prune", true, "delete mirror files where source has been removed")
	flag.BoolVar(&o.dryRun, "dry-run", false, "show what would happen, change nothing")
	flag.IntVar(&o.jobs, "jobs", runtime.NumCPU(), "number of parallel encodes")
	flag.Parse()

	src, dest, err := checkPaths(o.src, o.dest)
	if err != nil {
		fatalf("%v (--src %s, --dest %s)", err, o.src, o.dest)
	}
	o.src, o.dest = src, dest

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

	if o.dryRun {
		for _, j := range todo {
			fmt.Printf("%-6s %s\n", j.kind, j.dst)
		}
		return
	}
	fatalf("encoding not implemented yet, use --dry-run")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
