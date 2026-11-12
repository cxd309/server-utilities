package main

import "strings"

// lowerKeys returns tags keys converted to lowercase.
// ffprobe keeps each format's own spelling, ie.g. "title" but also "SUBTITLE"
func lowerKeys(tags map[string]string) (lower map[string]string) {
	lower = make(map[string]string, len(tags))
	for k, v := range tags {
		lower[strings.ToLower(k)] = v
	}
	return lower
}

// tidy trims s and collapses runs of whitespace,
// e.g. "Chapter  2." becomes "Chapter 2."
func tidy(s string) (clean string) {
	return strings.Join(strings.Fields(s), " ")
}
