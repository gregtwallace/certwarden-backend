package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
)

// emailToContact generates a string slice in the format ACME
// expects (i.e. 'mailto:' is prepended to the email)
func emailToContact(email string) (contact []string) {
	if email == "" {
		return contact
	}

	return append(contact, "mailto:"+email)
}

// newAccountPayload() generates the payload for ACME to post to the
// new-account endpoint
func (account *Account) newAccountPayload() acme.NewAccountPayload {
	return acme.NewAccountPayload{
		TosAgreed: *account.AcceptedTos,
		Contact:   emailToContact(*account.Email),
	}
}
