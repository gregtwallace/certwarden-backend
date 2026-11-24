package validation_test

import (
	"certwarden-backend/pkg/validation"
	"fmt"
	"testing"
)

func TestValidation_IsValidName(t *testing.T) {
	tc := []struct {
		name          string
		expectedValid bool
	}{
		{"test", true},
		{"aName", true},
		{"sOmENaMEee", true},
		{"name.com", true},
		{"name.com.", true},
		{"some.name.com...", true},
		{"som~name.here", true},
		{"myTest_-Name", true},
		{"ok~name", true},

		{"", false},
		{" ", false},
		{"    ", false},
		{"a Name", false},
		{" aName", false},
		{"aName ", false},
		{"some$name", false},
		{"a[name]", false},
		{"another`name", false},
		{"aga^in", false},
		{"ag\\ain", false},
	}

	// run tests
	for i := range tc {
		t.Run(fmt.Sprintf("%d: name: %q", i, tc[i].name), func(t *testing.T) {
			result := validation.IsValidName(tc[i].name)
			if tc[i].expectedValid != result {
				t.Errorf("id %q expected isvalidname '%t' but got '%t'", tc[i].name, tc[i].expectedValid, result)
			}
		})
	}
}
