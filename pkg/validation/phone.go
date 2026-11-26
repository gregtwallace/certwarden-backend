package validation

import "regexp"

// Phone number format:
// +<countrycode><subscriber number>
// Example: +351123456789
//
// Country code: 1–3 digits
// Subscriber number: 4–14 digits
//
// Example full lengths: +1XXXXXXXXXX, +351XXXXXXXXX
//
// Regex explanation:
// ^\+		 literal +
// [1-9]\d{0,2}	 country code, 1–3 digits, not starting with 0
// \d{4,14}$	 subscriber number: 4–14 digits
//
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{0,2}\d{4,14}$`)

// PhoneValid returns true if the phone number is valid.
func PhoneValid(phone string) bool {
    return phoneRegex.MatchString(phone)
}

// PhoneValidOrBlank returns true if the phone number is empty or valid.
func PhoneValidOrBlank(phone string) bool {
    return phone == "" || PhoneValid(phone)
}
