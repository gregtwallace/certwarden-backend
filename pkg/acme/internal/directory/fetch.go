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
	d.mu.Lock()
	defer d.mu.Unlock()

	// no-op if update is not due
	if !time.Now().After(d.nextFetchAt) {
		return
	}

	d.logger.Debug("attempting to run update")

	// require directory be validly formatted and start with https://
	if !validation.IsValidHTTPSURL(d.dirURI) {
		d.logger.Error("dirURI is invalid")
		return
	}

	// fetch directory
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.dirURI, http.NoBody)
	if err != nil {
		d.logger.Errorf("failed to create http request (%w)", err)
		return
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.logger.Errorf("failed to do http request (%w)", err)
		return
	}
	defer resp.Body.Close()

	// read and parse body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Errorf("failed to read http response (%w)", err)
		return
	}

	var fetchedDir directory
	err = json.Unmarshal(bodyBytes, &fetchedDir)
	// check for Unmarshal error
	if err != nil {
		d.logger.Errorf("failed to unmarshal json from http response (%w)", err)
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
		d.directoryObj = fetchedDir
		d.logger.Info("directory update detected changes and updated accordingly")
	}

	// update next update time (+ some jitter)
	d.nextFetchAt = time.Now().Add(dirUpdateInterval).Add(time.Duration(rand.IntN(60)) * time.Second)

	d.logger.Debugf("update complete, next run after %s", d.nextFetchAt.Format(time.RFC3339))
}
