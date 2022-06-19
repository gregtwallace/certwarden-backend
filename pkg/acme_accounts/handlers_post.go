package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
)

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
