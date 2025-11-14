package validation

import "testing"

// validPhones is a list of valid phone numbers in the format:
// +<countrycode><number>
var validPhones = []string{
	"+14155552671",	   // US
	"+351123456789",   // Portugal
	"+447911123456",   // UK
	"+4930123456",	   // Germany
	"+818012345678",   // Japan
	"+61234567890",	   // Australia (example)
	"+972501234567",   // Israel
}

// invalidPhones is a list of phone numbers that must fail validation.
var invalidPhones = []string{
	"14155552671",	    // missing +
	"+014155552671",    // country code starting with 0
	"+1",		    // too short
	"+123",		    // too short for subscriber number
	"+351-123456789",   // invalid character
	"+351 123456789",   // space not allowed
	"+351123456789abcdef", // letters not allowed
	"+---123456",	    // nonsense
	"+999999999999999999", // too long
	"+",		    // just plus
	"",		    // empty is handled separately
}

// makeValidPhones returns valid test cases
func makeValidPhones() []string {
	valid := []string{}
	for _, phone := range validPhones {
		valid = append(valid, phone)
	}
	return valid
}

// makeInvalidPhones returns invalid test cases
func makeInvalidPhones() []string {
	invalid := []string{}
	for _, phone := range invalidPhones {
		invalid = append(invalid, phone)
	}
	return invalid
}

func TestValidation_PhoneValid(t *testing.T) {
	// test valid phones
	for _, phone := range makeValidPhones() {
		if !PhoneValid(phone) {
			t.Errorf("valid phone test case '%s' returned invalid", phone)
		}
	}

	// test invalid phones
	for _, phone := range makeInvalidPhones() {
		if PhoneValid(phone) {
			t.Errorf("invalid phone test case '%s' returned valid", phone)
		}
	}
}

func TestValidation_PhoneValidOrBlank(t *testing.T) {
	// test blank input
	if !PhoneValidOrBlank("") {
		t.Error("valid phone or blank test case '' (i.e., blank) returned invalid")
	}

	// test valid phones
	for _, phone := range makeValidPhones() {
		if !PhoneValidOrBlank(phone) {
			t.Errorf("valid phone or blank test case '%s' returned invalid", phone)
		}
	}

	// test invalid phones
	for _, phone := range makeInvalidPhones() {
		if PhoneValidOrBlank(phone) {
			t.Errorf("invalid phone or blank test case '%s' returned valid", phone)
		}
	}
}
