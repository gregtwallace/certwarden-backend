package ringbuffer

import (
	"errors"
	"sync"
)

var errBufferEmpty = errors.New("ringbuffer is empty")

// Buffer is a FIFO ring buffer of a specified size; if additional data is
// written to the buffer when it is full, the oldest data is evicted
type Buffer struct {
	data      []string
	size      int
	readNext  int
	writeNext int
	isFull    bool
	mu        sync.Mutex
}

// New creates an empty RingBuffer of specified size
func New(size int) *Buffer {
	return &Buffer{
		data:      make([]string, size),
		size:      size,
		readNext:  0,
		writeNext: 0,
		isFull:    false,
	}
}

// lenUnsafe returns the number of values currently in the ring
// buffer but DOES NOT lock the mutex.  This should not be called
// unless the mutex is already at least read locked.
func (b *Buffer) lenUnsafe() int {
	// calculate current length
	// full or empty
	if b.readNext == b.writeNext {
		if b.isFull {
			return b.size
		}
		return 0
	}
	// write is ahead of read
	if b.writeNext > b.readNext {
		return b.writeNext - b.readNext
	}

	// read is ahead of write
	return b.size - b.readNext + b.writeNext
}

// Push adds the provided value to the end of the FIFO buffer. If the buffer
// is full, the oldest value is evicted (i.e., the front of the FIFO buffer
// is removed)
func (b *Buffer) Push(value string) {
	// lock ring
	b.mu.Lock()
	defer b.mu.Unlock()

	// write next, move write pointer
	b.data[b.writeNext] = value
	b.writeNext++

	// if write next passed the end of the buffer, start back at 0
	if b.writeNext == b.size {
		b.writeNext = 0
	}

	// increment readNext if buffer is full (evict oldest)
	if b.isFull {
		b.readNext++
		// passed end of buffer? go back to 0
		if b.readNext == b.size {
			b.readNext = 0
		}
	}

	// if read and write are now equal, buffer is full
	if b.readNext == b.writeNext {
		b.isFull = true
	}
}

// Pop removes and returns the next value in the FIFO buffer. An error is
// returned if the buffer is empty
func (b *Buffer) Pop() (string, error) {
	// Read must write lock ring (to update readNext and isFull)
	b.mu.Lock()
	defer b.mu.Unlock()

	// check if empty
	if b.lenUnsafe() == 0 {
		return "", errBufferEmpty
	}

	// read next, move read pointer
	oldestValue := b.data[b.readNext]
	b.readNext++

	// if read next passed the end of the buffer, start back at 0
	if b.readNext == b.size {
		b.readNext = 0
	}

	// if buffer was previously full, it isn't anymore
	if b.isFull {
		b.isFull = false
	}

	return oldestValue, nil
}
