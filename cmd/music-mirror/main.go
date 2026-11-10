package main

import (
	"flag"
	"fmt"
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
	flag.StringVar(&o.src, "src", "/mnt/prudence/music", "source library root (repeatable)")
	flag.StringVar(&o.dest, "dest", "/mnt/prudence/music-mirror", "mirror root")
	flag.StringVar(&o.bitrate, "bitrate", "350k", "MP3 bitrate passed to ffmpeg")
	flag.BoolVar(&o.prune, "prune", true, "delete mirror files where source has been removed")
	flag.BoolVar(&o.dryRun, "dry-run", false, "show what would happen, change nothing")
	flag.IntVar(&o.jobs, "jobs", runtime.NumCPU(), "number of parallel encodes")
	flag.Parse()

	fmt.Printf("%+v\n", o) // temporary
}
