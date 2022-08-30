package challenges

import (
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

	Http01Internal
	Dns01Script
)

// Define MethodDetails which contains details about the defined
// methods.
type MethodDetails struct {
	method Method             `json:"-"`
	Value  string             `json:"value"`
	Name   string             `json:"name"`
	Type   acme.ChallengeType `json:"type"`
}

var methodDetails = []MethodDetails{
	{
		// serve the http record from an internal http server
		method: Http01Internal,
		Value:  "http-01-internal",
		Name:   "HTTP (Self Served)",
		Type:   acme.Http01,
	},
	// TODO: Implement DNS
	// {
	// 	// call external scripts to create and delete dns records
	// 	Method:        Dns01Script,
	// 	Value:         "dns-01-script",
	// 	Name:          "DNS-01 (Manual Script)",
	// 	ChallengeType: acme.Dns01,
	// },
}

// ListOfMethods() returns a constant list of challenge methods
// The Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfMethods() []MethodDetails {
	return methodDetails
}

// MethodByValue returns a challenge method based on its Value.
// If a method isn't found, UnknownMethod is returned.
func MethodByValue(value string) Method {
	for i := range methodDetails {
		if value == methodDetails[i].Value {
			return methodDetails[i].method
		}
	}

	return UnknownMethod
}

// Type returns the Challenge Type for the Method.
func (method Method) Type() acme.ChallengeType {
	for i := range methodDetails {
		if method == methodDetails[i].method {
			return methodDetails[i].Type
		}
	}

	return acme.UnknownChallengeType
}

// validationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge.
func (method Method) validationResource(identifier acme.Identifier, key acme.AccountKey, token string) (name string, content string, err error) {
	return method.Type().ValidationResource(identifier, key, token)
}
