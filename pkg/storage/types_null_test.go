package storage_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestNullInt64UnixToTime(t *testing.T) {
	tc := []struct {
		nullInt      sql.NullInt64
		expectedTime *time.Time
	}{
		// nil
		{
			sql.NullInt64{
				Valid: false,
				Int64: -88884,
			},
			nil,
		},
		{
			sql.NullInt64{
				Valid: false,
				Int64: 64,
			},
			nil,
		},
		// value
		{
			sql.NullInt64{
				Valid: true,
				Int64: 123566666,
			},
			new(time.Unix(123566666, 0)),
		},
		{
			sql.NullInt64{
				Valid: true,
				Int64: -85672,
			},
			new(time.Unix(-85672, 0)),
		},
	}

	for i := range tc {
		t.Run(fmt.Sprintf("%d: valid: %t, int64: %d", i, tc[i].nullInt.Valid, tc[i].nullInt.Int64), func(t *testing.T) {
			result := storage.NullInt64UnixToTime(tc[i].nullInt)

			if result == nil && tc[i].expectedTime != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedTime == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedTime == nil {
				return
			}

			if !result.Equal(*tc[i].expectedTime) {
				t.Errorf("expected time was '%s' but got '%s'", *tc[i].expectedTime, *result)
			}
		})
	}
}

func TestTimePointerToNullInt64(t *testing.T) {
	tc := []struct {
		tim               *time.Time
		expectedNullInt64 sql.NullInt64
	}{
		// nil
		{
			nil,
			sql.NullInt64{
				Valid: false,
			},
		},
		// value
		{
			new(time.Unix(123566666, 0)),
			sql.NullInt64{
				Valid: true,
				Int64: 123566666,
			},
		},
		{
			new(time.Unix(-85672, 0)),
			sql.NullInt64{
				Valid: true,
				Int64: -85672,
			},
		},
	}

	for i := range tc {
		t.Run(fmt.Sprintf("%d: time: %s", i, helpers_test.TimeToVal(tc[i].tim)), func(t *testing.T) {
			result := storage.TimePointerToNullInt64(tc[i].tim)

			if result.Valid != tc[i].expectedNullInt64.Valid {
				t.Errorf("expected result valid was '%t' but got '%t'", tc[i].expectedNullInt64.Valid, result.Valid)
			}

			if result.Int64 != tc[i].expectedNullInt64.Int64 {
				t.Errorf("expected result int64 value was '%d' but got '%d'", tc[i].expectedNullInt64.Int64, result.Int64)
			}
		})
	}
}

func TestNullStringToString(t *testing.T) {
	tc := []struct {
		nullString     sql.NullString
		expectedString *string
	}{
		// nil
		{
			sql.NullString{
				Valid:  false,
				String: "some val",
			},
			nil,
		},
		{
			sql.NullString{
				Valid:  false,
				String: "another val",
			},
			nil,
		},
		// value
		{
			sql.NullString{
				Valid:  true,
				String: "some val",
			},
			new("some val"),
		},
		{
			sql.NullString{
				Valid:  true,
				String: "Some value goes here$ and stuff123.",
			},
			new("Some value goes here$ and stuff123."),
		},
	}

	for i := range tc {
		t.Run(fmt.Sprintf("%d: valid: %t, string: %s", i, tc[i].nullString.Valid, tc[i].nullString.String), func(t *testing.T) {
			result := storage.NullStringToString(tc[i].nullString)

			if result == nil && tc[i].expectedString != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedString == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedString == nil {
				return
			}

			if *result != *tc[i].expectedString {
				t.Errorf("expected time was '%s' but got '%s'", *tc[i].expectedString, *result)
			}
		})
	}
}
