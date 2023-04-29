package validation

import "testing"

// valid domains
var validDomains = []string{
	"test.greg.co",
	"domain.com",
	"aNoThER.oRG",
	"my.some.another.com.co",
	"x6hello.xom",
	"some-name.com",
	"www.example.com",
	"www.EXAMPLE.COM",
	"WWW.exampLE.CoM",
}

// invalid domains
var invalidDomains = []string{
	"",
	"fake.example.com.",
	"fake.example.com.x",
	"hello",
	".com",
	".example.com",
	".example.org",
	".aNoThER.oRG",
	"hel lo.com",
	" hello.com",
	"hello.com ",
	"my.some.another.com.c",
	"x_hello.xom",
	"-some-name.com",
	"some^name.com",
	"local",
	"президент.рф",
	"invalid..",
	"$invalid.com",
	"invalid$.com",
	"invalid.$com",
	"asyouallcanseethisemailaddressexceedsthemaximumnumberofcharactersallowedtobeintheemailaddresswhichisnomorethatn254accordingtovariousrfcokaycanistopnowornotyetnoineedmorecharacterstoaddi.really.cannot.thinkof.what.else.to.put.into.this.invalid.address.net",
}

func TestValidation_DomainValid(t *testing.T) {
	// test valid domains
	for _, domain := range validDomains {
		// wildcard on (no wildcard tests though)
		valid := DomainValid(domain, true)
		if !valid {
			t.Errorf("valid domain name test case '%s' returned invalid", domain)
		}

		// wildcard off (no wildcard tests though)
		valid = DomainValid(domain, false)
		if !valid {
			t.Errorf("valid domain name test case '%s' returned invalid", domain)
		}

		// test with wild card + valid domain + wildcard ok on
		domain = "*." + domain

		valid = DomainValid(domain, true)
		if !valid {
			t.Errorf("valid domain name wildcard test case '%s' returned invalid", domain)
		}

		// test with wild card + valid domain + wildcard ok NOT on
		valid = DomainValid(domain, false)
		// should NOT be valid since support is off
		if valid {
			t.Errorf("valid domain name wildcard test case with wildcard off '%s' returned valid", domain)
		}
	}

	// test invalid domains
	for _, domain := range invalidDomains {
		// wildcard on (no wildcard tests though)
		valid := DomainValid(domain, true)
		if valid {
			t.Errorf("invalid domain name test case '%s' returned valid", domain)
		}

		// wildcard off (no wildcard tests though)
		valid = DomainValid(domain, false)
		if valid {
			t.Errorf("invalid domain name test case '%s' returned valid", domain)
		}

		// test with wild card + invalid domain + wildcard ok on
		domain = "*." + domain

		valid = DomainValid(domain, true)
		if valid {
			t.Errorf("invalid domain name wildcard test case '%s' returned valid", domain)
		}

		// test with wild card + invalid domain + wildcard ok NOT on
		valid = DomainValid(domain, false)
		if valid {
			t.Errorf("invalid domain name wildcard test case '%s' returned valid", domain)
		}
	}

}
