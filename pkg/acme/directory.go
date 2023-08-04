package acme

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"legocerthub-backend/pkg/httpclient"
	"reflect"
	"sync"
	"time"
)

var (
	ErrDirMissingUrl = errors.New("missing required url(s)")
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
		TermsOfService          string   `json:"termsOfService"`
		Website                 string   `json:"website"`
		CaaIdentities           []string `json:"caaIdentities"`
		ExternalAccountRequired bool     `json:"externalAccountRequired"`
	} `json:"meta"`
}

// FetchAcmeDirectory uses the specified httpclient to fetch the specified
// dirUri and return a directory object. If the directory fails to fetch or what
// is fetched is invalid, an error is returned.
func FetchAcmeDirectory(httpClient *httpclient.Client, dirUri string) (directory, error) {
	response, err := httpClient.Get(dirUri)
	if err != nil {
		return directory{}, err
	}
	defer response.Body.Close()

	// No nonce to save. ACME spec provides nonce on new-nonce requests
	// and replies to POSTs only.

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return directory{}, err
	}
	var fetchedDir directory
	err = json.Unmarshal(body, &fetchedDir)
	// check for Unmarshal error
	if err != nil {
		return directory{}, err
	}

	// check for missing URLs in response
	if fetchedDir.NewNonce == "" ||
		fetchedDir.NewAccount == "" ||
		fetchedDir.NewOrder == "" ||
		// omit NewAuthz as it MUST be omitted if server does not implement pre-authorization
		fetchedDir.RevokeCert == "" ||
		fetchedDir.KeyChange == "" {
		return directory{}, ErrDirMissingUrl
	}

	return fetchedDir, nil
}

// upsateAcmeServiceDirectory updates the Service's directory object based on
// data fetched from the Service's directory URI. If it fails, an error is
// returned.
func (service *Service) updateAcmeServiceDirectory() error {
	service.logger.Infof("updating directory from %s", service.dirUri)

	// try to fetch the directory
	fetchedDir, err := FetchAcmeDirectory(service.httpClient, service.dirUri)
	if err != nil {
		return err
	}

	// Only update if the fetched directory content is different than current
	if reflect.DeepEqual(fetchedDir, *service.dir) {
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
func (service *Service) backgroundDirManager(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Infof("starting acme directory refresh service (%s)", service.dirUri)
	wg.Add(1)

	// service routine
	go func(service *Service, ctx context.Context, wg *sync.WaitGroup) {
		// can adjust wait times if desired
		defaultWaitTime := 24 * time.Hour
		failWaitTime := 15 * time.Minute

		// loop logic for periodic updates
		var waitTime time.Duration

		for {
			err := service.updateAcmeServiceDirectory()
			if err != nil {
				service.logger.Errorf("directory update failed, will retry shortly (%v)", err)
				// if something failed, decrease the wait to try again
				waitTime = failWaitTime
			} else {
				waitTime = defaultWaitTime
			}

			// sleep or wait for shutdown context to be done
			select {
			case <-ctx.Done():
				// close routine
				service.logger.Infof("acme directory refresh service shutdown complete (%s)", service.dirUri)
				wg.Done()
				return

			case <-time.After(waitTime):
				// sleep and retry
			}
		}
	}(service, ctx, wg)
}

// TosUrl returns the string for the url where the ToS are located
func (service *Service) TosUrl() string {
	return service.dir.Meta.TermsOfService
}

// RequiresEAB returns if the acme server requires External Account Binding
func (service *Service) RequiresEAB() bool {
	return service.dir.Meta.ExternalAccountRequired
}
