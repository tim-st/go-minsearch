package minsearch

import (
	"bytes"
	"testing"

	"github.com/tim-st/go-uniseg"
)

func normalizeString(s string) string {
	segments := uniseg.Segments([]byte(s))
	result := make([]byte, 0, len(s))
	for _, segment := range segments {
		result = append(result, normalizeSegment(segment)...)
		result = append(result, '|')
	}
	result = bytes.TrimRight(result, "|")
	return string(result)
}

func TestNormalize(t *testing.T) {
	var tests = map[string]string{
		"":            "",
		"0":           "0",
		"1234567":     "1234567",
		"12345678":    "",
		"¼":           "1/4",
		"²":           "2",
		"1A":          "1|a",
		"100jähriges": "100|jaehriges",
		"ä":           "ae",
		"Ä":           "ae",
		"ABC":         "abc",
		"aBc":         "abc",
		"AbC":         "abc",
		"Änderung":    "aenderung",
		"Café":        "cafe",
		"ß":           "ss",
		"ẞ":           "ss",
		"Straße":      "strasse",
		"Test":        "test",
	}

	for input, expected := range tests {
		if got := normalizeString(input); got != expected {
			t.Errorf("normalizeString(%s) = %s; expected %s", input, got, expected)
		}
	}

}
