package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/utils"
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

// PutExistingKey updates a key that already exists in storage.
// only the name and description are allowed to be modified

// PutNameDescAccount
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
