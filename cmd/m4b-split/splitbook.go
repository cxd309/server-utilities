package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// splitBook splits src into one tagged file per chapter in the folder dir,
// which must not exist yet.
// The work happens in a hidden temporary folder beside dir,
// renamed into place only once every chapter is done,
// so an interrupted run leaves nothing half-finished behind.
func splitBook(ctx context.Context, src, dir string, b book) (err error) {
	if _, err = os.Stat(dir); err == nil {
		return fmt.Errorf("%s: %w", dir, errAlreadySplit)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var work string
	if work, err = os.MkdirTemp(filepath.Dir(dir), ".m4b-split-*"); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(work)
		}
	}()

	if err = splitAudio(ctx, src, work, b.chapters); err != nil {
		return err
	}
	// A miscount must never be renamed into place looking finished
	var pieces []string
	if pieces, err = filepath.Glob(filepath.Join(work, "*.m4a")); err != nil {
		return err
	}
	if len(pieces) != len(b.chapters) {
		return fmt.Errorf("%d pieces for %d chapters: %w", len(pieces), len(b.chapters), errPieceCount)
	}
	for n := 1; n <= len(b.chapters); n++ {
		if err = tagPiece(ctx, work, b, n); err != nil {
			return fmt.Errorf("chapter %d: %w", n, err)
		}
	}

	// MkdirTemp makes the folder private to us; open it up like a normal one.
	if err = os.Chmod(work, 0o755); err != nil {
		return err
	}
	return os.Rename(work, dir)
}
