package ringbuffer

import (
	"certwarden-backend/pkg/helpers_test"
	"fmt"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	b := New(5)

	if b.size != 5 {
		t.Errorf("expected size 5 but got %d", b.size)
	}

	if b.readNext != 0 {
		t.Errorf("expected readNext 0 but got %d", b.readNext)
	}

	if b.writeNext != 0 {
		t.Errorf("expected writeNext 0 but got %d", b.writeNext)
	}

	if b.isFull {
		t.Errorf("expected isFull false but got true")
	}
}

func TestLenUnsafe(t *testing.T) {
	// manually set struct values that are needed for lenUnsafe
	testCases := []struct {
		size        int
		readNext    int
		writeNext   int
		isFull      bool
		expectedLen int
	}{
		{3, 0, 0, false, 0}, // empty (write = read = 0)
		{3, 0, 0, true, 3},  // full (write = read = 0)
		{4, 3, 3, false, 0}, // empty (write = read = size - 1)
		{4, 3, 3, true, 4},  // full (write = read = size - 1)
		{3, 1, 0, false, 2}, // read ahead of write
		{3, 0, 1, false, 1}, // write ahead of read
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (size: %d, readNext: %d, writeNext: %d, isFull: %t)", i, tc.size, tc.readNext, tc.writeNext, tc.isFull),
			func(t *testing.T) {
				b := &Buffer{
					size:      tc.size,
					readNext:  tc.readNext,
					writeNext: tc.writeNext,
					isFull:    tc.isFull,
				}

				actualLen := b.lenUnsafe()
				if actualLen != tc.expectedLen {
					t.Errorf("expected len '%d' but got '%d'", tc.expectedLen, actualLen)
				}
			})
	}
}

func TestPush(t *testing.T) {
	// manually set struct values that are needed for lenUnsafe
	testCases := []struct {
		data      []string
		size      int
		readNext  int
		writeNext int
		isFull    bool

		pushData []string

		expectedData      []string
		expectedReadNext  int
		expectedWriteNext int
		expectedIsFull    bool
	}{
		{[]string{"a", "b", "c"}, 3, 0, 0, false, []string{"new-val"}, []string{"new-val", "b", "c"}, 0, 1, false},                     // empty (write = read = 0)
		{[]string{"a", "b", "c"}, 3, 0, 0, true, []string{"new-val"}, []string{"new-val", "b", "c"}, 1, 1, true},                       // full (write = read = 0)
		{[]string{"a", "b", "c", "d"}, 4, 3, 3, false, []string{"a2", "b2"}, []string{"b2", "b", "c", "a2"}, 3, 1, false},              // empty (write = read = 0)
		{[]string{"a", "b", "c", "d"}, 4, 3, 3, true, []string{"a2", "b2"}, []string{"b2", "b", "c", "a2"}, 1, 1, true},                // full (write = read = 0)
		{[]string{"a", "b", "c", "d"}, 4, 0, 0, false, []string{"a2", "b2", "c2", "d2"}, []string{"a2", "b2", "c2", "d2"}, 0, 0, true}, // fill it (start 0)
		{[]string{"a", "b", "c", "d"}, 4, 2, 2, false, []string{"a2", "b2", "c2", "d2"}, []string{"c2", "d2", "a2", "b2"}, 2, 2, true}, // fill it (start 2)
		{[]string{"a", "b", "c"}, 3, 2, 1, false, []string{"a2"}, []string{"a", "a2", "c"}, 2, 2, true},                                // read ahead of write
		{[]string{"a", "b", "c"}, 3, 1, 2, false, []string{"a2"}, []string{"a", "b", "a2"}, 1, 0, false},                               // write ahead of read
		{[]string{"a", "b", "c"}, 3, 2, 1, false, []string{"a2", "b2"}, []string{"a", "a2", "b2"}, 0, 0, true},                         // read ahead of write x2
		{[]string{"a", "b", "c"}, 3, 1, 2, false, []string{"a2", "b2"}, []string{"b2", "b", "a2"}, 1, 1, true},                         // write ahead of read x2
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			b := &Buffer{
				data:      tc.data,
				size:      tc.size,
				readNext:  tc.readNext,
				writeNext: tc.writeNext,
				isFull:    tc.isFull,
			}

			// push new data
			for i := range tc.pushData {
				b.Push(tc.pushData[i])
			}

			// check results
			if !slices.Equal(b.data, tc.expectedData) {
				t.Errorf("expected data to be '%v' but got '%v'", tc.expectedData, b.data)
			}

			if b.readNext != tc.expectedReadNext {
				t.Errorf("expected readNext '%d' but got '%d'", tc.expectedReadNext, b.readNext)
			}

			if b.writeNext != tc.expectedWriteNext {
				t.Errorf("expected writeNext '%d' but got '%d'", tc.expectedWriteNext, b.writeNext)
			}

			if b.isFull != tc.expectedIsFull {
				t.Errorf("expected isFull '%t' but got '%t'", tc.expectedIsFull, b.isFull)
			}
		})
	}
}

