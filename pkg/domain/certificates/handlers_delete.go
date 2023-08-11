package certificates

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DeleteCert deletes a cert from storage
func (service *Service) DeleteCert(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// verify cert id exists
	_, err = service.GetCertificate(id)
	if err != nil {
		return err
	}

	// delete from storage
	err = service.storage.DeleteCert(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "deleted",
		ID:      id,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}

// RemoveOldApiKey discards a cert's api_key, replaces it with the key's
// api_key_new, and then blanks api_key_new
func (service *Service) RemoveOldApiKey(w http.ResponseWriter, r *http.Request) (err error) {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// get cert (validate exists)
	cert, err := service.GetCertificate(certId)
	if err != nil {
		return err
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
	// set new key to blank
	err = service.storage.PutCertNewApiKey(certId, "", int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "new api key promoted", // TODO?
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
