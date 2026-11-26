package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameTime(t *testing.T) {
	tests := []struct {
		name       string
		frame      int64
		sampleRate int
		want       string
	}{
		{"start", 0, 44100, "0.000000"},
		{"whole second", 1, 1024, "1.000000"},
		{"one frame at 44.1 kHz", 1, 44100, "0.023219"},
		{"rounds down", 2000, 44100, "46.439909"}, // exactly 46.4399092970…
		{"22 kHz", 1000, 22050, "46.439909"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, frameTime(tt.frame, tt.sampleRate))
		})
	}
}

func TestConcatList(t *testing.T) {
	chapters := []chapterFile{
		{path: "/b/001 One.m4a", frames: 100},
		{path: "/b/002 Can't Stop.m4a", frames: 4000},
	}
	tests := []struct {
		name    string
		part    []piece
		preroll bool
		want    string
	}{
		{"whole chapters", []piece{{0, 0, 100}}, false, "file '/b/001 One.m4a'\n"},
		{
			"quote in path, ends partway",
			[]piece{{0, 0, 100}, {1, 0, 2000}}, false,
			"file '/b/001 One.m4a'\n" +
				"file '/b/002 Can'\\''t Stop.m4a'\n" +
				"outpoint 46.439909\n",
		},
		{
			"starts partway",
			[]piece{{1, 2000, 4000}}, false,
			"file '/b/002 Can'\\''t Stop.m4a'\n" +
				"inpoint 46.439909\n", // frame 2000
		},
		{
			"starts partway with preroll",
			[]piece{{1, 2000, 4000}}, true,
			"file '/b/002 Can'\\''t Stop.m4a'\n" +
				"inpoint 46.416689\n", // frame 1999
		},
		{
			"middle of a chapter with preroll",
			[]piece{{1, 1000, 2000}}, true,
			"file '/b/002 Can'\\''t Stop.m4a'\n" +
				"inpoint 23.196734\n" + // frame 999
				"outpoint 46.439909\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, concatList(tt.part, chapters, 44100, tt.preroll))
		})
	}
}

func TestChapterMeta(t *testing.T) {
	chapters := []chapterFile{
		{title: "One", frames: 100},
		{title: "Two; a=b", frames: 4000},
	}
	part := []piece{{0, 0, 100}, {1, 0, 2000}}

	want := "[CHAPTER]\nTIMEBASE=1/44100\nSTART=0\nEND=102400\ntitle=One\n" +
		"[CHAPTER]\nTIMEBASE=1/44100\nSTART=102400\nEND=2150400\ntitle=Two\\; a\\=b\n"

	assert.Equal(t, want, chapterMeta(part, chapters, 44100))
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name      string
		cover     string
		copyAudio bool
		preroll   bool
		want      []string
	}{
		{
			"copy, no cover", "", true, false,
			[]string{
				"-f", "concat", "-safe", "0", "-i", "/tmp/x/list.txt",
				"-i", "/tmp/x/meta.txt",
				"-map", "0:a:0",
				"-map_metadata", "1", "-map_chapters", "1",
				"-c:a", "copy", "-c:v", "copy",
				"-f", "ipod", "/m/Book/Part 01.m4b",
			},
		},
		{
			"ALAC with cover and preroll", "/b/cover.jpg", false, true,
			[]string{
				"-f", "concat", "-safe", "0", "-i", "/tmp/x/list.txt",
				"-i", "/tmp/x/meta.txt",
				"-i", "/b/cover.jpg",
				"-map", "0:a:0",
				"-map", "2:v:0", "-disposition:v", "attached_pic",
				"-af", "atrim=start_sample=1024,asetpts=PTS-STARTPTS",
				"-map_metadata", "1", "-map_chapters", "1",
				"-c:a", "alac", "-sample_fmt", "s16p", "-c:v", "copy",
				"-f", "ipod", "/m/Book/Part 01.m4b",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinArgs("/tmp/x/list.txt", "/tmp/x/meta.txt", tt.cover, "/m/Book/Part 01.m4b", tt.copyAudio, tt.preroll)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCanCopy(t *testing.T) {
	tests := []struct {
		name  string
		codec string
		want  bool
	}{
		{name: "AAC is copied, whatever its sample rate", codec: "aac", want: true},
		{name: "FLAC is re-encoded", codec: "flac", want: false},
		{name: "MP3 is re-encoded", codec: "mp3", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canCopy(tt.codec))
		})
	}
}

// countSamples asks ffprobe how many samples the file's audio stream holds.
// Its time base is 1/sample rate, so the stream's duration_ts is a sample count
func countSamples(t *testing.T, path string) (samples int64) {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error",
		"-select_streams", "a:0", "-show_entries", "stream=duration_ts",
		"-of", "csv=p=0", path).Output()
	require.NoError(t, err)
	samples, err = strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	require.NoError(t, err)
	return samples
}

// countFrames asks ffprobe how many audio packets (AAC frames) the file holds
func countFrames(t *testing.T, path string) (frames int64) {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-count_packets",
		"-select_streams", "a:0", "-show_entries", "stream=nb_read_packets",
		"-of", "csv=p=0", path).Output()
	require.NoError(t, err)
	frames, err = strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	require.NoError(t, err)
	return frames
}

