# go-minsearch

Package `minsearch` implements a minimal solution to index text and retrieve search results with score.

Documentation at <https://godoc.org/github.com/tim-st/go-minsearch>.

Download and install package `minsearch` and its tools with `go get -u github.com/tim-st/go-minsearch/...`

## Commands

### wikiindex

`wikiindex` can create a full text index of a MediaWiki `xml.bz2` dump file. The indexing process is interruptable.

#### Indexing Example

`wikiindex -filename="dewiki-20190601-pages-articles.xml.bz2" -fullText -idLimit=1000 -noSync`

### wikisearch

`wikisearch` can print sorted search results of a query searched in an indexed file created by `wikiindex`.

#### Search Example

`wikisearch -filename="dewiki-20190601-pages-articles.xml.bz2.idx" -intersection -limit=10 -query="word1 word2 word3..."`
