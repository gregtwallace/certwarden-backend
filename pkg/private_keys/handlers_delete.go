package private_keys

import (
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteKey deletes a private key from storage
func (service *Service) DeleteKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		service.logger.Printf("keys: Delete: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// TODO: Validate note in use, though storage may also do this

	err = service.storage.DeleteKey(id)
	if err != nil {
		service.logger.Printf("keys: Delete: failed to db delete -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("keys: Delete: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
