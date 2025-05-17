package acme

import (
	"certwarden-backend/pkg/randomness"
	"certwarden-backend/pkg/validation"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
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

		// ACME Profiles Extension - https://datatracker.ietf.org/doc/draft-aaron-acme-profiles/
		Profiles map[string]string `json:"profiles,omitempty"`
	} `json:"meta"`

	// ACME ARI (ACME Renewal Information) Extension - https://datatracker.ietf.org/doc/draft-ietf-acme-ari/
	RenewalInfo *string `json:"renewalInfo,omitempty"`

	// raw stores the directory URLs raw response
	raw json.RawMessage `json:"-"`
}

// FetchAcmeDirectory uses the specified httpclient to fetch the specified
// dirUri and return a directory object. If the directory fails to fetch or what
// is fetched is invalid, an error is returned.
func FetchAcmeDirectory(httpClient *http.Client, dirUrl string) (directory, error) {

	// require directory be validly formatted and start with https://
	if !validation.HttpsUrlValid(dirUrl) {
		return directory{}, fmt.Errorf("acme: directory url (%s) must be a properly formatted url and start with 'https://'", dirUrl)
	}

	response, err := httpClient.Get(dirUrl)
	if err != nil {
		return directory{}, err
	}
	defer response.Body.Close()

	// No nonce to save. ACME spec provides nonce on new-nonce requests
	// and replies to POSTs only.

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return directory{}, err
	}
	var fetchedDir directory
	err = json.Unmarshal(bodyBytes, &fetchedDir)
	// check for Unmarshal error
	if err != nil {
		return directory{}, err
	}

	// add raw response to dir for storage
	fetchedDir.raw = bodyBytes

	// check for missing URLs in response
	if fetchedDir.NewNonce == "" ||
		fetchedDir.NewAccount == "" ||
		fetchedDir.NewOrder == "" ||
		// omit NewAuthz as it MUST be omitted if server does not implement pre-authorization
		// omit KeyChange (it isn't strictly required for ACME to work and some providers may not offer it (e.g., DigiCert))
		fetchedDir.RevokeCert == "" {
		return directory{}, fmt.Errorf("acme: directory (%s) missing one or more required urls", dirUrl)
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
func (service *Service) backgroundDirManager(shutdownCtx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Infof("starting acme directory refresh service (%s)", service.dirUri)

	// notify func to log dir update fails
	notifyFunc := func(err error, dur time.Duration) {
		service.logger.Errorf("directory %s update failed (%s), will retry again in %s", service.dirUri, err, dur.Round(100*time.Millisecond))
	}

	// normal run hour
	refreshHour := 1 // 1am

	// service routine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// backoff object for retry of failed dir update
			// not using randomness ACME bo because this one needs to be indefinite and have slower timing
			// not as aggressive because this should only fail if the service is totally down or undergoing maintenance
			bo := backoff.NewExponentialBackOff()
			bo.InitialInterval = 10 * time.Second
			bo.RandomizationFactor = 0.5
			bo.Multiplier = 2
			bo.MaxInterval = 15 * time.Minute
			bo.MaxElapsedTime = 0 // never stop trying

			boWithContext := backoff.WithContext(bo, shutdownCtx)

			// try update (and retry if failed - use exponential backoff)
			// will only return err if CTX is done (shutdown), otherwise never err because MaxElapsedTime is 0
			// thus, ignore err here and let shutdown below handle it
			_ = backoff.RetryNotify(service.updateAcmeServiceDirectory, boWithContext, notifyFunc)

			// update has now succeeded, schedule next regular update (omit minute and second for now)
			nextRunTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
				refreshHour, 0, 0, 0, time.Local)

			// if today's run already passed, run tomorrow
			if !nextRunTime.After(time.Now()) {
				nextRunTime = nextRunTime.Add(24 * time.Hour)
			}

			// add random minute and second after above calc to avoid duplicate runs on same day
			// (in event random #s are larger than previous, e.g. run at 5 min then 18 min it would
			// run at 1:05 and 1:18 the same day)
			// this randomness also prevents updating all directories at the same exact time
			// add 1 minute to avoid extreme edge case where this code runs at exactly run hour
			nextRunTime = nextRunTime.
				Add(time.Duration(randomness.GenerateInsecureInt(60)+1) * time.Minute).
				Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)

			service.logger.Debugf("next directory refresh for %s will occur at %s", service.dirUri, nextRunTime.String())

			select {
			case <-shutdownCtx.Done():
				// end routine
				service.logger.Infof("acme directory refresh service shutdown complete (%s)", service.dirUri)
				return

			case <-time.After(time.Until(nextRunTime)):
				// delay until next regular run
			}
		}
	}()
}

// TosUrl returns the string for the url where the ToS are located
func (service *Service) TosUrl() string {
	return service.dir.Meta.TermsOfService
}

// RequiresEAB returns if the acme server requires External Account Binding
func (service *Service) RequiresEAB() bool {
	return service.dir.Meta.ExternalAccountRequired
}

// DirectoryRawResponse returns the ACME Service's raw directory response
func (service *Service) DirectoryRawResponse() json.RawMessage {
	return service.dir.raw
}

// Profiles returns the map of available profiles for the ACME service, or nil if
// the profiles extension is not implemented (per the directory's meta response)
func (service *Service) Profiles() map[string]string {
	return service.dir.Meta.Profiles
}

// ProfileValidate returns true if the specified profileName exists in the ACME
// service's meta profile map. If the profile does not exist (including if there
// is no profile map at all), false is returned.
func (service *Service) ProfileValidate(profileName string) bool {
	for k := range service.dir.Meta.Profiles {
		if k == profileName {
			return true
		}
	}

	return false
}
