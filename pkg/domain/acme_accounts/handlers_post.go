package acme_accounts

import (
	"database/sql"
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NewPayload is the struct for creating a new account
type NewPayload struct {
	ID           *int    `json:"id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	PrivateKeyID *int    `json:"private_key_id"`
	IsStaging    *bool   `json:"is_staging"`
	AcceptedTos  *bool   `json:"accepted_tos"`
}

func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) (err error) {
	var payload NewPayload

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// do validation
	// id
	err = utils.IsIdNew(payload.ID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// name
	err = service.isNameValid(payload.ID, payload.Name)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// TOS must be accepted
	if !*payload.AcceptedTos {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// private key
	err = service.keys.IsPrivateKeyValid(payload.PrivateKeyID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	///

	// Save new account details to storage.
	// No ACME actions are performed.
	account, err := service.storage.PostNewAccount(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	_, err = output.WriteJSON(w, http.StatusCreated, account, "acme_account")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// NewAccount sends the account information to the ACME new-account endpoint
// which effectively registers the account with ACME
// endpoint: /api/v1/acmeaccounts/:id/new-account
func (service *Service) NewAcmeAccount(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		if err == sql.ErrNoRows {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// no need to validate, can try to register any account
	// that has a cryptokey

	// get crypto key
	key, err := account.PrivateKey.CryptoKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if account.IsStaging {
		acmeResponse, err = service.acmeStaging.NewAccount(account.newAccountPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewAccount(account.newAccountPayload(), key)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// fetch the updated account
	account, err = service.storage.GetOneAccountById(idParam)
	if err != nil {
		// no need for no row
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return modified account to client
	_, err = output.WriteJSON(w, http.StatusOK, account, "acme_account")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// Deactivate sets deactivated status for the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
// endpoint: /api/v1/acmeaccounts/:id/deactivate
func (service *Service) Deactivate(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		if err == sql.ErrNoRows {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// no need to validate, can try to deactivate any account
	// that has an accountKey

	// get AccountKey
	accountKey, err := account.accountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if account.IsStaging {
		acmeResponse, err = service.acmeStaging.DeactivateAccount(accountKey)
	} else {
		acmeResponse, err = service.acmeProd.DeactivateAccount(accountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// fetch the updated account
	account, err = service.storage.GetOneAccountById(idParam)
	if err != nil {
		// no need for no row
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return modified account to client
	_, err = output.WriteJSON(w, http.StatusOK, account, "acme_account")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
