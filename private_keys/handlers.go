package private_keys

import (
	"encoding/json"
	"errors"
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
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// if id is new provide algo options list
	err := utils.IsIdValidNew(idParam)
	if err == nil {
		// run the new key options handler if the id is new
		privateKeysApp.GetNewKeyOptions(w, r)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetOne: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	if id < 0 {
		// if id < 0, it is definitely not valid
		err = errors.New("privatekeys: GetOne: id param is invalid (less than 0 and not new)")
		privateKeysApp.Logger.Println(err)
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

// Get options for a new private key in our DB and write it as JSON to the API
func (privateKeysApp *PrivateKeysApp) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) {
	newKeyOptions := NewPrivateKeyOptions{}
	newKeyOptions.KeyAlgorithms = listOfAlgorithms()

	utils.WriteJSON(w, http.StatusOK, newKeyOptions, "private_key_options")
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

	privateKey.description.Valid = true
	privateKey.description.String = payload.Description

	privateKey.updatedAt = int(time.Now().Unix())

	err = privateKeysApp.dbPutExistingPrivateKey(privateKey)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
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

	/// key add method
	// error if no method specified
	if (payload.AlgorithmValue == "") && (payload.PemContent == "") {
		err = errors.New("privatekeys: PostNew: no add method specified")
		privateKeysApp.Logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// error if more than one method specified
	if (payload.AlgorithmValue != "") && (payload.PemContent != "") {
		err = errors.New("privatekeys: PostNew: multiple add methods specified")
		privateKeysApp.Logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// generate or verify the key
	var pem, algorithmValue string
	// generate with algorithm, error if fails
	if payload.AlgorithmValue != "" {
		pem, err = generatePrivateKeyPem(payload.AlgorithmValue)
		if err != nil {
			privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to generate key pem -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
		algorithmValue = payload.AlgorithmValue
	} else if payload.PemContent != "" {
		// pem inputted - verify pem and determine algorithm
		pem, algorithmValue, err = validatePrivateKeyPem(payload.PemContent)
		if err != nil {
			privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to verify pem -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	}
	///

	// generate api key
	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to generate api key -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// load fields
	var privateKey privateKeyDb
	privateKey.name = payload.Name

	privateKey.description.Valid = true
	privateKey.description.String = payload.Description

	privateKey.algorithmValue = algorithmValue
	privateKey.pem = pem
	privateKey.apiKey = apiKey
	privateKey.createdAt = int(time.Now().Unix())
	privateKey.updatedAt = privateKey.createdAt

	err = privateKeysApp.dbPostNewPrivateKey(privateKey)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostNew: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// delete a private key
func (privateKeysApp *PrivateKeysApp) DeletePrivateKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: Delete: id param error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = privateKeysApp.dbDeletePrivateKey(id)
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: Delete: failed to db delete -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: Delete: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
