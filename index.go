package minsearch

import (
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
// different (ID, Score) pairs and only the highest scores are chosen.
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

			const idNotFound = -1
			segmentsLen := Score(len(segments))
			for element, count := range relevantSegments {
				score := 1 + (count / segmentsLen)
				oldResultsData := bucket.Get([]byte(element))
				oldResults := asResults(oldResultsData)

				var targetIDIndex = idNotFound
				for idx, r := range oldResults {
					if r.ID == pair.ID {
						targetIDIndex = idx
						break
					}
				}

				var newResultsData []byte
				if targetIDIndex == idNotFound {
					if maxIDs <= 0 || len(oldResultsData) <= (maxIDs<<3) {
						newResultsData = make([]byte, len(oldResultsData)+sizeResult)
						copy(newResultsData, oldResultsData)
						newResults := asResults(newResultsData)
						newResults[len(newResults)-1].ID = pair.ID
						newResults[len(newResults)-1].Score = score
					}
				} else if prevScore := oldResults[targetIDIndex].Score; score > prevScore {
					newResultsData = make([]byte, len(oldResultsData))
					copy(newResultsData, oldResultsData)
					newResults := asResults(newResultsData)
					newResults[targetIDIndex].Score = score
				}

				if len(newResultsData) > 0 {
					if maxIDs > 0 && len(newResultsData) > (maxIDs<<3) {
						results := asResults(newResultsData)
						sortResults(results)
						newResultsData = newResultsData[:len(newResultsData)-sizeResult] // remove last result
					}
					if err := bucket.Put([]byte(element), newResultsData); err != nil {
						return err
					}
				}

			}

		}
		return nil
	})
}
