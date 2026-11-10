package main

import (
	"errors"
	"fmt"
	"strings"
)

// errHelp is returned when -h/--help is given.
var errHelp = errors.New("help requested")

const usageText = `usage: ripcd [flags] [whipper options]

flags:
  -b, --barcode CODE   pick the MusicBrainz release by the case's barcode
  -d, --drive DEV      CD drive (default $RIPCD_DRIVE or /dev/sr0)
  -i, --ingest DIR     where rips go (default $RIPCD_INGEST or /mnt/dolly/downloads/music_ingest)
  -n, --dry-run        print the whipper command instead of running it
  -h, --help           show this help

Anything after these flags is passed to "whipper cd rip".

examples:
  ripcd
  ripcd --unknown
  ripcd -b 634904074029
`

type options struct {
	Barcode string
	Drive   string
	Ingest  string
	DryRun  bool
	Pass    []string // extra options passed straight through to whipper
}

// parseArgs parses ripcd's own flags and passes everything after them to
// whipper. getenv supplies defaults for the drive and ingest folder.
func parseArgs(args []string, getenv func(string) string) (options, error) {
	o := options{
		Drive:  envOr(getenv, "RIPCD_DRIVE", "/dev/sr0"),
		Ingest: envOr(getenv, "RIPCD_INGEST", "/mnt/dolly/downloads/music_ingest"),
	}

	value := func() (string, error) {
		if len(args) < 2 || strings.HasPrefix(args[1], "-") {
			return "", fmt.Errorf("%s needs a value", args[0])
		}
		v := args[1]
		args = args[2:]
		return v, nil
	}

flags:
	for len(args) > 0 {
		var err error
		switch args[0] {
		case "-b", "--barcode":
			o.Barcode, err = value()
		case "-d", "--drive":
			o.Drive, err = value()
		case "-i", "--ingest":
			o.Ingest, err = value()
		case "-n", "--dry-run":
			o.DryRun = true
			args = args[1:]
		case "-h", "--help":
			return o, errHelp
		default:
			break flags
		}
		if err != nil {
			return o, err
		}
	}

	// whipper options all start with "-"; a bare word here is a mistake.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return o, fmt.Errorf("unexpected argument %q (whipper options must start with -)", args[0])
	}

	if len(args) > 0 {
		o.Pass = args
	}
	return o, nil
}

func envOr(getenv func(string) string, key, def string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return def
}

// whipperArgs builds the full whipper argv (including "whipper" itself).
func whipperArgs(o options, releaseID string) []string {
	argv := []string{"whipper", "cd", "-d", o.Drive, "rip"}
	if releaseID != "" {
		argv = append(argv, "-R", releaseID)
	}
	argv = append(argv, "--cover-art", "complete", "--keep-going")
	return append(argv, o.Pass...)
}
