package acme_accounts

// emailToContact generates a string slice in the format ACME
// expects (i.e. 'mailto:' is prepended to the email)
func emailToContact(email string) (contact []string) {
	if email == "" {
		return contact
	}

	return append(contact, "mailto:"+email)
}
