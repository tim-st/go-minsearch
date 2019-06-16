package minsearch

import (
	"testing"

	"github.com/tim-st/go-uniseg"
)

func normalizeString(s string) string {
	return string(normalizeSegment(uniseg.Segments([]byte(s))[0]))
}

func TestNormalize(t *testing.T) {
	var tests = map[string]string{
		"¼":        "1/4",
		"²":        "2",
		"Änderung": "aenderung",
		"Café":     "cafe",
		"Straße":   "strasse",
	}

	for input, expected := range tests {
		if got := normalizeString(input); got != expected {
			t.Errorf("normalizeString(%s) = %s; expected %s", input, got, expected)
		}
	}

}
