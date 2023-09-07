package acme

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/randomness"
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
		// run hour and fail wait duration
		refreshHour := 1 // 1am
		failWaitTime := 15 * time.Minute

		// loop logic for periodic updates
		var waitTime time.Duration

		for {
			err := service.updateAcmeServiceDirectory()
			if err != nil {
				service.logger.Errorf("directory %s update failed, will retry shortly (%v)", service.dirUri, err)
				// if something failed, use failed wait time
				waitTime = failWaitTime
			} else {
				// if not failed, schedule next run (omit minute and second for now)
				nextRunTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
					refreshHour, 0, 0, 0, time.Local)

				// if today's run already passed, run tomorrow
				if !nextRunTime.After(time.Now()) {
					nextRunTime = nextRunTime.Add(24 * time.Hour)
				}

				// add random minute and second after above calc to avoid duplicate runs on same day
				// (in event random #s are larger than previous, e.g. run at 5 min then 18 min it would
				// run at 1:05 and 1:18 the same day)
				// this randomness also prevents LeGo from updating all directories at the same exact time
				nextRunTime = nextRunTime.
					Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Minute).
					Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)

				service.logger.Debugf("next directory refresh for %s will occur at %s", service.dirUri, nextRunTime.String())

				// set as duration
				waitTime = time.Until(nextRunTime)
			}

			// sleep or wait for shutdown context to be done
			select {
			case <-ctx.Done():
				// close routine
				service.logger.Infof("acme directory refresh service shutdown complete (%s)", service.dirUri)
				wg.Done()
				return

			case <-time.After(waitTime):
				// sleep until retry
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
