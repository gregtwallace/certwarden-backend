package challenges

import "legocerthub-backend/pkg/acme"

// Define MethodDetails which contains details about the defined
// methods.
type challMethodDetails struct {
	method        Method
	storageValue  string
	name          string
	challengeType acme.ChallengeType
}

var methodDetails = []challMethodDetails{
	{
		// serve the http record from an internal http server
		method:        http01Internal,
		storageValue:  "http-01-internal",
		name:          "HTTP (Self Served)",
		challengeType: acme.ChallengeTypeHttp01,
	},
	{
		// create and delete dns records on Cloudflare
		method:        dns01Cloudflare,
		storageValue:  "dns-01-cloudflare",
		name:          "DNS-01 (Cloudflare)",
		challengeType: acme.ChallengeTypeDns01,
	},
}

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
func MethodByStorageValue(value string) Method {
	for i := range methodDetails {
		if value == methodDetails[i].storageValue {
			return methodDetails[i].method
		}
	}

	return UnknownMethod
}
