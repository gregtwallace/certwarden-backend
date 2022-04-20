package application

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type AcmeAccount struct {
	ID           int       `json:"id"`
	PrivateKeyID int       `json:"private_key_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Description  string    `json:"description"`
	AcceptedTos  bool      `json:"accepted_tos"`
	IsStaging    bool      `json:"is_staging"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (app *Application) GetAllAcmeAccounts(w http.ResponseWriter, r *http.Request) {
	acmeAccounts := []AcmeAccount{
		AcmeAccount{
			ID:           0,
			PrivateKeyID: 0,
			Name:         "Primary for Pub Domain",
			Email:        "greg@gregtwallace.com",
			Description:  "Main account",
			IsStaging:    false,
		},
		AcmeAccount{
			ID:           1,
			PrivateKeyID: 10,
			Name:         "Another Acct",
			Email:        "something@test.com",
			Description:  "Staging 1",
			IsStaging:    true,
		},
		AcmeAccount{
			ID:           2,
			PrivateKeyID: 4,
			Name:         "Account #3",
			Email:        "another@fake.com",
			Description:  "Staging Backup",
			IsStaging:    true,
		},
		AcmeAccount{
			ID:           3,
			PrivateKeyID: 7,
			Name:         "Primary gtw86.com",
			Email:        "greg@gtw86.com",
			Description:  "For LAN",
			IsStaging:    false,
		},
	}

	app.WriteJSON(w, http.StatusOK, acmeAccounts, "acme_accounts")

}

func (app *Application) GetOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.Logger.Print(errors.New("invalid id parameter"))
		//app.errorJSON(w, err)
		return
	}

	acmeAccount := AcmeAccount{
		ID:           id,
		PrivateKeyID: 10,
		Name:         "Another Acct",
		Email:        "something@test.com",
		Description:  "Staging 1",
		IsStaging:    true,
	}

	app.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
}
