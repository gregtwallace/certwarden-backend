package private_keys

import (
	"legocerthub-backend/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Get all of the private keys in our DB and write them as JSON to the API
func (privateKeysDB *PrivateKeysDB) GetAllPrivateKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := privateKeysDB.dbGetAllPrivateKeys()
	if err != nil {
		privateKeysDB.Logger.Printf("privatekeys: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, keys, "private_keys")
	if err != nil {
		privateKeysDB.Logger.Printf("privatekeys: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// Get a single private keys in our DB and write it as JSON to the API
func (privateKeysDB *PrivateKeysDB) GetOnePrivateKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		privateKeysDB.Logger.Printf("privatekeys: GetOne: id param issue -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	key, err := privateKeysDB.dbGetOnePrivateKey(id)
	if err != nil {
		privateKeysDB.Logger.Printf("privatekeys: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, key, "private_key")
	if err != nil {
		privateKeysDB.Logger.Printf("privatekeys: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
