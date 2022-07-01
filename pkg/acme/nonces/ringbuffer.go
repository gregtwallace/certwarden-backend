package nonces

import (
	"errors"
	"sync"
)

// RingBuffer is a FIFO ring buffer of a specified size
// it contains ints to point to the next spot in the ring to read
// and write
type ringBuffer struct {
	buf       []string
	size      int
	readNext  int
	writeNext int
	isFull    bool
	mu        sync.Mutex
}

// New initializes an empty RingBuffer of specified size
func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		buf:       make([]string, size),
		size:      size,
		readNext:  0,
		writeNext: 0,
		isFull:    false,
	}
}

// length returns the number of strings currently in the ring
// buffer
func (ring *ringBuffer) length() (len int) {
	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

	// get length
	return ring.lengthUnsafe()
}

// lengthUnsafe returns the number of strings currently in the ring
// buffer but DOES NOT lock the mutex.  This should not be called
// unless the mutex is already locked.
func (ring *ringBuffer) lengthUnsafe() (len int) {
	// calculate current length
	// full or empty
	if ring.readNext == ring.writeNext {
		if ring.isFull {
			return ring.size
		}
		return 0
	}
	// write is ahead of read
	if ring.writeNext > ring.readNext {
		return ring.writeNext - ring.readNext
	}

	// read is ahead of write
	return ring.size - ring.readNext + ring.writeNext
}

// read reads the string from the next read position and then
// updates the buffer properties accordingly.  An error is returned
// if the buffer is empty
func (ring *ringBuffer) read() (oldest string, err error) {
	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

	return ring.readUnsafe()
}

// readUnsafe reads the string from the next read position and then
// updates the buffer properties accordingly.  An error is returned
// if the buffer is empty. It DOES NOT lock the mutex!
func (ring *ringBuffer) readUnsafe() (oldest string, err error) {
	// check if empty
	if ring.lengthUnsafe() == 0 {
		return "", errors.New("ringbuffer is empty")
	}

	// read next, move read pointer
	oldest = ring.buf[ring.readNext]
	ring.readNext++

	// if read next passed the end of the buffer, start back at 0
	if ring.readNext == ring.size {
		ring.readNext = 0
	}

	// if buffer was previously full, it isn't anymore
	if ring.isFull {
		ring.isFull = false
	}

	return oldest, nil
}

// write writes new string to the nextWrite position and updates
// ring's properties accordingly.  An error is returns if the buffer
// is full
func (ring *ringBuffer) write(new string) (err error) {
	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

	return ring.writeUnsafe(new)
}

// writeUnsafe writes new string to the nextWrite position and updates
// ring's properties accordingly.  An error is returns if the buffer
// is full. It DOES NOT lock the mutex!
func (ring *ringBuffer) writeUnsafe(new string) (err error) {
	if ring.isFull == true {
		return errors.New("ringbuffer is full")
	}

	// write next, move write pointer
	ring.buf[ring.writeNext] = new
	ring.writeNext++

	// if write next passed the end of the buffer, start back at 0
	if ring.writeNext == ring.size {
		ring.writeNext = 0
	}

	// if read and write are now equal, buffer is full
	if ring.readNext == ring.writeNext {
		ring.isFull = true
	}

	return nil
}
