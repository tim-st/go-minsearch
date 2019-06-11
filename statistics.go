package minsearch

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/boltdb/bolt"
)

const (
	dbStatsLastID   = `lastID`
	dbStatsAvgCount = `avgCount`
	dbStatsKeyCount = `keyCount`
)

// SetLastID stores the given ID (that can be some unrelated type with same byte length)
// in the statistics of the database.
// This function can be helpful to store the last state of some operation.
// The stored value can be retrieved using a call to LastID.
// Setting the value has no effect on the indexed data.
func (f *File) SetLastID(id ID) error {
	return f.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{bucketStats})
		var idBytes [sizeID]byte
		binary.LittleEndian.PutUint32(idBytes[:], id)
		return bucket.Put([]byte(dbStatsLastID), idBytes[:])
	})
}

// LastID returns the last ID that was saved using SetLastID.
// This function can be helpful to get the last state of an operation.
func (f *File) LastID() (ID, error) {
	var id ID
	var err = f.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{bucketStats})
		data := bucket.Get([]byte(dbStatsLastID))
		if len(data) != sizeID {
			return errors.New("minsearch: LastID not set before")
		}
		id = binary.LittleEndian.Uint32(data[:])
		return nil
	})
	return id, err
}

// AvgCount returns the average number of IDs per key in the database at last calculation.
// If it wasn't calculated before (UpdateStatistics does it), an error is returned.
func (f *File) AvgCount() (float32, error) {
	avgCount := float32(-1)
	err := f.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{bucketStats})
		data := bucket.Get([]byte(dbStatsAvgCount))
		if len(data) != 4 {
			return errors.New("minsearch: AvgCount not calculated before")
		}
		avgCount = math.Float32frombits(binary.LittleEndian.Uint32(data[:]))
		return nil
	})
	return avgCount, err
}

// KeyCount returns the number of keys in the database at last calculation.
// If it wasn't calculated before (UpdateStatistics does it), an error is returned.
func (f *File) KeyCount() (uint32, error) {
	keyCount := uint32(0)
	err := f.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{bucketStats})
		data := bucket.Get([]byte(dbStatsKeyCount))
		if len(data) != 4 {
			return errors.New("minsearch: KeyCount not calculated before")
		}
		keyCount = binary.LittleEndian.Uint32(data[:])
		return nil
	})
	return keyCount, err
}

// UpdateStatistics calculates the current number of keys and the average data length.
func (f *File) UpdateStatistics() error {
	return f.db.Update(func(tx *bolt.Tx) error {
		keyCount := 0
		bucket := tx.Bucket([]byte{bucketWords})
		dataLen := uint64(0)
		err := bucket.ForEach(func(_, v []byte) error {
			keyCount++
			dataLen += uint64(len(v) >> 3)
			return nil
		})
		if err != nil {
			return err
		}
		avgCount := float32(float64(dataLen) / float64(keyCount))
		bucket = tx.Bucket([]byte{bucketStats})
		var avgLenBytes [4]byte
		binary.LittleEndian.PutUint32(avgLenBytes[:], math.Float32bits(avgCount))
		err = bucket.Put([]byte(dbStatsAvgCount), avgLenBytes[:])
		if err != nil {
			return err
		}
		var keyCountBytes [4]byte
		binary.LittleEndian.PutUint32(keyCountBytes[:], uint32(keyCount))
		return bucket.Put([]byte(dbStatsKeyCount), keyCountBytes[:])
	})
}
