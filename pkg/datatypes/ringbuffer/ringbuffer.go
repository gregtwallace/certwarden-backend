package ringbuffer

import (
	"errors"
	"sync"
)

// RingBuffer is a FIFO ring buffer of a specified size
// it contains ints to point to the next spot in the ring to read
// and write
type RingBuffer[V any] struct {
	buf       []V
	size      int
	readNext  int
	writeNext int
	isFull    bool
	mu        sync.Mutex
}

// New initializes an empty RingBuffer of specified size
func NewRingBuffer[V any](size int) *RingBuffer[V] {
	return &RingBuffer[V]{
		buf:       make([]V, size),
		size:      size,
		readNext:  0,
		writeNext: 0,
		isFull:    false,
	}
}

// lenUnsafe returns the number of values currently in the ring
// buffer but DOES NOT lock the mutex.  This should not be called
// unless the mutex is already at least read locked.
func (rb *RingBuffer[V]) lenUnsafe() int {
	// calculate current length
	// full or empty
	if rb.readNext == rb.writeNext {
		if rb.isFull {
			return rb.size
		}
		return 0
	}
	// write is ahead of read
	if rb.writeNext > rb.readNext {
		return rb.writeNext - rb.readNext
	}

	// read is ahead of write
	return rb.size - rb.readNext + rb.writeNext
}

// Read reads the value from the next read position and then
// updates the buffer properties accordingly.  An error is returned
// if the buffer is empty
func (rb *RingBuffer[V]) Read() (V, error) {
	// Read must write lock ring (to update readNext and isFull)
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// check if empty
	if rb.lenUnsafe() == 0 {
		return *new(V), errors.New("ringbuffer is empty")
	}

	// read next, move read pointer
	oldestValue := rb.buf[rb.readNext]
	rb.readNext++

	// if read next passed the end of the buffer, start back at 0
	if rb.readNext == rb.size {
		rb.readNext = 0
	}

	// if buffer was previously full, it isn't anymore
	if rb.isFull {
		rb.isFull = false
	}

	return oldestValue, nil
}

// write writes the Value to the nextWrite position and updates
// ring's properties accordingly. If the buffer is full, evictOldest
// is checked and if it is true the oldest value is evicted. If it
// is false an error is returned.
func (rb *RingBuffer[V]) Write(value V, evictOldest bool) error {
	// lock ring
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// full handler
	if rb.isFull && !evictOldest {
		return errors.New("ringbuffer is full and evict oldest is false")
	}

	// write next, move write pointer
	rb.buf[rb.writeNext] = value
	rb.writeNext++

	// if write next passed the end of the buffer, start back at 0
	if rb.writeNext == rb.size {
		rb.writeNext = 0
	}

	// if read and write are now equal, buffer is full
	if rb.readNext == rb.writeNext {
		rb.isFull = true
	}

	return nil
}
