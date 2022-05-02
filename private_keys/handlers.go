package private_keys

import (
	"encoding/json"
	"legocerthub-backend/utils"
	"net/http"
	"strconv"
	"time"

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

	// load fields that are permitted to be updated
	var privateKey privateKeyDb
	privateKey.id, err = strconv.Atoi(payload.ID)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	privateKey.name = payload.Name
	privateKey.description.String = payload.Description
	privateKey.updatedAt = int(time.Now().Unix())

	err = privateKeysApp.dbPutExistingPrivateKey(privateKey)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := jsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// Post (create) a new single private key in DB
func (privateKeysApp *PrivateKeysApp) PostNewPrivateKey(w http.ResponseWriter, r *http.Request) {
	var payload privateKeyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// do validation
	// id
	err = utils.IsIdValidNew(payload.ID)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} // check db for duplicate name? probably unneeded as sql will error on insert
	// algorithm - returns error if value is invalid, generates key if valid
	pem, err := generatePrivateKeyPem(payload.AlgorithmValue)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to generate key pem -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// TODO PEM Content and PEM File and Logic to ensure only 1 of the 3.
	///

	// load fields
	var privateKey privateKeyDb
	privateKey.name = payload.Name
	privateKey.description.String = payload.Description
	privateKey.algorithmValue = payload.AlgorithmValue
	privateKey.pem = pem
	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to generate api key -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	privateKey.apiKey = apiKey
	privateKey.createdAt = int(time.Now().Unix())
	privateKey.updatedAt = privateKey.createdAt

	err = privateKeysApp.dbPostNewPrivateKey(privateKey)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := jsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
