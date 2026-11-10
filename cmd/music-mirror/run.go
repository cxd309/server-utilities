package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var errUnknownKind = errors.New("unknown job kind")

// process brings one job's mirror file up to date.
// All work happens in a .part file beside the destination,
// which is renamed into place only once complete,
// so an interupted run never leaves a half-written mp3 that looks finished.
func process(ctx context.Context, j job, bitrate string) (err error) {
	if err = os.MkdirAll(filepath.Dir(j.dst), 0o755); err != nil {
		return err
	}

	part := j.dst + ".part"
	defer func() {
		if err != nil {
			_ = os.Remove(part)
			if ctx.Err() != nil {
				err = ctx.Err() // report "cancelled", not ffmpeg being killed
			}
		}
	}()

	var tags trackTags
	switch j.kind {
	case kindEncode:
		if tags, err = readFLAC(j.src); err != nil {
			return err
		}
		if err = encode(ctx, j.src, part, bitrate); err != nil {
			return err
		}
	case kindCopy:
		if tags, err = readMP3(j.src); err != nil {
			return err
		}
		if err = copyFile(j.src, part); err != nil {
			return err
		}
	default:
		return errUnknownKind
	}

	if err = writeID3(part, tags); err != nil {
		return err
	}
	return os.Rename(part, j.dst)
}

// copyFile copies src to dst, creating or truncating dst.
func copyFile(src, dst string) (err error) {
	var in, out *os.File
	if in, err = os.Open(src); err != nil {
		return err
	}
	defer in.Close()

	if out, err = os.Create(dst); err != nil {
		return err
	}
	defer func() {
		// A failed Close on a written file can mean the data never landed.
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}

// runJobs processes jobs using a given number of workers, calling onDone after each one.
// onDone is never called concurrently, so it doesn't need its own locking.
// Cancelling ctx stops new jobs starting and kills any running encodes.
func runJobs(ctx context.Context, jobs []job, workers int, bitrate string, onDone func(j job, err error)) {
	queue := make(chan job)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for range workers {
		wg.Go(func() {
			for j := range queue {
				err := process(ctx, j, bitrate)
				mu.Lock()
				onDone(j, err)
				mu.Unlock()
			}
		})
	}

feed:
	for _, j := range jobs {
		select {
		case queue <- j:
		case <-ctx.Done():
			break feed
		}
	}
	close(queue)
	wg.Wait()
}
