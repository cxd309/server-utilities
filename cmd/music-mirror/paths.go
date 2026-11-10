package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cxd309/server-utilities/internal/pathutil"
)

var (
	errSrcNotDir = errors.New("source is not a directory")
	errDestInSrc = errors.New("mirror is inside the library, it would scan itself")
	errSrcInDest = errors.New("library is inside the mirror, prine could delete library")
)

// checkPaths makes src and dest absolute and rejects layouts where the
// library and the mirror overlap.
func checkPaths(rawSrc, rawDest string) (src, dest string, err error) {
	if src, err = filepath.Abs(rawSrc); err != nil {
		return "", "", err
	}
	if dest, err = filepath.Abs(rawDest); err != nil {
		return "", "", err
	}
	if err = checkDir(src); err != nil {
		return "", "", err
	}

	switch {
	case pathutil.Within(src, dest):
		return "", "", errDestInSrc
	case pathutil.Within(dest, src):
		return "", "", errSrcInDest
	}
	return src, dest, nil
}

// checkDir returns an error unless path exists and is a directory.
func checkDir(path string) (err error) {
	var info fs.FileInfo
	if info, err = os.Stat(path); err != nil {
		return err
	}
	if !info.IsDir() {
		return errSrcNotDir
	}
	return nil
}
