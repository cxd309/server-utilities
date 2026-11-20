package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// newestIn is the latest modification time of dir, its subfolders and their files.
// Folders count too, because deleting a file only changes its folder's time.
// Hidden files and folders are skipped
func newestIn(dir string) (newest time.Time, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != dir && hidden(d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return newest, nil
}

// upToDate reports whether the mirror folder for b exists
// and is at least as new as everything in the source folder
func upToDate(b book) (done bool, err error) {
	var dstInfo fs.FileInfo
	if dstInfo, err = os.Stat(b.dst); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	var newest time.Time
	if newest, err = newestIn(b.src); err != nil {
		return false, err
	}
	return !dstInfo.ModTime().Before(newest), nil
}
