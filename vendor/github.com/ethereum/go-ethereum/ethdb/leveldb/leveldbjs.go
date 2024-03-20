//go:build js
// +build js

// Package leveldb implements the key-value database layer based on LevelDB.
package leveldb

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	// degradationWarnInterval specifies how often warning should be printed if the
	// leveldb database cannot keep up with requested writes.
	degradationWarnInterval = time.Minute

	// minCache is the minimum amount of memory in megabytes to allocate to leveldb
	// read and write caching, split half and half.
	minCache = 16

	// minHandles is the minimum number of files handles to allocate to the open
	// database files.
	minHandles = 16

	// metricsGatheringInterval specifies the interval to retrieve leveldb database
	// compaction, io and pause stats to report to the user.
	metricsGatheringInterval = 3 * time.Second
)

// Database is a persistent key-value store. Apart from basic data storage
// functionality it also supports batch writes and iterating over the keyspace in
// binary-alphabetical order.
type Database struct {
	fn string      // filename for reporting
	db *leveldb.DB // LevelDB instance

	compTimeMeter      metrics.Meter // Meter for measuring the total time spent in database compaction
	compReadMeter      metrics.Meter // Meter for measuring the data read during compaction
	compWriteMeter     metrics.Meter // Meter for measuring the data written during compaction
	writeDelayNMeter   metrics.Meter // Meter for measuring the write delay number due to database compaction
	writeDelayMeter    metrics.Meter // Meter for measuring the write delay duration due to database compaction
	diskSizeGauge      metrics.Gauge // Gauge for tracking the size of all the levels in the database
	diskReadMeter      metrics.Meter // Meter for measuring the effective amount of data read
	diskWriteMeter     metrics.Meter // Meter for measuring the effective amount of data written
	memCompGauge       metrics.Gauge // Gauge for tracking the number of memory compaction
	level0CompGauge    metrics.Gauge // Gauge for tracking the number of table compaction in level0
	nonlevel0CompGauge metrics.Gauge // Gauge for tracking the number of table compaction in non0 level
	seekCompGauge      metrics.Gauge // Gauge for tracking the number of table compaction caused by read opt

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	log log.Logger // Contextual logger tracking the database path
}

// New returns a wrapped LevelDB object. The namespace is the prefix that the
// metrics reporting should use for surfacing internal stats.
func New(file string, cache int, handles int, namespace string, readonly bool) (*Database, error) {
	return NewCustom(file, namespace, func(options *opt.Options) {
		// Ensure we have some minimal caching and file guarantees
		if cache < minCache {
			cache = minCache
		}
		if handles < minHandles {
			handles = minHandles
		}
		// Set default options
		options.OpenFilesCacheCapacity = handles
		options.BlockCacheCapacity = cache / 2 * opt.MiB
		options.WriteBuffer = cache / 4 * opt.MiB // Two of these are used internally
		if readonly {
			options.ReadOnly = true
		}
	})
}

// NewCustom returns a wrapped LevelDB object. The namespace is the prefix that the
// metrics reporting should use for surfacing internal stats.
// The customize function allows the caller to modify the leveldb options.
func NewCustom(file string, namespace string, customize func(options *opt.Options)) (*Database, error) {
	return nil, fmt.Errorf("Not implemented")
}


// Close stops the metrics collection, flushes any pending data to disk and closes
// all io accesses to the underlying key-value store.
func (db *Database) Close() error {
	return fmt.Errorf("Not implemented")
}

// Has retrieves if a key is present in the key-value store.
func (db *Database) Has(key []byte) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

// Get retrieves the given key if it's present in the key-value store.
func (db *Database) Get(key []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Put inserts the given value into the key-value store.
func (db *Database) Put(key []byte, value []byte) error {
	return fmt.Errorf("Not implemented")
}

// Delete removes the key from the key-value store.
func (db *Database) Delete(key []byte) error {
	return fmt.Errorf("Not implemented")
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Database) NewBatch() ethdb.Batch {
	return &batch{
		db: db.db,
		b:  new(leveldb.Batch),
	}
}

// NewBatchWithSize creates a write-only database batch with pre-allocated buffer.
func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &batch{
		db: db.db,
		b:  leveldb.MakeBatch(size),
	}
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (db *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	return db.db.NewIterator(bytesPrefixRange(prefix, start), nil)
}

// NewSnapshot creates a database snapshot based on the current state.
// The created snapshot will not be affected by all following mutations
// happened on the database.
// Note don't forget to release the snapshot once it's used up, otherwise
// the stale data will never be cleaned up by the underlying compactor.
func (db *Database) NewSnapshot() (ethdb.Snapshot, error) {
	snap, err := db.db.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return &snapshot{db: snap}, nil
}

// Stat returns a particular internal stat of the database.
func (db *Database) Stat(property string) (string, error) {
	return db.db.GetProperty(property)
}

// Compact flattens the underlying data store for the given key range. In essence,
// deleted and overwritten versions are discarded, and the data is rearranged to
// reduce the cost of operations needed to access them.
//
// A nil start is treated as a key before all keys in the data store; a nil limit
// is treated as a key after all keys in the data store. If both is nil then it
// will compact entire data store.
func (db *Database) Compact(start []byte, limit []byte) error {
	return db.db.CompactRange(util.Range{Start: start, Limit: limit})
}

// Path returns the path to the database directory.
func (db *Database) Path() string {
	return db.fn
}

// batch is a write-only leveldb batch that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
type batch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

// Put inserts the given value into the batch for later committing.
func (b *batch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(key) + len(value)
	return nil
}

// Delete inserts the a key removal into the batch for later committing.
func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *batch) Write() error {
	return b.db.Write(b.b, nil)
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

// Replay replays the batch contents.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	return b.b.Replay(&replayer{writer: w})
}

// replayer is a small wrapper to implement the correct replay methods.
type replayer struct {
	writer  ethdb.KeyValueWriter
	failure error
}

// Put inserts the given value into the key-value data store.
func (r *replayer) Put(key, value []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Put(key, value)
}

// Delete removes the key from the key-value data store.
func (r *replayer) Delete(key []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Delete(key)
}

// bytesPrefixRange returns key range that satisfy
// - the given prefix, and
// - the given seek position
func bytesPrefixRange(prefix, start []byte) *util.Range {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return r
}

// snapshot wraps a leveldb snapshot for implementing the Snapshot interface.
type snapshot struct {
	db *leveldb.Snapshot
}

// Has retrieves if a key is present in the snapshot backing by a key-value
// data store.
func (snap *snapshot) Has(key []byte) (bool, error) {
	return snap.db.Has(key, nil)
}

// Get retrieves the given key if it's present in the snapshot backing by
// key-value data store.
func (snap *snapshot) Get(key []byte) ([]byte, error) {
	return snap.db.Get(key, nil)
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (snap *snapshot) Release() {
	snap.db.Release()
}
