package private_keys

import (
	"legocerthub-backend/utils"
	"log"
	"net/http"
)

func (privateKeysDB *PrivateKeysDB) GetAllPrivateKeys(w http.ResponseWriter, r *http.Request) {

	accounts, err := privateKeysDB.dbGetAllPrivateKeys()
	if err != nil {
		log.Printf("Failed to get all private keys %s", err)
	}

	utils.WriteJSON(w, http.StatusOK, accounts, "private_keys")

}
