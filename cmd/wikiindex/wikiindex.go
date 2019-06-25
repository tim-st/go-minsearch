package main

import (
	"compress/bzip2"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dustin/go-wikiparse"
	"github.com/tim-st/go-minsearch"
)

func main() {

	var filename string
	var fullText bool
	var idLimit int
	var noSync bool

	flag.StringVar(&filename, "filename", "", "Filename of the MediaWiki xml.bz2 file to index.")
	flag.BoolVar(&fullText, "fullText", false, "Index also full text.")
	flag.IntVar(&idLimit, "idLimit", -1, "If idLimit>0 only the highest idLimit scores will be indexed per key.")
	flag.BoolVar(&noSync, "noSync", false, "If nosync=true indexing will be much faster but data can be lost if system crashes.")
	flag.Parse()

	if flag.NFlag() < 1 || len(filename) == 0 {
		flag.PrintDefaults()
		return
	}

	f, fErr := os.Open(filename)

	if fErr != nil {
		log.Fatal(fErr)
	}

	index, openErr := minsearch.Open(filename+".idx", noSync)

	if openErr != nil {
		log.Fatal(openErr)
	}

	bz2Reader := bzip2.NewReader(f)

	parser, parserErr := wikiparse.NewParser(bz2Reader)

	if parserErr != nil {
		log.Fatal(parserErr)
	}

	var batchPairsTitles []minsearch.Pair
	var batchPairsTexts []minsearch.Pair

	lastID, lastIDErr := index.LastID()
	lastPos := uint64(lastID)

	skipped := false
	pagesIndexed := uint64(0)

	if lastIDErr != nil {
		skipped = true
	} else {
		fmt.Printf("\rSkipping to Page with ID %d...", lastPos)
	}

	for {
		page, pageErr := parser.Next()
		if pageErr != nil {
			break
		}

		if !skipped {
			if page.ID != lastPos {
				pagesIndexed++
				continue
			}
			skipped = true
		}

		if page.Ns == 0 {
			pagesIndexed++

			if pagesIndexed%256 == 0 {
				titlePrefix := page.Title
				if len(titlePrefix) > 6 {
					titlePrefix = titlePrefix[:6]
				}
				fmt.Printf("\rIndexing Page %d with ID %d (Title-Prefix: %s)...",
					pagesIndexed, page.ID, titlePrefix)
			}

			batchPairsTitles = append(batchPairsTitles, minsearch.Pair{
				ID:   minsearch.ID(page.ID),
				Text: []byte(page.Title)})

			if fullText {

				if len(page.Revisions) > 0 && len(page.Redir.Title) == 0 {
					batchPairsTexts = append(
						batchPairsTexts, minsearch.Pair{
							ID:   minsearch.ID(page.ID),
							Text: []byte(page.Revisions[0].Text)})
				}

			}

			if len(batchPairsTitles) >= 300 {
				index.IndexBatch(batchPairsTitles, 0)
				batchPairsTitles = batchPairsTitles[:0]

				index.IndexBatch(batchPairsTexts, idLimit)
				batchPairsTexts = batchPairsTexts[:0]

				if err := index.SetLastID(uint32(page.ID)); err != nil {
					log.Fatal(err)
				}
			}

		}

	}

	index.IndexBatch(batchPairsTitles, 0)
	index.IndexBatch(batchPairsTexts, idLimit)

	if updateErr := index.UpdateStatistics(); updateErr != nil {
		log.Fatal(updateErr)
	}

	fmt.Println("\rFinished!")

}
