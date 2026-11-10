// Command ripcd rips a CD with whipper into an ingest folder,
// optionally using the case's barcode to pick the exact MusicBrainz release.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cxd309/server-utilities/internal/discid"
	"github.com/cxd309/server-utilities/internal/musicbrainz"
)

const userAgent = "ripcd/0.1 (personal CD ripping tool)"

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errHelp) {
			fmt.Print(usageText)
			return
		}
		fmt.Fprintln(os.Stderr, "ripcd:", err)
		os.Exit(1)
	}
}

func run() error {
	o, err := parseArgs(os.Args[1:], os.Getenv)
	if err != nil {
		if !errors.Is(err, errHelp) {
			fmt.Fprint(os.Stderr, usageText)
		}
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	releaseID := ""
	if o.Barcode != "" {
		releaseID, err = chooseRelease(ctx, o, musicbrainz.New(userAgent), os.Stdin, os.Stdout)
		if err != nil {
			return err
		}
	}

	argv := whipperArgs(o, releaseID)
	if o.DryRun {
		fmt.Printf("\ncd %s && %s\n", shellQuote(o.Ingest), shellJoin(argv))
		return nil
	}
	return execWhipper(o.Ingest, argv)
}

// chooseRelease looks up the barcode and returns the chosen release ID ("" = none).
func chooseRelease(ctx context.Context, o options, mb *musicbrainz.Client, in io.Reader, out io.Writer) (string, error) {
	onDisc := map[string]bool{}
	if toc, err := discid.ReadTOC(o.Drive); err != nil {
		fmt.Fprintln(out, "warning: couldn't read the disc's table of contents:", err)
	} else {
		id := toc.DiscID()
		fmt.Fprintln(out, "Disc ID:", id)
		ids, err := mb.ReleaseIDsByDiscID(ctx, id)
		if err != nil {
			fmt.Fprintln(out, "warning: disc ID lookup failed:", err)
		}
		for _, rid := range ids {
			onDisc[rid] = true
		}
	}

	rels, err := mb.ReleasesByBarcode(ctx, o.Barcode)
	if err != nil {
		return "", fmt.Errorf("barcode lookup: %w", err)
	}
	if len(rels) == 0 {
		fmt.Fprintf(out, "No MusicBrainz release has barcode %s; using the disc ID only.\n", o.Barcode)
		return "", nil
	}

	fmt.Fprintln(out)
	printReleases(out, rels, onDisc)
	fmt.Fprintln(out)

	i, ok := autoPick(rels, onDisc)
	if !ok {
		if i, err = promptPick(in, out, len(rels)); err != nil {
			return "", err
		}
		if i < 0 {
			return "", nil
		}
	} else {
		fmt.Fprintf(out, "Using #%d.\n", i+1)
	}

	fmt.Fprintln(out, "Release page: https://musicbrainz.org/release/"+rels[i].ID)
	return rels[i].ID, nil
}

// execWhipper replaces this process with whipper, run from the ingest folder,
// so whipper owns the terminal (prompts, Ctrl+C).
func execWhipper(ingest string, argv []string) error {
	if err := os.MkdirAll(ingest, 0o755); err != nil {
		return err
	}
	if err := os.Chdir(ingest); err != nil {
		return err
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("whipper not found on PATH: %w", err)
	}
	fmt.Println()
	return syscall.Exec(path, argv, os.Environ())
}

func shellJoin(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		q[i] = shellQuote(a)
	}
	return strings.Join(q, " ")
}

func shellQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t'\"$&;|<>()*?[]#~`\\") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
