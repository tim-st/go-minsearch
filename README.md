# go-minsearch

Package `minsearch` implements a minimal solution to index text and retrieve search results with score.

Documentation at <https://godoc.org/github.com/tim-st/go-minsearch>.

Download and install package `minsearch` and its tools with `go get -u github.com/tim-st/go-minsearch/...`

## Commands

### wikiindex

`wikiindex` can create a full text index of a MediaWiki `xml.bz2` dump file. The indexing process is interruptable.

#### Indexing Example

`wikiindex -filename="dewiki-20190601-pages-articles.xml.bz2" -fullText -idLimit=1000 -noSync`

creates an index file `dewiki-20190601-pages-articles.xml.bz2.idx` (file size: 3.00 GB; number segments: 13628084; avg number IDs per segment: 14.05).

### wikisearch

`wikisearch` can print sorted search results of a query searched in an indexed file created by `wikiindex`.

The ID in the search results is the page ID used by the MediaWiki (`hostname/w/index.php?curid=ID`).

#### Search Example

`wikisearch -filename="dewiki-20190601-pages-articles.xml.bz2.idx" -intersection -limit=10 -query="word1 word2 word3..."`
