package validation_test

import (
	"certwarden-backend/pkg/validation"
	"fmt"
	"testing"
)

func TestValidation_IsNewID(t *testing.T) {
	tc := []struct {
		id          int
		expectIsNew bool
	}{
		{-1, true},
		{-15, false},
		{-15000, false},
		{0, false},
		{1, false},
		{15, false},
		{12345, false},
	}

	// run tests
	for i := range tc {
		t.Run(fmt.Sprintf("%d: id: '%d'", i, tc[i].id), func(t *testing.T) {
			result := validation.IsNewID(tc[i].id)
			if tc[i].expectIsNew != result {
				t.Errorf("id '%d' expected isnewid '%t' but got '%t'", tc[i].id, tc[i].expectIsNew, result)
			}
		})
	}
}

func TestValidation_IsValidIDValue(t *testing.T) {
	tc := []struct {
		id            int
		expectIsValid bool
	}{
		{-1, false},
		{-15, false},
		{-15000, false},
		{0, true},
		{1, true},
		{15, true},
		{12345, true},
	}

	// run tests
	for i := range tc {
		t.Run(fmt.Sprintf("%d: id: '%d'", i, tc[i].id), func(t *testing.T) {
			result := validation.IsValidIDValue(tc[i].id)
			if tc[i].expectIsValid != result {
				t.Errorf("id '%d' expected isvalididvalue '%t' but got '%t'", tc[i].id, tc[i].expectIsValid, result)
			}
		})
	}
}
