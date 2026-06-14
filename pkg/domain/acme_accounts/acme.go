package acme_accounts

import (
	"certwarden-backend/pkg/acme"
	"strings"
	"time"
)

func acmeAcctToUpdatePayload(acctId int, acmeAcct acme.Account) UpdatePayload {
	// convert to email if the first contact is a valid `mailto:`,
	// otherwise leave it blank but not null (i.e., update it to blank)
	email := ""
	if len(acmeAcct.Contact) > 0 && strings.HasPrefix(acmeAcct.Contact[0], "mailto:") {
		email = strings.TrimPrefix(acmeAcct.Contact[0], "mailto:")
	}

	return UpdatePayload{
		ID:        acctId,
		Status:    &acmeAcct.Status,
		Email:     &email,
		CreatedAt: acmeAcct.CreatedAt,
		UpdatedAt: time.Now(),
		KID:       acmeAcct.Location,
	}
}

// emailToContact generates a string slice in the format ACME
// expects (i.e. 'mailto:' is prepended to the email)
func emailToContact(email string) (contact []string) {
	if email == "" {
		return contact
	}

	return append(contact, "mailto:"+email)
}
