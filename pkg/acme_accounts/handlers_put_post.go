package acme_accounts

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (htmlPayload *accountHtmlPayload) toPayload() (AccountPayload, error) {
	var payload AccountPayload
	var err error

	payload.ID, err = StringToInt(htmlPayload.ID)
	if err != nil {
		return AccountPayload{}, err
	}

	payload.Name = htmlPayload.Name

	payload.Description = htmlPayload.Description

	payload.Email = htmlPayload.Email

	payload.PrivateKeyID, err = StringToInt(htmlPayload.PrivateKeyID)
	if err != nil {
		return AccountPayload{}, err
	}

	payload.AcceptedTos = StringToBool(htmlPayload.AcceptedTos)

	payload.IsStaging = StringToBool(htmlPayload.IsStaging)

	return payload, nil
}

func decodePayload(r *http.Request) (AccountPayload, error) {
	var payload AccountPayload

	// read body to allow for multiple unmarshal attempts
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return AccountPayload{}, err
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		var htmlPayload accountHtmlPayload
		err = json.Unmarshal(body, &htmlPayload)
		if err != nil {
			return AccountPayload{}, err
		}
		payload, err = htmlPayload.toPayload()
		if err != nil {
			return AccountPayload{}, err
		}
	}

	return payload, nil
}

// TODO move these to a common package
func StringToInt(s *string) (*int, error) {
	i := new(int)
	var err error

	// return i as a nil pointer if s is a nil pointer
	if s == nil {
		i = nil
	} else {
		// else try to convert s into an int and return the pointer to that
		*i, err = strconv.Atoi(*s)
		if err != nil {
			return i, err
		}
	}

	return i, nil
}

func StringToBool(s string) bool {
	var b bool

	if s == "true" {
		b = true
	} else {
		b = false
	}

	return b
}

// PutOneAccount is an http handler that overwrites the specified (by id) acme account with the
//  data PUT by the client

// TODO: fetch data from LE and update kid
// if kid already exists, use it to update the email field

func (service *Service) PutOneAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	payload, err := decodePayload(r)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to decode payload: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	// TODO change IsIdValid to use ints
	payloadId := strconv.Itoa(*payload.ID)

	err = utils.IsIdValidExisting(idParam, payloadId)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		service.logger.Printf("accounts: PutOne: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// check db for account kid
	kid, err := service.storage.GetAccountKid(*payload.ID)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to check for kid -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// if no kid, do new LE registration
	if kid == "" {
		// always update name and desc
		err = service.storage.PutExistingAccountNameDesc(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to update account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// do new LE reg
		err = service.createNewLEAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to create new LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	} else {
		// if there is a kid, update the existing account with new data
		// do LE update
		err = service.updateLEAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to update LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// update db
		err = service.storage.PutExistingAccount(payload)
		if err != nil {
			service.logger.Printf("accounts: PutOne: failed to write to db -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("accounts: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) {
	payload, err := decodePayload(r)
	if err != nil {
		service.logger.Printf("accounts: PutOne: failed to decode payload: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// do validation
	// id
	// TODO change IsIdValid to use ints
	payloadId := strconv.Itoa(*payload.ID)

	err = utils.IsIdValidNew(payloadId)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} // check db for duplicate name? probably unneeded as sql will error on insert
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		service.logger.Printf("accounts: PostNew: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// private key (pem valid is checked later in account creation)
	// TOS must be accepted
	if payload.AcceptedTos != true {
		service.logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	///

	// payload -> LE create -> save to db
	err = service.createNewAccount(payload)
	if err != nil {
		service.logger.Printf("accounts: PostNew: failed to create account -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		service.logger.Printf("accounts: PostNew: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
