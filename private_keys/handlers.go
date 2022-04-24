package private_keys

import (
	"legocerthub-backend/utils"
	"log"
	"net/http"
)

func (privateKeys *PrivateKeys) GetAllPrivateKeys(w http.ResponseWriter, r *http.Request) {

	accounts, err := privateKeys.dbGetAllPrivateKeys()
	if err != nil {
		log.Printf("Failed to get all private keys %s", err)
	}

	utils.WriteJSON(w, http.StatusOK, accounts, "private_keys")

}
