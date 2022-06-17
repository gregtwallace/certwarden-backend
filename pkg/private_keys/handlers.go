package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Get all of the private keys in our DB and write them as JSON to the API
func (service *Service) GetAllKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := service.storage.GetAllKeys()
	if err != nil {
		service.logger.Printf("keys: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, keys, "private_keys")
	if err != nil {
		service.logger.Printf("keys: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
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

	key, err := service.storage.GetOneKey(id)
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

// Get options for a new private key in our DB and write it as JSON to the API
func (service *Service) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) {
	newKeyOptions := newKeyOptions{}
	newKeyOptions.KeyAlgorithms = utils.ListOfAlgorithms()

	utils.WriteJSON(w, http.StatusOK, newKeyOptions, "private_key_options")
}

// Put (update) a single private key in DB
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

// Post (create) a new single private key in DB
func (service *Service) PostNewKey(w http.ResponseWriter, r *http.Request) {
	var payload KeyPayload
	var err error

	payload, err = decodePayload(r)
	if err != nil {
		service.logger.Printf("keys: PostNew: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// do validation
	// id
	err = utils.IsIdNew(payload.ID)
	if err != nil {
		service.logger.Printf("keys: PostNew: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(*payload.Name)
	if err != nil {
		service.logger.Printf("keys: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// check db for duplicate name? probably unneeded as sql will error on insert

	/// key add method
	// error if no method specified
	if (payload.AlgorithmValue == nil) && (payload.PemContent == nil) {
		err = errors.New("keys: PostNew: no add method specified")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// error if more than one method specified
	if (payload.AlgorithmValue != nil) && (payload.PemContent != nil) {
		err = errors.New("keys: PostNew: multiple add methods specified")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// generate or verify the key
	// generate with algorithm, error if fails
	if payload.AlgorithmValue != nil {
		// must initialize to avoid invalid address
		payload.PemContent = new(string)
		*payload.PemContent, err = utils.GeneratePrivateKeyPem(*payload.AlgorithmValue)
		if err != nil {
			service.logger.Printf("keys: PostNew: failed to generate key pem -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	} else if payload.PemContent != nil {
		// pem inputted - verify pem and determine algorithm
		// must initialize to avoid invalid address
		payload.AlgorithmValue = new(string)
		*payload.PemContent, *payload.AlgorithmValue, err = utils.ValidatePrivateKeyPem(*payload.PemContent)
		if err != nil {
			service.logger.Printf("keys: PostNew: failed to verify pem -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	}
	///

	err = service.storage.PostNewKey(payload)
	if err != nil {
		service.logger.Printf("keys: PostNew: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("keys: PostNew: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// delete a private key
func (service *Service) DeleteKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		service.logger.Printf("keys: Delete: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

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
