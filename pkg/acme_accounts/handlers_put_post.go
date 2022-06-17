package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// PutOneAccount is an http handler that overwrites the specified (by id) acme account with the
// data PUT by the client

// TODO: fetch data from LE and update kid
// if kid already exists, use it to update the email field

func (service *Service) PutOneAccount(w http.ResponseWriter, r *http.Request) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid idParam -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	payload, err := decodePayload(r)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to decode payload: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = service.isAccountIdExisting(idParam, payload.ID)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(*payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// email
	err = utils.IsEmailValidOrBlank(*payload.Email)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// check db for account kid
	kid, err := service.storage.GetAccountKid(*payload.ID)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to check for kid -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// if no kid, do new LE registration
	if kid == "" {
		// always update name and desc
		err = service.storage.PutExistingAccountNameDesc(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to update account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// do new LE reg
		err = service.createNewLEAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to create new LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	} else {
		// if there is a kid, update the existing account with new data
		// do LE update
		err = service.updateLEAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to update LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// update db
		err = service.storage.PutExistingAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to write to db -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
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

func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) {
	payload, err := decodePayload(r)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to decode payload: %s", err)
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
	err = utils.IsNameValid(*payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// check db for duplicate name? probably unneeded as sql will error on insert
	// email
	err = utils.IsEmailValidOrBlank(*payload.Email)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// private key (pem valid is checked later in account creation)
	// TOS must be accepted
	if *payload.AcceptedTos != true {
		service.logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	///

	// payload -> LE create -> save to db
	err = service.createNewAccount(payload)
	if err != nil {
		service.logger.Printf("accounts: PostNew: failed to create account -- err: %s", err)
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
