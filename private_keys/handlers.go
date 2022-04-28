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

func (privateKeysApp *PrivateKeysApp) GetAllKeyAlgorithms(w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJSON(w, http.StatusOK, supportedKeyAlgorithms(), "key_algorithms")
	if err != nil {
		privateKeysApp.Logger.Printf("privatekeys: GetAllKeyAlgorithms: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
