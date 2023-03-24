package validation

import "testing"

var validUsernames = []string{
	"greg",
	"BoB",
	"B-o.B",
	"B_o_B_SmitH",
}

var invalidUsernames = []string{
	"@greg",
	"greg@",
	"-bob",
	"bob-",
	"bob_",
	"bob__smith",
	"bob_$Smith",
	"bob smith",
	"asyouallcanseethisemailaddressexceedsthemaximumnumberofcharacterX",
}

// makeValidEmails makes the array of emails to test that should yield
// a valid result
func makeValidEmails() []string {
	validEmails := []string{}

	// valid usernames with known valid domain
	for _, username := range validUsernames {
		validEmails = append(validEmails, username+"@example.com")
	}

	// valid domains with known valid user
	for _, domain := range validDomains {
		validEmails = append(validEmails, "greg@"+domain)
	}

	return validEmails
}

// makeInvalidEmails makes the array of emails to test that should yield
// an invalid result
func makeInvalidEmails() []string {
	invalidEmails := []string{}

	// example without an @
	invalidEmails = append(invalidEmails, "john.smith.example.com")

	// invalid usernames with known valid domain
	for _, username := range invalidUsernames {
		invalidEmails = append(invalidEmails, username+"@example.com")
	}

	// invalid domains with known valid user
	for _, domain := range invalidDomains {
		invalidEmails = append(invalidEmails, "greg@"+domain)
	}

	return invalidEmails
}

func TestValidation_EmailValid(t *testing.T) {
	// test valid emails
	for _, email := range makeValidEmails() {
		if !EmailValid(email) {
			t.Errorf("valid email test case '%s' returned invalid", email)
		}
	}

	// test invalid emails
	for _, email := range makeInvalidEmails() {
		if EmailValid(email) {
			t.Errorf("invalid email test case '%s' returned valid", email)
		}
	}
}

func TestValidation_EmailValidOrBlank(t *testing.T) {
	// test blank
	if !EmailValidOrBlank("") {
		t.Error("valid email or blank test case '' (i.e. blank) returned invalid")
	}

	// test valid emails
	for _, email := range makeValidEmails() {
		if !EmailValidOrBlank(email) {
			t.Errorf("valid email or blank test case '%s' returned invalid", email)
		}
	}

	// test invalid emails
	for _, email := range makeInvalidEmails() {
		if EmailValidOrBlank(email) {
			t.Errorf("invalid email or blank test case '%s' returned valid", email)
		}
	}
}
