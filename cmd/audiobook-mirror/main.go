// Command audiobook-mirror turns the audiobook library into iPod-safe parts:
// each book's chapters are joined, without re-encoding, into tagged .m4b parts
// of at most 79 minutes, mirroring the library's Author/Title folders.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cxd309/server-utilities/internal/pathutil"
)

func main() {
	var err error

	var o options
	flag.StringVar(&o.src, "src", "/mnt/prudence/audiobooks", "library root, holding Author/Title folders")
	flag.StringVar(&o.dest, "dest", "/mnt/dolly/mirror/audiobooks", "mirror root")
	flag.BoolVar(&o.prune, "prune", true, "delete mirrored books that are no longer in the library")
	flag.BoolVar(&o.dryRun, "dry-run", false, "show what would happen, change nothing")
	flag.Parse()

	if o, err = checkOptions(o); err != nil {
		fatalf("%v (--src %s, --dest %s)", err, o.src, o.dest)
	}
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err = exec.LookPath(tool); err != nil {
			fatalf("%s not found on PATH: install it with `brew install ffmpeg`", tool)
		}
	}

	// Ctrl-C cancels ctx: the book being built is abandoned
	// (its hidden build folder is cleaned up next run), finished books are kept
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var failed int
	failed, err = run(ctx, o, os.Stdout)
	switch {
	case errors.Is(err, context.Canceled):
		fatalf("interrupted: finished books are kept, the rest will be done next run")
	case err != nil:
		fatalf("%v", err)
	case failed > 0:
		fatalf("%d book(s) failed, see above", failed)
	}
}

// checkOptions makes src and dest absolute,
// and rejects a src that isn't a folder or a layout where the library and mirror overlap
func checkOptions(raw options) (o options, err error) {
	o = raw
	if o.src, err = filepath.Abs(raw.src); err != nil {
		return raw, err
	}
	if o.dest, err = filepath.Abs(raw.dest); err != nil {
		return raw, err
	}

	var info fs.FileInfo
	if info, err = os.Stat(o.src); err != nil {
		return raw, err
	}
	if !info.IsDir() {
		return raw, errSrcNotDir
	}

	switch {
	case pathutil.Within(o.src, o.dest):
		return raw, errDestInSrc
	case pathutil.Within(o.dest, o.src):
		return raw, errSrcInDest
	}
	return o, nil
}

// fatalf prints an error and exits with status 1
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
