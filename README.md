# go-minsearch

Package `minsearch` implements a minimal solution to index text and retrieve search results with score.

Documentation at <https://godoc.org/github.com/tim-st/go-minsearch>.

Download package `minsearch` with `go get -u github.com/tim-st/go-minsearch/...`

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/tim-st/go-minsearch"
)

func main() {
	const noSync = true
	index, indexErr := minsearch.Open("index.idx", noSync)
	if indexErr != nil {
		log.Fatal(indexErr)
	}

	var texts = [...]string{
		"First text to index", // "first", "text", "to", "index"
		"Second text 2.",      // "second", "text", "2"
		"third-",               // "third"
		"...",
	}

	maxDifferentIDsPerKeySegment := 1000 // 0 = unlimited = max file size

	for i := 0; i < 10000; i++ {
		id := i % 120 // ID can be derived from unique position or FNV hash etc.
		text := texts[i%4]
		// better: index.IndexBatch(...)
		err := index.IndexPair(minsearch.Pair{
			ID:   uint32(id),
			Text: []byte(text),
		}, maxDifferentIDsPerKeySegment)
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = index.UpdateStatistics() // not required

	maxResultsDuringCalculation := 0 // unlimited

	results, queryErr := index.Search([]byte("First Text"), minsearch.Union, maxResultsDuringCalculation)
	// results is sorted by score

	if queryErr != nil {
		log.Fatal(queryErr)
	}

	for idx, result := range results {
		// you have to store semantics about result.ID on your own.
		fmt.Println(idx, result.ID, result.Score)
	}

}
```