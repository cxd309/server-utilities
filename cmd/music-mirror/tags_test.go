package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinValues(t *testing.T) {
	assert.Equal(t, "Rock; Pop", joinValues([]string{"Rock", " Pop "}))
	assert.Equal(t, "Rock", joinValues([]string{"", "Rock", "  "}))
	assert.Equal(t, "", joinValues(nil))
}

func TestYearOf(t *testing.T) {
	assert.Equal(t, "1968", yearOf("1968-10-16", "2012-09-10"), "first usable date wins")
	assert.Equal(t, "2012", yearOf("", "2012-09-10"), "falls back when original is missing")
	assert.Equal(t, "1999", yearOf("1999"))
	assert.Equal(t, "", yearOf("n/a", ""))
}
