package minsearch

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/boltdb/bolt"
	"github.com/tim-st/go-uniseg"
)

// Pair is a pair of an ID and the text which should get indexed for the ID.
type Pair struct {
	ID   ID
	Text []byte
}

// IndexPair indexes all relevant segments of the given Pair.
// If maxIDs > 0 each indexed segment will only have up to maxIDs
// different (ID, Score) pairs and only the highest are chosen.
// If maxIDs > 0 and the value is chosen too small,
// the results could become too bad.
// Maybe maxIDs in [1000, 10000] is a good choice that limits
// too common words of a language like "the" or "a" in English.
// If maxIDs <= 0 the number of scores per segment is not limited.
// This will yield the best results (under the assumption that
// the result set is not limited) but definetly the biggest file size
// and higher temporary memory usage.
func (f *File) IndexPair(pair Pair, maxIDs int) error {
	var pairs = [1]Pair{pair}
	return f.IndexBatch(pairs[:], maxIDs)
}

// IndexBatch indexes all relevant segments for each Pair as a batch operation.
// See IndexPair for more information.
func (f *File) IndexBatch(pairs []Pair, maxIDs int) error {
	return f.db.Update(func(tx *bolt.Tx) error {
		relevantSegments := make(map[string]Score)
		bucket := tx.Bucket([]byte{bucketWords})
		for _, pair := range pairs {
			// idiom optimized by compiler since go 1.11
			for k := range relevantSegments {
				delete(relevantSegments, k)
			}
			segments := uniseg.Segments(pair.Text)
			for _, segment := range segments {
				if norm := normalizeSegment(segment); len(norm) > 0 {
					relevantSegments[string(norm)]++
				}
			}

			segmentsLen := Score(len(segments))

			for element, count := range relevantSegments {
				score := 1 + (count / segmentsLen)
				keyData := bucket.Get([]byte(element))
				var targetIDIndex = -1
				for i := 0; i < len(keyData); i += sizePair {
					currentID := ID(binary.LittleEndian.Uint32(keyData[i : i+sizeID]))
					if currentID == pair.ID {
						targetIDIndex = i
						break
					}
				}
				var newData []byte
				if targetIDIndex == -1 {
					newData = make([]byte, len(keyData)+sizePair)
					copy(newData, keyData)
					binary.LittleEndian.PutUint32(
						newData[len(keyData):len(keyData)+sizeID], pair.ID)
					binary.LittleEndian.PutUint32(
						newData[len(keyData)+sizeID:len(keyData)+sizePair], math.Float32bits(score))
				} else {
					prevScore := *(*Score)(unsafe.Pointer(&keyData[targetIDIndex+sizeID]))
					if prevScore < score {
						newData = make([]byte, len(keyData))
						copy(newData, keyData)
						binary.LittleEndian.PutUint32(
							newData[targetIDIndex+sizeID:targetIDIndex+sizePair], math.Float32bits(score))
					}
				}

				if len(newData) > 0 {
					if maxIDs > 0 && len(newData)>>3 > maxIDs {
						results := ((*[(1 << 31) - 1]Result)(unsafe.Pointer(&newData[0])))[:len(newData)>>3]
						sortResults(results)
						newData = newData[:maxIDs<<3]
					}
					if err := bucket.Put([]byte(element), newData); err != nil {
						return err
					}
				}

			}

		}
		return nil
	})
}
