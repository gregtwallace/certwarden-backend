package acme

import (
	"encoding/json"
	"io"
	"reflect"
	"time"
)

// acmeDirectory struct holds ACME directory object
type directory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	NewAuthz   string `json:"newAuthz"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       struct {
		CaaIdentities  []string `json:"caaIdentities"`
		TermsOfService string   `json:"termsOfService"`
		Website        string   `json:"website"`
	} `json:"meta"`
}

// getAcmeDirectory returns an AcmeDirectory object based on data fetched
// from an ACME directory URI.
func (service *Service) fetchAcmeDirectory() error {
	service.logger.Infof("updating directory from %s", service.dirUri)

	response, err := service.httpClient.Get(service.dirUri)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	// No nonce to save. ACME spec provides nonce on new-nonce requests
	// and replies to POSTs only.

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var fetchedDir directory
	err = json.Unmarshal(body, &fetchedDir)
	if err != nil {
		return err
	} else if reflect.DeepEqual(fetchedDir, *service.dir) {
		// directory already up to date
		service.logger.Infof("directory %s already up to date", service.dirUri)
	} else {
		// fetched directory is different
		*service.dir = fetchedDir
		service.logger.Infof("directory %s updated succesfully", service.dirUri)
	}

	return nil
}

// backgroundDirManager starts a goroutine that is an indefinite for loop
// that checks for directory updates at the specified time interval.
// The interval is shorter if the previous update encountered an error.
func (service *Service) backgroundDirManager() {
	go func() {
		// can adjust wait times if desired
		defaultWaitTime := 24 * time.Hour
		failWaitTime := 15 * time.Minute

		// loop logic for periodic updates
		var waitTime time.Duration

		for {
			err := service.fetchAcmeDirectory()
			if err != nil {
				service.logger.Errorf("directory update failed, will retry shortly: %v", err)
				// if something failed, decrease the wait to try again
				waitTime = failWaitTime
			} else {
				waitTime = defaultWaitTime
			}

			time.Sleep(waitTime)
		}
	}()
}

// TosUrl returns the string for the url where the ToS are located
func (service *Service) TosUrl() string {
	return service.dir.Meta.TermsOfService
}
