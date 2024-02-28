package sqlite

import (
	"encoding/json"
	"legocerthub-backend/pkg/domain/certificates"
)

// jsonStringSlice is a string type in storage that is a json formatted
// array of strings
type jsonStringSlice string

// transform JSS into string slice
func (jss jsonStringSlice) toSlice() []string {
	if jss == "" {
		return []string{}
	}

	strSlice := []string{}
	err := json.Unmarshal([]byte(jss), &strSlice)
	if err != nil {
		return []string{}
	}

	return strSlice
}

// makeCommaJoinedString creates a JSS from a slice of strings
func makeJsonStringSlice(stringSlice []string) jsonStringSlice {
	if len(stringSlice) == 0 {
		return "[]"
	}

	jss, err := json.Marshal(stringSlice)
	if err != nil {
		return "[]"
	}

	return jsonStringSlice(jss)
}

// jsonCertExtensionSlice is a json formatted string that is a slice of CertExtension
type jsonCertExtensionSlice string

// transform JCES into a slice of proper CertExtension
func (jces jsonCertExtensionSlice) toCertExtensionSlice() ([]certificates.CertExtension, error) {
	if jces == "" {
		return []certificates.CertExtension{}, nil
	}

	// unmarshal the json to the json object
	extSlice := []certificates.CertExtensionJSON{}
	err := json.Unmarshal([]byte(jces), &extSlice)
	if err != nil {
		return nil, err
	}

	// convert json objs to real objs
	certExtSlice := []certificates.CertExtension{}
	for i := range extSlice {
		certExt, err := extSlice[i].ToCertExtension()
		if err != nil {
			// if invalid data stored, return err
			return nil, err
		}
		certExtSlice = append(certExtSlice, certExt)
	}

	return certExtSlice, nil
}

// makeJsonCertExtensionSlice creates a JCES from a slice of CertExtensionJSON
func makeJsonCertExtensionSlice(extensionSlice []certificates.CertExtensionJSON) jsonCertExtensionSlice {
	if len(extensionSlice) == 0 {
		return "[]"
	}

	jpes, err := json.Marshal(extensionSlice)
	if err != nil {
		return "[]"
	}

	return jsonCertExtensionSlice(jpes)
}
