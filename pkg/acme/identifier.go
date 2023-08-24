package acme

// Identifier is the ACME Identifier object
type Identifier struct {
	Type  IdentifierType `json:"type"`
	Value string         `json:"value"`
}

// Define ACME identifier types (per RFC 8555 9.7.7)
type IdentifierType string

const (
	UnknownIdentifierType IdentifierType = ""

	IdentifierTypeDns = "dns"
)

// IdentifierSlice is a slice of Identifier
type IdentifierSlice []Identifier

// DnsIdentifiers returns a slice of the value strings for a slice of Identifiers
func (ids *IdentifierSlice) DnsIdentifiers() []string {
	var s []string

	for _, id := range *ids {
		if id.Type == IdentifierTypeDns {
			s = append(s, id.Value)
		}
	}

	return s
}
