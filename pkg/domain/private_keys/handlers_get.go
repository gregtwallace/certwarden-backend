package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllKeys returns all of the private keys in storage as JSON
func (service *Service) GetAllKeys(w http.ResponseWriter, r *http.Request) error {

	keys, err := service.storage.GetAllKeys()
	if err != nil {
		return err
	}

	err = utils.WriteJSON(w, http.StatusOK, keys, "private_keys")
	if err != nil {
		return err
	}

	return nil
}

// Get a single private keys in our DB and write it as JSON to the API
func (service *Service) GetOneKey(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Printf("keys: GetOne: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// if id is new provide algo options list
	err = utils.IsIdNew(&id)
	if err == nil {
		// run the new key options handler if the id is new
		service.GetNewKeyOptions(w, r)
		return
	} else if id < 0 {
		// if id < 0 & not new, it is definitely not valid
		err = errors.New("keys: GetOne: id param is invalid (less than 0 and not new)")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	key, err := service.storage.GetOneKeyById(id)
	if err != nil {
		service.logger.Printf("keys: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, key, "private_key")
	if err != nil {
		service.logger.Printf("keys: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetNewKeyOptions returns configuration options for a new private key as JSON
func (service *Service) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) {
	newKeyOptions := newKeyOptions{}
	newKeyOptions.KeyAlgorithms = key_crypto.ListOfAlgorithms()

	err := utils.WriteJSON(w, http.StatusOK, newKeyOptions, "private_key_options")
	if err != nil {
		service.logger.Printf("keys: GetNewKeyOptions: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
