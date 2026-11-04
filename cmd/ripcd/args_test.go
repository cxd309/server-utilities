package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultIngest = "/mnt/dolly/downloads/music_ingest"

func noEnv(string) string { return "" }

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{"defaults", nil,
			options{Drive: "/dev/sr0", Ingest: defaultIngest}},
		{"passthrough only", []string{"--unknown"},
			options{Drive: "/dev/sr0", Ingest: defaultIngest, Pass: []string{"--unknown"}}},
		{"barcode", []string{"-b", "634904074029"},
			options{Barcode: "634904074029", Drive: "/dev/sr0", Ingest: defaultIngest}},
		{"barcode, dry run, passthrough", []string{"--barcode", "123", "-n", "--unknown"},
			options{Barcode: "123", DryRun: true, Drive: "/dev/sr0", Ingest: defaultIngest,
				Pass: []string{"--unknown"}}},
		{"passthrough with a value", []string{"--max-retries", "10"},
			options{Drive: "/dev/sr0", Ingest: defaultIngest, Pass: []string{"--max-retries", "10"}}},
		{"drive and ingest", []string{"-d", "/dev/sr1", "-i", "/tmp/rips"},
			options{Drive: "/dev/sr1", Ingest: "/tmp/rips"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args, noEnv)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseArgsEnvDefaults(t *testing.T) {
	env := map[string]string{"RIPCD_DRIVE": "/dev/sr9", "RIPCD_INGEST": "/data/in"}
	got, err := parseArgs(nil, func(k string) string { return env[k] })
	require.NoError(t, err)
	assert.Equal(t, "/dev/sr9", got.Drive)
	assert.Equal(t, "/data/in", got.Ingest)
}

func TestParseArgsErrors(t *testing.T) {
	for _, args := range [][]string{
		{"-b"},              // missing value
		{"-b", "--unknown"}, // flag where value expected
		{"-d"},              // missing value
		{"Some Album"},      // stray word
		{"-n", "oops"},      // stray word after our flags
	} {
		_, err := parseArgs(args, noEnv)
		assert.Error(t, err, "args %q", args)
	}

	_, err := parseArgs([]string{"-h"}, noEnv)
	assert.ErrorIs(t, err, errHelp)
}

func TestWhipperArgs(t *testing.T) {
	assert.Equal(t,
		[]string{"whipper", "cd", "-d", "/dev/sr0", "rip", "-R", "abc",
			"--cover-art", "complete", "--keep-going", "--unknown"},
		whipperArgs(options{Drive: "/dev/sr0", Pass: []string{"--unknown"}}, "abc"))

	assert.Equal(t,
		[]string{"whipper", "cd", "-d", "/dev/sr0", "rip", "--cover-art", "complete", "--keep-going"},
		whipperArgs(options{Drive: "/dev/sr0"}, ""))
}
