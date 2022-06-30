package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NameDescPayload is the struct for editing an existing key
type NameDescPayload struct {
	ID          *int    `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// PutNameDescAccount is a handler that sets the name and description of an account
// within storage
func (service *Service) PutNameDescAccount(w http.ResponseWriter, r *http.Request) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid idParam -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	var payload NameDescPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to decode payload: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = service.isIdExisting(idParam, payload.ID)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = service.isNameValid(payload.ID, payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// update db
	err = service.storage.PutNameDescAccount(payload)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("accounts: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// NameDescPayload is the struct for editing an existing key
type ChangeEmailPayload struct {
	ID    *int    `json:"id"`
	Email *string `json:"email"`
}

// ChangeEmail() is a handler that updates an ACME account with the specified
// email address and saves the updated address to storage
func (service *Service) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	// get id param and payload
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("accounts: update email: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	var payload ChangeEmailPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Printf("accounts: update email: failed to decode payload: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = service.isIdExisting(idParam, payload.ID)
	if err != nil {
		service.logger.Printf("accounts: update email: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// email (update cannot be to blank)
	err = utils.IsEmailValid(payload.Email)
	if err != nil {
		service.logger.Printf("accounts: update email: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// fetch the relevant account
	acmeAccount, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Printf("accounts: update email: failed to fetch account -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// get AccountKey
	accountKey, err := acmeAccount.AccountKey()
	if err != nil {
		service.logger.Printf("accounts: update email: failed to get crypto private key -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// ACME update email payload
	acmePayload := acme.UpdateAccountPayload{
		Contact: emailToContact(*payload.Email),
	}
	log.Println(acmePayload.Contact)

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if acmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.UpdateAccount(acmePayload, accountKey)
	} else {
		acmeResponse, err = service.acmeProd.UpdateAccount(acmePayload, accountKey)
	}
	if err != nil {
		service.logger.Printf("accounts: acme: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Printf("accounts: update email: failed to save to storage -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	clientResponse := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, clientResponse, "response")
	if err != nil {
		service.logger.Printf("accounts: update email: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
