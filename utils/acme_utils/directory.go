package acme_utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Directory URLs for Let's Encrypt
const leProdDirectory string = "https://acme-v02.api.letsencrypt.org/directory"
const leStagingDirectory string = "https://acme-staging-v02.api.letsencrypt.org/directory"

// Directory struct that holds the returned data from querying directory URL
type AcmeDirectory struct {
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

// GetAcmeDirectory returns an AcmeDirectory object based on data fetched
// from an ACME directory URL.
func GetAcmeDirectory(env string) (AcmeDirectory, error) {
	var response *http.Response
	var err error

	switch env {
	case "prod":
		response, err = http.Get(leProdDirectory)
	case "staging":
		response, err = http.Get(leStagingDirectory)
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
