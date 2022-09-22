package acme_accounts

import "legocerthub-backend/pkg/acme"

// AcmeAccount is the ACME Account object plus some additional
// details for storage.
type AcmeAccount struct {
	acme.Account
	ID        int `json:"-"`
	UpdatedAt int `json:"-"`
}

// emailToContact generates a string slice in the format ACME
// expects (i.e. 'mailto:' is prepended to the email)
func emailToContact(email string) (contact []string) {
	if email == "" {
		return contact
	}

	return append(contact, "mailto:"+email)
}
