package challenges

import "errors"

var errUnsupportedMethod = errors.New("unsupported challenge method")

// Method is a struct to hold various challenge methods.
// This is not "challenge type" as the spec is specifc to types and this app
// is more general.
// In particular, multiple DNS providers may be integrated in addition
// to a generic DNS option that relies on an external script.
type Method struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// ListOfMethods() returns a constant list of challenge methods
// The Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfMethods() []Method {
	return []Method{
		{
			// serve the http record from this server
			Value: "http-01-internal",
			Name:  "HTTP (Self Served)",
			Type:  "http-01",
		},
		// TODO: Implement DNS
		// {
		// 	// call external scripts to create and delete dns records
		// 	Value: "dns-01-script",
		// 	Name:  "DNS-01 (Manual)",
		// 	Type:  "dns-01",
		// },
	}
}

// MethodByValue returns a challenge method based on its Value
// Returns an error if the challenge method is not supported
func MethodByValue(value string) (Method, error) {
	allMethods := ListOfMethods()

	for i := range allMethods {
		if value == allMethods[i].Value {
			return allMethods[i], nil
		}
	}

	return Method{}, errUnsupportedMethod
}
