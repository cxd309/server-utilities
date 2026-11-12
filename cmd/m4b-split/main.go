// Command m4b-split splits a chaptered audiobook file into a set of tagged files per chapter,
// ready for Picard.
// No re-encoding performed, only split
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	var out string
	var dryRun bool
	flag.StringVar(&out, "out", "", "folder to write the chapters to (default: beside the book, named after it)")
	flag.BoolVar(&dryRun, "dry-run", false, "show what would be written, change nothing")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: m4b-split [--out DIR] [--dry-run] BOOK.m4b")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	src := flag.Arg(0)
	if out == "" {
		out = strings.TrimSuffix(src, filepath.Ext(src))
	}
	if out == src {
		fatalf("%s has no extension, so give the output folder with --out", src)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	b, err := probeBook(ctx, src)
	if err != nil {
		fatalf("%s: %v", src, err)
	}

	if dryRun {
		fmt.Printf("%s/\n", out)
		for n, c := range b.chapters {
			fmt.Printf("  %s\n", chapterFileName(n+1, len(b.chapters), c.title))
		}
		return
	}

	if err = splitBook(ctx, src, out, b); err != nil {
		fatalf("%v", err)
	}
	fmt.Printf("%d chapters -> %s\n", len(b.chapters), out)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
