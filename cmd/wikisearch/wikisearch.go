package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tim-st/go-minsearch"
)

func main() {

	var filename string
	var query string
	var limit int
	var intersection bool

	flag.StringVar(&filename, "filename", "", "Filename of the index file to use.")
	flag.StringVar(&query, "query", "", "The text to search in the index file.")
	flag.IntVar(&limit, "limit", -1, "Limit the output of the result to the given number.")
	flag.BoolVar(&intersection, "intersection", false, "true = intersection set; false = union set")
	flag.Parse()

	if flag.NFlag() < 2 || len(filename) == 0 || len(query) == 0 {
		flag.PrintDefaults()
		return
	}

	if index, openErr := minsearch.Open(filename, true); openErr == nil {
		start := time.Now()
		var setOp = minsearch.Union
		if intersection {
			setOp = minsearch.Intersection
		}
		queryResults, queryErr := index.Search([]byte(query), setOp, 0)
		fmt.Printf("Took: %s\n", time.Since(start))

		if queryErr != nil {
			log.Fatal(queryErr)
		}

		for idx, result := range queryResults {
			if limit > 0 && idx == limit {
				break
			}
			fmt.Printf("Idx: %d; ID: %d; Score: %.15f\n", idx, result.ID, result.Score)
		}

	} else {
		log.Fatal(openErr)
	}

}
