package acme_accounts

import (
	"certwarden-backend/pkg/output"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// UpdatePayload is the struct for editing an existing account; most fields
// are inaccessible via json marshal to prevent an api user from manually
// editing them
// DO NOT change `-` fields json value!
type UpdatePayload struct {
	ID          int        `json:"-"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Status      *string    `json:"-"`
	Email       *string    `json:"-"`
	KID         *string    `json:"-"`
	CreatedAt   *time.Time `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}

// UpdatePayload is a handler that sets the name and description of an account
// within storage
func (service *Service) PutNameDescAccount(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// payload decoding
	var payload UpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// id
	_, outErr := service.getAccount(payload.ID)
	if outErr != nil {
		return outErr
	}

	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.JsonErrValidationFailed(ErrNameBad)
	}
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = time.Now()

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	updatedAcct, err := service.storage.PutAcmeAccountUpdate(payload)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	detailedResp, err := updatedAcct.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated account"
	response.Account = detailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
