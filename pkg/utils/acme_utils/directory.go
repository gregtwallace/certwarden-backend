package acme_utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Directory URLs for Let's Encrypt
const LeProdDirectory string = "https://acme-v02.api.letsencrypt.org/directory"
const LeStagingDirectory string = "https://acme-staging-v02.api.letsencrypt.org/directory"

// GetAcmeDirectory returns an AcmeDirectory object based on data fetched
// from an ACME directory URL.
func GetAcmeDirectory(env string) (AcmeDirectory, error) {
	var response *http.Response
	var err error

	switch env {
	case "prod":
		response, err = http.Get(LeProdDirectory)
	case "staging":
		response, err = http.Get(LeStagingDirectory)
	default:
		return AcmeDirectory{}, errors.New("invalid environment")
	}

	if err != nil {
		return AcmeDirectory{}, err
	}
	defer response.Body.Close()

	// No nonce to save. ACME spec provides nonce on new-nonce requests
	// and replies to POSTs only.

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return AcmeDirectory{}, err
	}

	var directory = AcmeDirectory{}
	err = json.Unmarshal(body, &directory)
	if err != nil {
		return AcmeDirectory{}, err
	}

	return directory, nil
}
