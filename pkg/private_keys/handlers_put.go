package private_keys

import (
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// PutExistingKey updates a key that already exists in storage.
func (service *Service) PutExistingKey(w http.ResponseWriter, r *http.Request) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Printf("keys: PutExisting: invalid idParam -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	var payload KeyPayload
	payload, err = decodePayload(r)
	if err != nil {
		service.logger.Printf("keys: PutOne: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = service.isIdExisting(idParam, payload.ID)
	if err != nil {
		service.logger.Printf("keys: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(*payload.Name)
	if err != nil {
		service.logger.Printf("keys: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// PUT key payload
	err = service.storage.PutExistingKey(payload)
	if err != nil {
		service.logger.Printf("keys: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("keys: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
