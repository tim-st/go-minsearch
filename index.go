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

			const idxNotFound = -1
			segmentsLen := Score(len(segments))
			for element, count := range relevantSegments {
				oldResultsData := bucket.Get([]byte(element))
				oldResults := asResults(oldResultsData)
				score := 1 + (count / segmentsLen)

				if maxIDs > 0 && len(oldResults) >= maxIDs && oldResults[len(oldResults)-1].Score > score {
					continue
				}

				var oldResultIdx = idxNotFound
				var newResultIdx = idxNotFound
				for idx, r := range oldResults {

					if newResultIdx == idxNotFound && (score > r.Score ||
						(score == r.Score && pair.ID < r.ID)) {
						newResultIdx = idx
					}

					if r.ID == pair.ID {
						oldResultIdx = idx
						break
					}

				}

				if newResultIdx == idxNotFound {
					newResultIdx = len(oldResults)
				}

				var newResultsData []byte
				if oldResultIdx == idxNotFound {

					newResultsData = make([]byte, len(oldResultsData)+sizeResult)
					newResults := asResults(newResultsData)

					copy(newResults, oldResults[:newResultIdx])
					newResults[newResultIdx].ID = pair.ID
					newResults[newResultIdx].Score = score
					copy(newResults[newResultIdx+1:], oldResults[newResultIdx:])

					if maxIDs > 0 && len(newResults) > maxIDs {
						newResultsData = newResultsData[:len(newResultsData)-sizeResult] // remove last result
					}

				} else if prevScore := oldResults[oldResultIdx].Score; score > prevScore {
					newResultsData = make([]byte, len(oldResultsData))
					newResults := asResults(newResultsData)

					copy(newResults, oldResults[:newResultIdx])
					newResults[newResultIdx].ID = pair.ID
					newResults[newResultIdx].Score = score
					copy(newResults[newResultIdx+1:], oldResults[newResultIdx:oldResultIdx])
					copy(newResults[oldResultIdx+1:], oldResults[oldResultIdx+1:])

				}

				if len(newResultsData) > 0 {
					if err := bucket.Put([]byte(element), newResultsData); err != nil {
						return err
					}
				}

			}

		}
		return nil
	})
}
