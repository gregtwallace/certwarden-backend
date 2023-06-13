package challenges

import (
	"legocerthub-backend/pkg/acme"
)

// Define challenge methods (which are more than just a challenge
// type). This allows for multiple methods using the same RFC 8555
// challenge type.

// MethodValue is the string value that methods are stored in
// storage as and is also used to refer to specific, unique methods
type MethodValue string

// Method contains all of the details about a defined method.
type Method struct {
	Value         MethodValue        `json:"value"`
	Name          string             `json:"name"`
	ChallengeType acme.ChallengeType `json:"type"`
}

// validationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge using the Method.
func (method Method) validationResource(identifier acme.Identifier, key acme.AccountKey, token string) (name string, content string, err error) {
	return method.ChallengeType.ValidationResource(identifier, key, token)
}

func (method Method) addStatus(enabled bool) MethodWithStatus {
	return MethodWithStatus{
		Method:  method,
		Enabled: enabled,
	}
}

// Define values. These values should be assigned once and NEVER
// changed to avoid storage issues.
const (
	unknownMethodValue         MethodValue = ""
	methodValueHttp01Internal  MethodValue = "http-01-internal"
	methodValueDns01Manual     MethodValue = "dns-01-manual"
	methodValueDns01AcmeDns    MethodValue = "dns-01-acme-dns"
	methodValueDns01AcmeSh     MethodValue = "dns-01-acme-sh"
	methodValueDns01Cloudflare MethodValue = "dns-01-cloudflare"
)

// UnknownMethod is used when a Method does not match any known Method.
var UnknownMethod = Method{
	Value:         unknownMethodValue,
	Name:          "Unknown Method",
	ChallengeType: acme.UnknownChallengeType,
}

var allMethods = []Method{
	{
		// serve the http record from an internal http server
		Value:         methodValueHttp01Internal,
		Name:          "HTTP on API Server",
		ChallengeType: acme.ChallengeTypeHttp01,
	},
	{
		// use external scripts to create and delete dns records
		Value:         methodValueDns01Manual,
		Name:          "DNS Manual Script",
		ChallengeType: acme.ChallengeTypeDns01,
	},
	{
		// updates dns record values on acme-dns
		Value:         methodValueDns01AcmeDns,
		Name:          "DNS acme-dns",
		ChallengeType: acme.ChallengeTypeDns01,
	},
	{
		// updates dns record values using acme.sh script
		Value:         methodValueDns01AcmeSh,
		Name:          "DNS acme.sh Script",
		ChallengeType: acme.ChallengeTypeDns01,
	},
	{
		// create and delete dns records on Cloudflare
		Value:         methodValueDns01Cloudflare,
		Name:          "DNS Cloudflare",
		ChallengeType: acme.ChallengeTypeDns01,
	},
}

// MethodByStorageValue returns a challenge method based on its Value.
// If a method isn't found, UnknownMethod is returned.
func MethodByStorageValue(value MethodValue) Method {
	for i := range allMethods {
		if value == allMethods[i].Value {
			return allMethods[i]
		}
	}

	return UnknownMethod
}

// MethodWithStatus is a struct to return service status with a Method
type MethodWithStatus struct {
	Method
	Enabled bool `json:"enabled"`
}

// ListOfMethodsWithStatus returns a slice of all possible challenge methods
// with an additional field indicating if each is currently enabled in the challenges
// service.
func (service *Service) ListOfMethodsWithStatus() []MethodWithStatus {
	return service.methods
}
