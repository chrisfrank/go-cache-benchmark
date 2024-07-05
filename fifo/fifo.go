package fifo

import "sync"

// Cache is a memory-backed cache
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Add(key K, val V) bool
}

// New returns a fixed-size cache with a first-in-first-out eviction policy.
func New[K comparable, V any](maxSize int) (Cache[K, V], error) {
	if maxSize <= 0 {
		maxSize = 10
	}
	return &fifoCache[K, V]{
		maxSize: maxSize,
		entries: make([]*entry[K, V], maxSize),
		index:   make(map[K]int, maxSize),
	}, nil
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

type fifoCache[K comparable, V any] struct {
	// blank is a static empty value that we return on miss
	blank V
	// cursor tracks the position we'll write to next in the `entries` slice
	cursor int
	// index maps keys to their position in the `entries` slice
	index map[K]int
	// maxSize is the upper bound on how many entries can be cached.
	maxSize int
	// mtx makes cache access safe from concurrent goroutines
	mtx sync.Mutex
	// entries contains the cached values
	entries []*entry[K, V]
}

// Get fetches a value from cache
func (c *fifoCache[K, V]) Get(key K) (V, bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	idx, ok := c.index[key]
	if !ok {
		return c.blank, false
	}

	return c.entries[idx].value, true
}

// Add sets a new value in the cache. The returned boolean is true if an older
// value was evicted, and false otherwise.
func (c *fifoCache[K, V]) Add(key K, val V) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	prevEntry := c.entries[c.cursor]
	evictionOcurred := (prevEntry != nil)
	if evictionOcurred {
		// we're about to overwrite an existing entry in the entries() slice,
		// so we also need to clear the old entry's key from the index.
		delete(c.index, prevEntry.key)
	}

	c.index[key] = c.cursor
	c.entries[c.cursor] = &entry[K, V]{key: key, value: val}

	// advance the cursor, or rewind it back to zero if we exceed maxSize
	c.cursor += 1
	if c.cursor == c.maxSize {
		c.cursor = 0
	}

	return evictionOcurred
}
