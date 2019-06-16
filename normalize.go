package minsearch

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/tim-st/go-uniseg"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func normalizeSegment(segment uniseg.Segment) []byte {
	const maxRunes = 30
	if segment.RuneCount > maxRunes {
		return nil
	}
	switch segment.Category {
	case uniseg.UnicodeNd, uniseg.UnicodeNl, uniseg.UnicodeNo:
		if segment.RuneCount <= 7 {
			return normalizeN(segment.Segment)
		}
	case uniseg.WordAllLower, uniseg.UnicodeLl:
		return normalizeLlWl(segment.Segment)
	case uniseg.WordFirstUpper, uniseg.WordAllUpper, uniseg.WordMixedLetters,
		uniseg.UnicodeLm, uniseg.UnicodeLo, uniseg.UnicodeLt, uniseg.UnicodeLu:
		return normalizeWfWaWmLmLoLtLu(segment.Segment)
	}
	return nil
}

func normalizeN(n []byte) []byte {
	if len(bytes.TrimLeftFunc(n, func(r rune) bool {
		switch {
		case r >= '0' && r <= '9':
			return true
		default:
			return false
		}
	})) == 0 {
		return n
	}
	return normalize(n)
}

func normalizeLlWl(l []byte) []byte {
	if len(bytes.TrimLeftFunc(l, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return true
		default:
			return false
		}
	})) == 0 {
		return l
	}
	return normalize(l)
}

func normalizeWfWaWmLmLoLtLu(l []byte) []byte {
	if len(bytes.TrimLeftFunc(l, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return true
		case r >= 'A' && r <= 'Z':
			return true
		default:
			return false
		}
	})) == 0 {
		return bytes.ToLower(l)
	}
	return normalize(l)
}

func normalize(b []byte) []byte {
	result, _, err := transform.Bytes(t, b)
	if err != nil {
		return nil
	}
	return result
}

var t = transform.Chain(
	norm.NFKC,
	multiRuneTransformer{
		// case-sensitive
		'ä': "ae",
		'ö': "oe",
		'ü': "ue",
		'Ä': "ae",
		'Ö': "oe",
		'Ü': "ue",
	},
	norm.NFD,
	runes.Remove(runes.In(unicode.M)),
	runes.Map(unicode.ToLower),
	multiRuneTransformer{
		// not case-sensitive
		'⁄': "/",
		'æ': "ae",
		'ð': "d",
		'ł': "l",
		'ø': "oe",
		'œ': "oe",
		'ß': "ss",
		'þ': "th",
	})

type multiRuneTransformer map[rune]string

func (multiRuneTransformer) Reset() {}

func (m multiRuneTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		if !atEOF && !utf8.FullRune(src[nSrc:]) {
			err = transform.ErrShortSrc
			return
		}

		r, width := utf8.DecodeRune(src[nSrc:])
		if d, shouldTransform := m[r]; shouldTransform {
			if nDst+len(d) > len(dst) {
				err = transform.ErrShortDst
				return
			}
			copy(dst[nDst:], d)
			nSrc += width
			nDst += len(d)
			continue
		}

		if nDst+width > len(dst) {
			err = transform.ErrShortDst
			return
		}
		copy(dst[nDst:], src[nSrc:nSrc+width])
		nDst += width
		nSrc += width
	}
	return
}
