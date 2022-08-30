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

// Define MethodDetails which contains details about the defined
// methods.
type challMethodDetails struct {
	method Method             `json:"-"`
	Value  string             `json:"value"`
	Name   string             `json:"name"`
	Type   acme.ChallengeType `json:"type"`
}

var methodDetails = []challMethodDetails{
	{
		// serve the http record from an internal http server
		method: http01Internal,
		Value:  "http-01-internal",
		Name:   "HTTP (Self Served)",
		Type:   acme.Http01,
	},
	// TODO: Implement DNS
	// {
	// 	// call external scripts to create and delete dns records
	// 	Method:        dns01Script,
	// 	Value:         "dns-01-script",
	// 	Name:          "DNS-01 (Manual Script)",
	// 	ChallengeType: acme.Dns01,
	// },
}

// Method custom JSON Marshal (turns the Method into MethodDetails for output)
func (method *Method) MarshalJSON() (data []byte, err error) {
	// return details marshalled
	return json.Marshal(method.details())
}

// custom UnmarshalJSON not needed at present

// ListOfMethods() returns a slice of challenge methods
func ListOfMethods() (methods []Method) {
	// loop through details to make slice of methods
	for i := range methodDetails {
		methods = append(methods, methodDetails[i].method)
	}

	return methods
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
func (method Method) challType() acme.ChallengeType {
	for i := range methodDetails {
		if method == methodDetails[i].method {
			return methodDetails[i].Type
		}
	}

	return acme.UnknownChallengeType
}

// Type returns the full details for the Method.
func (method Method) details() challMethodDetails {
	for i := range methodDetails {
		if method == methodDetails[i].method {
			return methodDetails[i]
		}
	}

	// no details exist
	return challMethodDetails{}
}

// validationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge.
func (method Method) validationResource(identifier acme.Identifier, key acme.AccountKey, token string) (name string, content string, err error) {
	return method.challType().ValidationResource(identifier, key, token)
}
