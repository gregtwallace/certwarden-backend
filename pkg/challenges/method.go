package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
)

var errNoProviders = errors.New("no challenge providers are properly configured (at least one must be enabled)")

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
	Enabled       bool               `json:"enabled"`
	ChallengeType acme.ChallengeType `json:"type"`
}

// Define values. These values should be assigned once and NEVER
// changed to avoid storage issues.
const (
	unknownMethodValue MethodValue = ""

	methodValueHttp01Internal  = "http-01-internal"
	methodValueDns01Manual     = "dns-01-manual"
	methodValueDns01Cloudflare = "dns-01-cloudflare"
)

// UnknownMethod is used when a Method does not match any known Method.
var UnknownMethod = Method{
	Value:         unknownMethodValue,
	Name:          "Unknown Method",
	ChallengeType: acme.UnknownChallengeType,
}

// Configure the service's methods by loading them and determining
// if they're currently enabled.
func (service *Service) configureMethods() error {

	// methods contains the details corresponding to all of the defined
	// methodValues
	var methodsList = []Method{
		// Note: Enabled is configured later by the service
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
			// create and delete dns records on Cloudflare
			Value:         methodValueDns01Cloudflare,
			Name:          "DNS Cloudflare",
			ChallengeType: acme.ChallengeTypeDns01,
		},
	}

	// range through MethodDetailed to set the Enabled field according
	// to the Service configuration & confirm at least one method is actually
	// enabled.
	atLeastOneEnabled := false
	for i := range methodsList {
		if _, ok := service.providers[methodsList[i].Value]; ok {
			methodsList[i].Enabled = true
			atLeastOneEnabled = true
		}
	}
	if !atLeastOneEnabled {
		return errNoProviders
	}

	service.methods = methodsList

	return nil
}

// validationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge using the Method.
func (method Method) validationResource(identifier acme.Identifier, key acme.AccountKey, token string) (name string, content string, err error) {
	return method.ChallengeType.ValidationResource(identifier, key, token)
}

// ListOfMethods() returns a slice of challenge methods as currently
// configured (i.e. with enabled/disabled status)
func (service *Service) ListOfMethods() (methods []Method) {
	return service.methods
}

// MethodByValue returns a challenge method based on its Value.
// If a method isn't found, UnknownMethod is returned.
func (service *Service) MethodByStorageValue(value MethodValue) Method {
	for i := range service.methods {
		if value == service.methods[i].Value {
			return service.methods[i]
		}
	}

	return UnknownMethod
}
