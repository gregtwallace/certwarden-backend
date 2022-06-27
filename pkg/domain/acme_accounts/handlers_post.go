package acme_accounts

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
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
	// private key (pem valid is checked later in account creation)
	// TOS must be accepted
	if !*payload.AcceptedTos {
		service.logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	///

	// Save new account details to storage
	// No ACME actions are performed. To actually register, a post should be
	// sent to /register
	service.storage.PostNewAccount(payload)
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
