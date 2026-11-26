package directory

import (
	"certwarden-backend/pkg/validation"
	"context"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"
	"reflect"
	"time"
)

const dirUpdateInterval = 15 * time.Minute

func (d *DirectoryMgr) updateDirectory(ctx context.Context) {
	// acquire semaphore lock w/ context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := d.semHttp.Acquire(ctx, 1); err != nil {
		// parent ctx aborted or child timer ran out
		return
	}
	defer d.semHttp.Release(1)

	// no-op if update is not due
	if !time.Now().After(d.nextFetchAt) {
		return
	}

	// prevent immediate back-to-back fetches (which would happen on quick failures)
	d.nextFetchAt = time.Now().Add(2 * time.Second).Add(time.Duration(rand.IntN(300)) * time.Millisecond)

	d.logger.Debug("attempting to run update")

	// require directory be validly formatted and start with https://
	if !validation.IsValidHTTPSURL(d.dirURI) {
		d.logger.Error("dirURI is invalid")
		return
	}

	// fetch directory
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.dirURI, http.NoBody)
	if err != nil {
		d.logger.Errorf("failed to create http request (%s)", err)
		return
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.logger.Errorf("failed to do http request (%s)", err)
		return
	}
	defer resp.Body.Close()

	// read and parse body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Errorf("failed to read http response (%s)", err)
		return
	}

	var fetchedDir directory
	err = json.Unmarshal(bodyBytes, &fetchedDir)
	// check for Unmarshal error
	if err != nil {
		d.logger.Errorf("failed to unmarshal json from http response (%s)", err)
		return
	}

	// add raw response to dir for storage
	fetchedDir.raw = bodyBytes

	// check for missing URLs in response
	if fetchedDir.NewNonce == "" {
		d.logger.Error("failed to get 'newNonce' url from response")
		return
	}

	if fetchedDir.NewAccount == "" {
		d.logger.Error("failed to get 'newAccount' url from response")
		return
	}

	if fetchedDir.NewOrder == "" {
		d.logger.Error("failed to get 'newOrder' url from response")
		return
	}

	if fetchedDir.RevokeCert == "" {
		d.logger.Error("failed to get 'revokeCert' url from response")
		return
	}

	// omit NewAuthz as it MUST be omitted if server does not implement pre-authorization
	// omit KeyChange (it isn't strictly required for ACME to work and some providers may not offer it (e.g., DigiCert))

	// Only update if the fetched directory content is different than current
	if !reflect.DeepEqual(fetchedDir, d.directoryObj) {
		d.muData.Lock()
		d.directoryObj = fetchedDir
		d.muData.Unlock()

		d.logger.Info("directory update detected changes and updated accordingly")
	}

	// update next update time (+ some jitter)
	d.nextFetchAt = time.Now().Add(dirUpdateInterval).Add(time.Duration(rand.IntN(60)) * time.Second)

	d.logger.Debugf("update complete, next run after %s", d.nextFetchAt.Format(time.RFC3339))
}
