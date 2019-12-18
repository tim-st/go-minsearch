package minsearch

import (
	"sort"
	"unsafe"

	"github.com/boltdb/bolt"
	"github.com/tim-st/go-uniseg"
)

// SetOperation is the operation that is done on the result set
// when the query consists of multiple relevant segments.
type SetOperation uint8

const (
	// Union collects all search results that match at least one relevant segment of the query.
	Union SetOperation = iota
	// Intersection collects all results that match each relevant segment of the query.
	Intersection
)

// Result is a single search result of a result set.
// It stores the ID and the score depending on the search query.
type Result struct {
	ID    ID
	Score Score
}

// DefaultMaxResults is a default value for the maximum temporary results during calculation of a search.
const DefaultMaxResults = 1000000

// Search searches the relevant segments of the query in the index file
// and returns a result set ordered by score.
// If maxResults > 0 the maximum temporary results _during_ calculation
// of the search results, which can be much higher than the end result, are limited to maxResults.
// If for at least one segment the number of results > maxResults it's possible
// that the result set misses results with higher score.
// If maxResults <= 0 the memory is not limited.
// It's recommend to set maxResults > 0 to limit the maximum RAM usage
// (especially if the SetOperation is set to Union or query is user input).
func (f *File) Search(query []byte, setOp SetOperation, maxResults int) ([]Result, error) {
	var qr = make(map[ID]Score, 1024) // TODO: cap
	err := f.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{bucketWords})
		segments := uniseg.Segments(query)
		for _, segment := range segments {
			element := normalizeSegment(segment)
			if len(element) == 0 {
				continue
			}
			results := asResults(bucket.Get(element))
			switch setOp {
			case Union:
				union(results, qr, maxResults)
			case Intersection:
				intersection(results, qr, maxResults)
				if len(qr) == 0 {
					return nil
				}
			}
		}
		return nil
	})
	var results = make([]Result, 0, len(qr))
	for id, score := range qr {
		results = append(results, Result{ID: id, Score: score})
	}
	sortResults(results)
	return results, err
}

func union(results []Result, qr map[ID]Score, maxResults int) {
	numberIDs := Score(len(results))
	for _, r := range results {
		if maxResults < 1 || len(qr) < maxResults {
			qr[r.ID] += 1 + r.Score/numberIDs
		}
	}
}

func intersection(results []Result, qr map[ID]Score, maxResults int) {
	isFirst := len(qr) == 0
	numberIDs := Score(len(results))
	for _, r := range results {
		if prevScore, currentIDExists := qr[r.ID]; currentIDExists ||
			(isFirst && (maxResults < 1 || len(qr) < maxResults)) {
			currentScore := 1 + r.Score/numberIDs
			qr[r.ID] = (prevScore + currentScore) * -1 // mark as matched again
		}
	}

	for id, score := range qr {
		if score >= 0 {
			delete(qr, id) // didn't match in last round: delete from set
		} else {
			qr[id] *= -1 // unmark the matched elements
		}
	}
}

// asResults casts the given byte slice to a Result slice.
// The Result slice will have an incorrect capacity value,
// so appending to the Result slice is not allowed!
// Modifying the Result slice is ok, if modifying the
// byte slice is ok.
func asResults(data []byte) []Result {
	if len(data) == 0 {
		return nil
	}
	return ((*[1 << 27]Result)(unsafe.Pointer(&data[0])))[:len(data)>>3]
}

func sortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].ID < results[j].ID
		}
		return results[i].Score > results[j].Score // reversed
	})
}

// TODO: suggest keys by edit distance
