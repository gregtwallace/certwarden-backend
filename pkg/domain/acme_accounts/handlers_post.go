package acme_accounts

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"legocerthub-backend/pkg/acme"
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

func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) {
	var payload NewPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Printf("accounts: PostNew: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// do validation
	// id
	err = utils.IsIdNew(payload.ID)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = service.isNameValid(payload.ID, payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// TOS must be accepted
	if !*payload.AcceptedTos {
		service.logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	// private key
	err = service.keys.IsPrivateKeyValid(payload.PrivateKeyID)
	if err != nil {
		service.logger.Printf("accounts: PostNew: private key error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// Save new account details to storage.
	// No ACME actions are performed.
	err = service.storage.PostNewAccount(payload)
	if err != nil {
		service.logger.Printf("accounts: post new: failed to create account: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("accounts: PostNew: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// NewAccount sends the account information to the ACME new-account endpoint
// which effectively registers the account with ACME
// endpoint: /api/v1/acmeaccounts/:id/new-account
func (service *Service) NewAccount(w http.ResponseWriter, r *http.Request) {
	// account id
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid idParam -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// read body to verify empty (there should not be a payload)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		service.logger.Printf("accounts: new-account: failed to read request body -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} else if len(b) != 0 {
		err = errors.New("accounts: new-account: request body should be undefined")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// fetch the relevant account
	acmeAccount, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Printf("accounts: new-account: failed to fetch account -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// get crypto key
	key, err := acmeAccount.PrivateKey.CryptoKey()
	if err != nil {
		service.logger.Printf("accounts: new-account: failed to get crypto private key -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if acmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.NewAccount(acmeAccount.newAccountPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewAccount(acmeAccount.newAccountPayload(), key)
	}
	if err != nil {
		service.logger.Printf("accounts: acme: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Printf("accounts: new-account: failed to save to storage -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	clientResponse := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, clientResponse, "response")
	if err != nil {
		service.logger.Printf("accounts: new-account: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// Deactivate sets deactivated status for the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
// endpoint: /api/v1/acmeaccounts/:id/deactivate
func (service *Service) Deactivate(w http.ResponseWriter, r *http.Request) {
	// account id
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("accounts: deactivate: invalid idParam -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// read body to verify empty (there should not be a payload)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		service.logger.Printf("accounts: deactivate: failed to read request body -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} else if len(b) != 0 {
		err = errors.New("accounts: deactivate: request body should be undefined")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// fetch the relevant account
	acmeAccount, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Printf("accounts: deactivate: failed to fetch account -- err: %s", err)
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

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if acmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.DeactivateAccount(accountKey)
	} else {
		acmeResponse, err = service.acmeProd.DeactivateAccount(accountKey)
	}
	if err != nil {
		service.logger.Printf("accounts: acme: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Printf("accounts: new-account: failed to save to storage -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	clientResponse := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, clientResponse, "response")
	if err != nil {
		service.logger.Printf("accounts: new-account: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

}
