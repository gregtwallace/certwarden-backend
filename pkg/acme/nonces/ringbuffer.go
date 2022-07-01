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
func newRingBuffer(len int) *ringBuffer {
	return &ringBuffer{
		buf:       make([]string, len),
		size:      len,
		readNext:  0,
		writeNext: 0,
		isFull:    false,
	}
}

// Length returns the number of strings currently in the ring
// buffer
func (ring *ringBuffer) length() (len int) {
	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

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

// Read reads the string from the next read position and then
// updates the buffer properties accordingly.  An error is returned
// if the buffer is empty
func (ring *ringBuffer) read() (oldest string, err error) {
	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

	// check if empty
	if ring.readNext == ring.writeNext && !ring.isFull {
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

// Write writes new string to the nextWrite position and updates
// ring's properties accordingly.  An error is returns if the buffer
// is full
func (ring *ringBuffer) write(new string) (err error) {
	if ring.isFull == true {
		return errors.New("ringbuffer is full")
	}

	// lock ring
	ring.mu.Lock()
	defer ring.mu.Unlock()

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
