package challenges

import "errors"

var errUnsupportedChallengeType = errors.New("unsupported challenge type")

// ChallengeType is a type to hold challenge types
type ChallengeType struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// ListOfChallengeTypes() returns a constant list of challenge types
// The Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfChallengeTypes() []ChallengeType {
	return []ChallengeType{
		{
			Value: "http-01",
			Name:  "HTTP-01",
		},
		{
			Value: "dns-01-cloudflare",
			Name:  "DNS-01 (Cloudflare)",
		},
	}
}

// ChallengeTypeByValue returns a challenge type based on its Value
// Returns an error if the challenge type is not supported
func ChallengeTypeByValue(value string) (ChallengeType, error) {
	// TODO: Rework using range
	challengeTypes := ListOfChallengeTypes()

	for i := 0; i < len(challengeTypes); i++ {
		if value == challengeTypes[i].Value {
			return challengeTypes[i], nil
		}
	}

	return ChallengeType{}, errUnsupportedChallengeType
}
