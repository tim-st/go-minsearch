package minsearch

import (
	"bytes"

	"github.com/tim-st/go-uniseg"
)

func normalizeSegment(segment uniseg.Segment) []byte {
	const maxRunes = 30
	if segment.RuneCount > maxRunes {
		return nil
	}
	switch segment.Category {
	case uniseg.UnicodeNd:
		if segment.RuneCount <= 7 {
			return segment.Segment
		}
	case uniseg.WordAllLower, uniseg.UnicodeLl:
		return segment.Segment
	case uniseg.WordFirstUpper, uniseg.WordAllUpper, uniseg.WordMixedLetters,
		uniseg.UnicodeLm, uniseg.UnicodeLo, uniseg.UnicodeLt, uniseg.UnicodeLu:
		return bytes.ToLower(segment.Segment)
	}
	return nil
}

// TODO: better word normalization (ß -> ss, é -> e, ...)
