package main

import (
	"errors"
	"fmt"
)

var (
	errNoManifest    = errors.New("give the manifest's path: podcast-import [--dry-run] episodes.csv")
	errMissingColumn = errors.New("manifest is missing a column")
	errEmptyField    = errors.New("field is empty")
	errNoAudio       = errors.New("source has no audio stream")
)

// rowError is returned for a bad row in the manifest. Row counts from 1 at the header
type rowError struct {
	Row int
	Err error
}

func (e *rowError) Error() string {
	return fmt.Sprintf("manifest row %d: %v", e.Row, e.Err)
}

func (e *rowError) Unwrap() error { return e.Err }
