package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// kind is what needs doing to get a source file into the mirror
type kind int

const (
	kindEncode kind = iota // FLAC -> mp3
	kindCopy               // mp3 -> mp3 (copy)
)

func (k kind) String() string {
	switch k {
	case kindEncode:
		return "encode"
	case kindCopy:
		return "copy"
	}
	return "unknown"
}

// job is one source file and where it ends up in the mirror
type job struct {
	src  string
	dst  string
	kind kind
}

var kindByExt = map[string]kind{
	".flac": kindEncode,
	".mp3":  kindCopy,
}

// scan walks src and returns one job per audio file.
// Folder structure is reproduced under dest, with the extension changed to .mp3.
// Hidden files and folders are skipped
func scan(src, dest string) (jobs []job, err error) {
	seen := make(map[string]string) // dst -> src, to catch collisions

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != src && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		k, ok := kindByExt[strings.ToLower(ext)]
		if !ok {
			return nil // cover.jpg, .cue, .log etc
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dest, strings.TrimSuffix(rel, ext)+".mp3")

		if prev, dup := seen[dst]; dup {
			fmt.Fprintf(os.Stderr, "warning: %s and %s both map to %s, skipping the second\n", prev, path, dst)
			return nil
		}
		seen[dst] = path
		jobs = append(jobs, job{src: path, dst: dst, kind: k})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", src, err)
	}
	return jobs, nil
}

// upToDate reports whether j.dst exists and is at least as new as j.src,
// meaning an earlier run already produced it
func upToDate(j job) (done bool, err error) {
	var dstInfo, srcInfo fs.FileInfo
	if dstInfo, err = os.Stat(j.dst); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if srcInfo, err = os.Stat(j.src); err != nil {
		return false, err
	}
	return !dstInfo.ModTime().Before(srcInfo.ModTime()), nil
}
