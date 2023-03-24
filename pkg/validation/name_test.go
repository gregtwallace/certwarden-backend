package validation

import "testing"

var validNames = []string{
	"test",
	"aName",
	"sOmENaMEee",
	"name.com",
	"name.com.",
	"some.name.com...",
	"som~name.here",
	"myTest_-Name",
}

var invalidNames = []string{
	"",
	" ",
	"    ",
	"a Name",
	" aName",
	"aName ",
	"some$name",
}

func TestValidation_NameValid(t *testing.T) {
	// test valid names
	for _, name := range validNames {
		valid := NameValid(name)
		if !valid {
			t.Errorf("valid name test case '%s' returned invalid", name)
		}
	}

	// test invalid names
	for _, name := range invalidNames {
		valid := NameValid(name)
		if valid {
			t.Errorf("invalid name test case '%s' returned valid", name)
		}
	}
}
