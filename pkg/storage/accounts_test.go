package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"testing"
)

// compareAcmeAccount compares acct to expectedAcct and throws appropriate errors for any differences
func compareAcmeAccount(t *testing.T, acct, expectedAcct *acme_accounts.Account) {
	if acct.ID != expectedAcct.ID {
		t.Errorf("acme account: id expected '%d' but got '%d'", expectedAcct.ID, acct.ID)
	}

	if acct.Name != expectedAcct.Name {
		t.Errorf("acme account: name expected '%s' but got '%s'", expectedAcct.Name, acct.Name)
	}

	if acct.Description != expectedAcct.Description {
		t.Errorf("acme account: description expected '%s' but got '%s'", expectedAcct.Description, acct.Description)
	}

	compareAcmeServer(t, &acct.AcmeServer, &expectedAcct.AcmeServer)

	compareKey(t, &acct.AccountKey, &expectedAcct.AccountKey)

	if acct.Status != expectedAcct.Status {
		t.Errorf("acme account: status expected '%s' but got '%s'", expectedAcct.Status, acct.Status)
	}

	if acct.Email != expectedAcct.Email {
		t.Errorf("acme account: email expected '%s' but got '%s'", expectedAcct.Email, acct.Email)
	}

	if acct.AcceptedTos != expectedAcct.AcceptedTos {
		t.Errorf("acme account: accepted tos expected '%t' but got '%t'", expectedAcct.AcceptedTos, acct.AcceptedTos)
	}

	if !acct.CreatedAt.Equal(expectedAcct.CreatedAt) {
		t.Errorf("acme account: created at expected '%s' but got '%s'", expectedAcct.CreatedAt.UTC(), acct.CreatedAt.UTC())
	}

	if !acct.UpdatedAt.Equal(expectedAcct.UpdatedAt) {
		t.Errorf("acme account: updated at expected '%s' but got '%s'", expectedAcct.UpdatedAt.UTC(), acct.UpdatedAt.UTC())
	}

	if acct.Kid != expectedAcct.Kid {
		t.Errorf("acme account: kid expected '%s' but got '%s'", expectedAcct.Kid, acct.Kid)
	}
}
