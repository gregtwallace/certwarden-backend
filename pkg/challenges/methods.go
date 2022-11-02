package challenges

import "legocerthub-backend/pkg/acme"

// Define MethodDetails which contains details about the defined
// methods.
type challMethodDetails struct {
	method        Method
	value         string
	name          string
	challengeType acme.ChallengeType
}

var methodDetails = []challMethodDetails{
	{
		// serve the http record from an internal http server
		method:        http01Internal,
		value:         "http-01-internal",
		name:          "HTTP (Self Served)",
		challengeType: acme.ChallengeTypeHttp01,
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
		if value == methodDetails[i].value {
			return methodDetails[i].method
		}
	}

	return UnknownMethod
}
