package private_keys

import (
	"errors"
	"legocerthub-backend/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (privateKeysDB *PrivateKeysDB) GetAllPrivateKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := privateKeysDB.dbGetAllPrivateKeys()
	if err != nil {
		log.Printf("Failed to get all private keys %s", err)
	}

	utils.WriteJSON(w, http.StatusOK, keys, "private_keys")

}

func (privateKeysDB *PrivateKeysDB) GetOnePrivateKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		privateKeysDB.Logger.Println(errors.New("invalid id parameter"))
		// Write error json
		return
	}

	key, err := privateKeysDB.dbGetOnePrivateKey(id)
	if err != nil {
		// Write error json
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, key, "private_key")
	if err != nil {
		// Write error json
		return
	}

}
