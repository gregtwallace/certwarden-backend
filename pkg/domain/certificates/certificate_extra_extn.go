package certificates

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// CertExtension us a pkix.Extension with an additional field for
// a description
type CertExtension struct {
	pkix.Extension
	Description string
}

// String prints a log friendly version of the Certificate Extension (useful for testing)
func (cext CertExtension) String() string {
	return fmt.Sprintf("CertExtension{Description: %s, Id: %s, Critical: %t, Value: %x}", cext.Description, cext.Id.String(), cext.Critical, cext.Value)
}

// UnmarshalJSON implements the unmarshalling interface for CertExtension
// json should be in the shape:
//
//	{
//		Description    string `json:"description"`
//		OID            string `json:"oid"`
//		Critical       bool   `json:"critical"`
//		ValueHexString string `json:"value_hex"`
//	}
func (cext *CertExtension) UnmarshalJSON(b []byte) error {
	// generic unmarshal
	m := map[string]interface{}{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	result := CertExtension{}

	// description
	desc, ok := m["description"]
	if !ok {
		return errors.New("'description' missing")
	}
	result.Description, ok = desc.(string)
	if !ok {
		return errors.New("'description' is not string type")
	}

	// oid - must convert into asn1.ObjectIdentifier (must be in dot notation)
	desc, ok = m["oid"]
	if !ok {
		return errors.New("'oid' missing")
	}
	oidStr, ok := desc.(string)
	if !ok {
		return errors.New("'oid' is not string type")
	}

	oidParts := strings.Split(oidStr, ".")
	id := make(asn1.ObjectIdentifier, len(oidParts))
	for i := range oidParts {
		id[i], err = strconv.Atoi(oidParts[i])
		if err != nil {
			return errors.New("invalid oid format")
		}
	}
	result.Id = id

	// critical
	desc, ok = m["critical"]
	if !ok {
		return errors.New("'critical' missing")
	}
	result.Critical, ok = desc.(bool)
	if !ok {
		return errors.New("'critical' is not bool type")
	}

	// value - must convert from hex string to []byte
	// allow no delimiter, ':' (colon), or ' ' (space) as delimiter
	desc, ok = m["value_hex"]
	if !ok {
		return errors.New("'value_hex' missing")
	}
	valueHex, ok := desc.(string)
	if !ok {
		return errors.New("'value_hex' is not string type")
	}

	valueHexParts := []string{}
	// check for and deal with ':' or ' ' separation
	if strings.Contains(valueHex, ":") {
		// has colons
		valueHexParts = strings.Split(valueHex, ":")
	} else if strings.Contains(valueHex, " ") {
		// has spaces
		valueHexParts = strings.Split(valueHex, " ")
	}

	// if we made value parts, build hex without separator string from them, if we did
	// not, use original hex value
	valueHexNoSep := ""
	if len(valueHexParts) > 0 {
		for i := range valueHexParts {
			// each byte must be explicityly two chars long
			if len(valueHexParts[i]) != 2 {
				// fail if not
				return fmt.Errorf("invalid value byte '%s' (not 2 chars)", valueHexParts[i])
			}

			// add byte to the no seperation string
			valueHexNoSep += valueHexParts[i]
		}
	} else {
		// no separator was found, use as-is
		valueHexNoSep = valueHex
	}

	// decode hex string
	result.Value, err = hex.DecodeString(valueHexNoSep)
	if err != nil {
		return fmt.Errorf("failed to decode hex '%s'", valueHexNoSep)
	}

	// good to go
	*cext = result
	return nil
}

// MarshalJSON implements the marshalling interface for CertExtension
// and the output json will be in the shape:
//
//	{
//		Description    string `json:"description"`
//		OID            string `json:"oid"`
//		Critical       bool   `json:"critical"`
//		ValueHexString string `json:"value_hex"`
//	}
func (cext *CertExtension) MarshalJSON() ([]byte, error) {
	out := struct {
		Description    string `json:"description"`
		OID            string `json:"oid"`
		Critical       bool   `json:"critical"`
		ValueHexString string `json:"value_hex"`
	}{
		Description:    cext.Description,
		OID:            cext.Id.String(),
		Critical:       cext.Critical,
		ValueHexString: hex.EncodeToString(cext.Value),
	}

	return json.Marshal(out)
}
