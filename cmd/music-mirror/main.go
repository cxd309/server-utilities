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

	jobs, err := scan(o.src, o.dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(jobs) == 0 {
		fmt.Fprintf(os.Stderr, "error: no audio files found in %s\n", o.src)
		os.Exit(1)
	}

	for _, j := range jobs {
		fmt.Printf("%-6s %s\n", j.kind, j.dst)
	}
	fmt.Printf("%d tracks\n", len(jobs))
}
