package certificates

import (
	"certwarden-backend/pkg/output"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DeleteCert deletes a cert from storage
func (service *Service) DeleteCert(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// verify cert id exists
	_, outErr := service.GetCertificate(id)
	if outErr != nil {
		return outErr
	}

	// delete from storage
	err = service.storage.DeleteCert(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// write response
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("deleted certificate (id: %d)", id)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// RemoveOldApiKey discards a cert's api_key, replaces it with the key's
// api_key_new, and then blanks api_key_new
func (service *Service) RemoveOldApiKey(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// get cert (validate exists)
	cert, outErr := service.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}

	// verify new api key is not empty (need something to promote)
	if cert.ApiKeyNew == "" {
		service.logger.Debug(errors.New("new api key does not exist"))
		return output.ErrValidationFailed
	}
	// validation -- end

	// update storage
	// set current api key from new key
	err = service.storage.PutCertApiKey(certId, cert.ApiKeyNew, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	cert.ApiKey = cert.ApiKeyNew

	// set new key to blank
	err = service.storage.PutCertNewApiKey(certId, "", int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	cert.ApiKeyNew = ""

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "certificate old api key deleted, new api key promoted"
	response.Certificate = cert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// DisableClientKey discards a cert's client key (replacing it with a blank string,
// which disables the client functionality)
func (service *Service) DisableClientKey(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// get cert (validate exists)
	cert, outErr := service.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}
	// validation -- end

	// update storage
	err = service.storage.PutCertClientKey(certId, "", int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	cert.PostProcessingClientKeyB64 = ""

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "certificate client key deleted (disabled)"
	response.Certificate = cert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
