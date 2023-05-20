package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// NameDescPayload is the struct for editing an existing account name and desc
type NameDescPayload struct {
	ID          int     `json:"-"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	UpdatedAt   int     `json:"-"`
}

// PutNameDescAccount is a handler that sets the name and description of an account
// within storage
func (service *Service) PutNameDescAccount(w http.ResponseWriter, r *http.Request) (err error) {
	// payload decoding
	var payload NameDescPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	_, err = service.getAccount(payload.ID)
	if err != nil {
		return err
	}

	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	err = service.storage.PutNameDescAccount(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      payload.ID,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// changeEmailPayload is the struct for updating an account's email address
// do not export and do not add id/updatedAt fields since this does not get
// sent to storage
type changeEmailPayload struct {
	Email *string `json:"email"`
}

// ChangeEmail() is a handler that updates an ACME account with the specified
// email address and saves the updated address to storage
func (service *Service) ChangeEmail(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// decode payload
	var payload changeEmailPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	account, err := service.getAccount(id)
	if err != nil {
		return err
	}

	// email (update cannot be to blank)
	if payload.Email == nil || !validation.EmailValid(*payload.Email) {
		service.logger.Debug(ErrEmailBad)
		return output.ErrValidationFailed
	}
	// end validation

	// get AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// make ACME update email payload
	acmePayload := acme.UpdateAccountPayload{
		Contact: emailToContact(*payload.Email),
	}

	// send the email update to ACME
	var acmeAccount AcmeAccount
	if account.IsStaging {
		acmeAccount.Account, err = service.acmeStaging.UpdateAccount(acmePayload, acmeAccountKey)
	} else {
		acmeAccount.Account, err = service.acmeProd.UpdateAccount(acmePayload, acmeAccountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// add additional details to the payload before saving
	acmeAccount.ID = id
	acmeAccount.UpdatedAt = int(time.Now().Unix())

	// save ACME response to account
	err = service.storage.PutAcmeAccountResponse(acmeAccount)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      id,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// RolloverKeyPayload is used to change an account's private key
type RolloverKeyPayload struct {
	ID           int  `json:"-"`
	PrivateKeyID *int `json:"private_key_id"`
	UpdatedAt    int  `json:"-"`
}

// RolloverKey changes the private key used for an account
func (service *Service) RolloverKey(w http.ResponseWriter, r *http.Request) (err error) {
	// decode payload
	var payload RolloverKeyPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	account, err := service.getAccount(payload.ID)
	if err != nil {
		return err
	}

	// new private key
	if payload.PrivateKeyID == nil || !service.keys.KeyAvailable(*payload.PrivateKeyID) {
		service.logger.Debug("invalid private key specified for account key rollover")
		return output.ErrValidationFailed
	}
	// end validation

	// get AccountKey
	oldAcmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// fetch new private key
	newKey, err := service.storage.GetOneKeyById(*payload.PrivateKeyID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// get crypto key from the new key
	newCryptoKey, err := newKey.CryptoPrivateKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the rollover to ACME
	if account.IsStaging {
		err = service.acmeStaging.RolloverAccountKey(newCryptoKey, oldAcmeAccountKey)
	} else {
		err = service.acmeProd.RolloverAccountKey(newCryptoKey, oldAcmeAccountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// update private key id in db
	err = service.storage.PutNewAccountKey(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "key rolled over",
		ID:      payload.ID,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