func TestPop(t *testing.T) {
	// manually set struct values that are needed for lenUnsafe
	testCases := []struct {
		data      []string
		size      int
		readNext  int
		writeNext int
		isFull    bool

		expectedPopVals []string
		expectedPopErrs []error

		expectedData      []string
		expectedReadNext  int
		expectedWriteNext int
		expectedIsFull    bool
	}{
		// empty
		{[]string{"a", "b", "c"}, 3, 0, 0, false, []string{""}, []error{errBufferEmpty}, []string{"a", "b", "c"}, 0, 0, false},                     // empty (write = read = 0)
		{[]string{"a", "b", "c"}, 3, 0, 0, false, []string{"", ""}, []error{errBufferEmpty, errBufferEmpty}, []string{"a", "b", "c"}, 0, 0, false}, // empty (write = read = 0)

		// drain from full to empty
		{[]string{"a", "b", "c"}, 3, 0, 0, true, []string{"a", "b", "c", ""}, []error{nil, nil, nil, errBufferEmpty}, []string{"a", "b", "c"}, 0, 0, false},
		{[]string{"a", "b", "c"}, 3, 2, 2, true, []string{"c", "a", "b", ""}, []error{nil, nil, nil, errBufferEmpty}, []string{"a", "b", "c"}, 2, 2, false},

		// differing read/write locations
		{[]string{"a", "b", "c"}, 3, 2, 1, false, []string{"c"}, []error{nil}, []string{"a", "b", "c"}, 0, 1, false},
		{[]string{"a", "b", "c"}, 3, 2, 1, false, []string{"c", "a"}, []error{nil, nil}, []string{"a", "b", "c"}, 1, 1, false}, // x2
		{[]string{"a", "b", "c"}, 3, 1, 2, false, []string{"b"}, []error{nil}, []string{"a", "b", "c"}, 2, 2, false},
		{[]string{"a", "b", "c"}, 3, 1, 2, false, []string{"b", ""}, []error{nil, errBufferEmpty}, []string{"a", "b", "c"}, 2, 2, false}, // x2
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			b := &Buffer{
				data:      tc.data,
				size:      tc.size,
				readNext:  tc.readNext,
				writeNext: tc.writeNext,
				isFull:    tc.isFull,
			}

			// push new data
			for i := range tc.expectedPopVals {
				val, err := b.Pop()
				if val != tc.expectedPopVals[i] {
					t.Errorf("expected pop value '%s' but got '%s'", tc.expectedPopVals[i], val)
				}

				if !helpers_test.ErrorsIs(err, tc.expectedPopErrs[i]) {
					t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPopErrs[i]), helpers_test.ErrorToVal(err))
				}
			}

			// check results
			if !slices.Equal(b.data, tc.expectedData) {
				t.Errorf("expected data to be '%v' but got '%v'", tc.expectedData, b.data)
			}

			if b.readNext != tc.expectedReadNext {
				t.Errorf("expected readNext '%d' but got '%d'", tc.expectedReadNext, b.readNext)
			}

			if b.writeNext != tc.expectedWriteNext {
				t.Errorf("expected writeNext '%d' but got '%d'", tc.expectedWriteNext, b.writeNext)
			}

			if b.isFull != tc.expectedIsFull {
				t.Errorf("expected isFull '%t' but got '%t'", tc.expectedIsFull, b.isFull)
			}
		})
	}
}
