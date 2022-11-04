package challenges

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/pkg/acme"
)

var errUnsupportedMethod = errors.New("unsupported challenge method")

// Define challenge methods (which are more than just a challenge
// type). This allows for multiple methods using the same RFC 8555
// challenge type.
type Method int

const (
	UnknownMethod Method = iota

	http01Internal
	dns01Script
)

// Method custom JSON Marshal (turns the Method into MethodDetails for output)
func (method Method) MarshalJSON() (data []byte, err error) {
	// get method details
	details := method.details()

	// put the exportable details into an exportable struct
	output := struct {
		StorageValue  string             `json:"value"`
		Name          string             `json:"name"`
		ChallengeType acme.ChallengeType `json:"type"`
	}{
		StorageValue:  details.storageValue,
		Name:          details.name,
		ChallengeType: details.challengeType,
	}

	// return details marshalled
	return json.Marshal(output)
}

// custom UnmarshalJSON not needed at present

// details returns the full details for the Method.
func (method Method) details() challMethodDetails {
	for i := range methodDetails {
		if method == methodDetails[i].method {
			return methodDetails[i]
		}
	}

	// no details exist
	return challMethodDetails{}
}

// Type returns the Challenge Type for the Method.
func (method Method) challengeType() acme.ChallengeType {
	return method.details().challengeType
}

// validationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge.
func (method Method) validationResource(identifier acme.Identifier, key acme.AccountKey, token string) (name string, content string, err error) {
	return method.challengeType().ValidationResource(identifier, key, token)
}