func TestJoinPart(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	for _, sampleRate := range []int{44100, 22050} { // AAC is copied at either rate
		t.Run(strconv.Itoa(sampleRate), func(t *testing.T) {
			ctx := context.Background()
			dir := t.TempDir()

			// A 3-second book, cut into 1-second chapters the way m4b-split does
			src := filepath.Join(dir, "book.m4a")
			require.NoError(t, ffmpeg.Run(ctx,
				"-f", "lavfi", "-i", "sine=duration=3:sample_rate="+strconv.Itoa(sampleRate),
				"-c:a", "aac", "-f", "ipod", src))
			require.NoError(t, ffmpeg.Run(ctx,
				"-i", src, "-c", "copy", "-f", "segment", "-segment_format", "ipod",
				"-reset_timestamps", "1", "-segment_times", "1,2",
				filepath.Join(dir, "ch%d.m4a")))

			var chapters []chapterFile
			for i, title := range []string{"One", "Can't Two"} {
				path := filepath.Join(dir, "ch"+strconv.Itoa(i)+".m4a")
				chapters = append(chapters, chapterFile{path: path, title: title, frames: countFrames(t, path), codec: "aac"})
			}

			cover := filepath.Join(dir, "cover.jpg")
			require.NoError(t, ffmpeg.Run(ctx,
				"-f", "lavfi", "-i", "color=c=red:s=64x64", "-frames:v", "1", cover))
			tags := tagLines(bookTags{title: "Test Book", author: "An Author"}, "Test Book, Part 1 of 2", 1, 2)

			// All of chapter 0, then the first 10 frames of chapter 1
			out := filepath.Join(dir, "Part 01.m4b")
			part := []piece{{0, 0, chapters[0].frames}, {1, 0, 10}}
			require.NoError(t, joinPart(ctx, part, chapters, sampleRate, tags, cover, out))
			assert.Equal(t, (chapters[0].frames+10)*samplesPerFrame, countSamples(t, out))

			info, err := ffmpeg.Probe(ctx, out)
			require.NoError(t, err)
			assert.Equal(t, "aac", info.Streams[0].CodecName)
			assert.Equal(t, strconv.Itoa(sampleRate), info.Streams[0].SampleRate)
			assert.Equal(t, "Test Book, Part 1 of 2", info.Format.Tags["title"])
			assert.Equal(t, "1/2", info.Format.Tags["track"])
			assert.Equal(t, "2", info.Format.Tags["media_type"])

			var hasCover bool
			for _, s := range info.Streams {
				hasCover = hasCover || s.CodecType == "video"
			}
			assert.True(t, hasCover, "cover art stream")
			require.Len(t, info.Chapters, 2)
			assert.Equal(t, "One", info.Chapters[0].Tags["title"])
			assert.Equal(t, "Can't Two", info.Chapters[1].Tags["title"])

			// A part that starts mid-chapter. Copied AAC can only start on a whole packet.
			// Depending on the ffmpeg version, the concat demuxer may also keep the frame just
			// before the inpoint: at most 23 ms heard twice at a part boundary,
			// and only in a chapter too long for one part
			mid := filepath.Join(dir, "Part 02.m4b")
			require.NoError(t, joinPart(ctx, []piece{{1, 10, 20}}, chapters, sampleRate, tags, "", mid))
			assert.InDelta(t, 10*samplesPerFrame, countSamples(t, mid), samplesPerFrame, "10 frames, give or take one")
		})
	}
}

// FLAC from CDs is converted to ALAC. Only whole chapters are tested:
// a cut inside a FLAC chapter lands on a FLAC block (4,096 samples), not a frame,
// but that only happens to a CD track too long for a part, about 6 h
func TestJoinPartFLAC(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()
	dir := t.TempDir()

	// A file per track, as ripped from a CD
	var chapters []chapterFile
	var wantSamples int64
	for i, title := range []string{"One", "Two"} {
		path := filepath.Join(dir, "ch"+strconv.Itoa(i)+".flac")
		require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "sine=duration=1:sample_rate=44100", "-c:a", "flac", path))
		samples := countSamples(t, path)
		wantSamples += samples
		frames := (samples + samplesPerFrame - 1) / samplesPerFrame // as framesFor does
		chapters = append(chapters, chapterFile{path: path, title: title, frames: frames, codec: "flac"})
	}

	out := filepath.Join(dir, "Part 1.m4b")
	part := []piece{{0, 0, chapters[0].frames}, {1, 0, chapters[1].frames}}
	tags := tagLines(bookTags{title: "Test Book"}, "Test Book", 1, 1)
	require.NoError(t, joinPart(ctx, part, chapters, 44100, tags, "", out))
	assert.Equal(t, wantSamples, countSamples(t, out))

	info, err := ffmpeg.Probe(ctx, out)
	require.NoError(t, err)
	assert.Equal(t, "alac", info.Streams[0].CodecName)
	require.Len(t, info.Chapters, 2)
	assert.Equal(t, "Two", info.Chapters[1].Tags["title"])
}
