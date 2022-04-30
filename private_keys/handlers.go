package private_keys

import (
	"legocerthub-backend/utils"
	"net/http"
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
		privateKeysApp.Logger.Printf("privatekeys: GetOne: id param issue -- err: %s", err)
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
	params := httprouter.ParamsFromContext(r.Context())

	err := utils.WriteJSON(w, http.StatusOK, params, "status")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PutOne: temp error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

}

// Post (create) a new single private key in DB
func (privateKeysApp *PrivateKeysApp) PostNewPrivateKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	err := utils.WriteJSON(w, http.StatusOK, params, "status")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: PostOne: temp error -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

}
