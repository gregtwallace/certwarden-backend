package private_keys

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/utils"
	"net/http"
	"regexp"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Get all of the private keys in our DB and write them as JSON to the API
func (privateKeysApp *PrivateKeysApp) GetAllPrivateKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := privateKeysApp.dbGetAllPrivateKeys()
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, keys, "private_keys")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// Get a single private keys in our DB and write it as JSON to the API
func (privateKeysApp *PrivateKeysApp) GetOnePrivateKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetOne: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	key, err := privateKeysApp.dbGetOnePrivateKey(id)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, key, "private_key")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// Put (update) a single private key in DB
func (privateKeysApp *PrivateKeysApp) PutOnePrivateKey(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	var payload privateKeyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = utils.IsIdValidExisting(idParam, payload.ID)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// TODO Update the database and return a proper response to the client
	utils.WriteJSON(w, http.StatusOK, payload, "payload")

}

// TODO
// Post (create) a new single private key in DB
func (privateKeysApp *PrivateKeysApp) PostNewPrivateKey(w http.ResponseWriter, r *http.Request) {
	var payload privateKeyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// do validation
	// id
	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: invalid id param -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} else if id != -1 {
		// the only valid 'new' id is -1
		privateKeysApp.Logger.Println("privatekeys: PostNew: invalid id")
		utils.WriteErrorJSON(w, errors.New("invalid id"))
		return
	}
	// name

	// TODO: Stopped around here

	regex := "/[^-_.~A-z0-9]|[\\^]/g"
	valid, err := regexp.Match(regex, []byte(payload.Name))
	if err != nil {
		privateKeysApp.Logger.Println("privatekeys: PostNew: invalid name")
		utils.WriteErrorJSON(w, errors.New("invalid name"))
	}

	err = utils.WriteJSON(w, http.StatusOK, valid, "temp")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostOne: temp error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

}
