package helpers_test

import (
	"errors"
	"testing"
	"time"
)

func TestTimePointerEquals(t *testing.T) {
	var tim, expectedTim *time.Time

	// both nil
	err := TimePointerEquals(tim, expectedTim)
	if !errors.Is(err, nil) {
		t.Errorf("expected no error, but got %s", err)
	}

	// one nil, not other (x2)
	tim = new(time.Unix(123, 0))
	err = TimePointerEquals(tim, expectedTim)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}

	tim = nil
	expectedTim = new(time.Unix(123, 0))
	err = TimePointerEquals(tim, expectedTim)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}

	// non nil, equal value
	tim = new(time.Unix(456, 0))
	expectedTim = new(time.Unix(456, 0))
	err = TimePointerEquals(tim, expectedTim)
	if !errors.Is(err, nil) {
		t.Errorf("expected no error, but got %s", err)
	}

	// non nil, not equal value
	tim = new(time.Unix(777, 0))
	expectedTim = new(time.Unix(888, 0))
	err = TimePointerEquals(tim, expectedTim)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}
}

func TestStringPointerEquals(t *testing.T) {
	var s, expectedS *string

	// both nil
	err := StringPointerEquals(s, expectedS)
	if !errors.Is(err, nil) {
		t.Errorf("expected no error, but got %s", err)
	}

	// one nil, not other (x2)
	s = new("hi")
	err = StringPointerEquals(s, expectedS)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}

	s = nil
	expectedS = new("hello")
	err = StringPointerEquals(s, expectedS)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}

	// non nil, equal value
	s = new("another")
	expectedS = new("another")
	err = StringPointerEquals(s, expectedS)
	if !errors.Is(err, nil) {
		t.Errorf("expected no error, but got %s", err)
	}

	// non nil, not equal value
	s = new("not-ok")
	expectedS = new("nope")
	err = StringPointerEquals(s, expectedS)
	if !errors.Is(err, ErrTimePointerNotEqual) {
		t.Errorf("expected error %s, but got %s", ErrTimePointerNotEqual, err)
	}
}
