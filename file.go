// Package minsearch implements a minimal solution to index text and retrieve search results with score.
package minsearch

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// ID is a unique uint32 number like a position or an FNV hash,
// which is indexed together with a Score.
type ID = uint32

// Score is a priority value calculated for each indexed segment per ID.
type Score = float32

const sizeID = 4
const sizeScore = 4
const sizeResult = sizeID + sizeScore

const (
	bucketStats byte = iota
	bucketWords
)

// File is the index file.
type File struct {
	db       *bolt.DB
	keyCount uint32
	avgCount float32
}

// Open opens the File or creates a new File if it doesn't exist.
// Setting the noSync flag will cause the database to skip fsync()
// calls after each commit. In the event of a system failure
// data can get lost, so setting it is unsafe but makes indexing much faster.
func Open(filename string, noSync bool) (*File, error) {
	var f = &File{}
	var err error
	f.db, err = bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	f.db.NoSync = noSync
	err = f.db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte{bucketWords})
		if e != nil {
			return e
		}
		_, e = tx.CreateBucketIfNotExists([]byte{bucketStats})
		return e
	})

	if err != nil {
		return nil, err
	}

	keyCount, keyCountErr := f.KeyCount()
	if keyCountErr == nil {
		f.keyCount = keyCount
	}

	avgCount, avgCountErr := f.AvgCount()
	if avgCountErr == nil {
		f.avgCount = avgCount
	}

	return f, nil
}

// Close closes the file.
func (f *File) Close() {
	f.db.Close()
}

func (f File) String() string {
	return fmt.Sprintf("File{KeyCount: %d, AvgCount: %.2f}", f.keyCount, f.avgCount)
}

// TODO: compact bolt.DB
