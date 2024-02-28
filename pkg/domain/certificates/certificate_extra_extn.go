package certificates

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var (
	errCertExtOIDBadFormat = errors.New("certificates extension: OID string invalid (must be in dot notation)")
	errCertExtValueBad     = errors.New("certificates extension: Value invalid (must be hex string, hex string with colons, or hex string with spaces)")
)

// CertExtension us a pkix.Extension with an additional field for
// a description
type CertExtension struct {
	pkix.Extension
	Description string
}

// CertExtensionJSON is the object to use in the API (both input and output)
// to represent the custom CertificateExtension
type CertExtensionJSON struct {
	Description    string `json:"description"`
	OID            string `json:"oid"`
	Critical       bool   `json:"critical"`
	ValueHexString string `json:"value_hex"`
}

// toJSONObj returns the JSON object of the custom CertExtension
func (ce CertExtension) toJSONObj() CertExtensionJSON {
	return CertExtensionJSON{
		Description:    ce.Description,
		OID:            ce.Id.String(),
		Critical:       ce.Critical,
		ValueHexString: hex.EncodeToString(ce.Value),
	}
}

// toCertExtension validates the CertExtensionJSON and then returns
// the CertExtension object; if any fields fail to validate, an error
// is returned instead
func (cej CertExtensionJSON) ToCertExtension() (CertExtension, error) {
	ce := CertExtension{}

	// Description - no validation needed
	ce.Description = cej.Description

	// OID - must convert into asn1.ObjectIdentifier (must be in dot notation)
	var err error
	oidParts := strings.Split(cej.OID, ".")
	id := make(asn1.ObjectIdentifier, len(oidParts))
	for i := range oidParts {
		id[i], err = strconv.Atoi(oidParts[i])
		if err != nil {
			return CertExtension{}, errCertExtOIDBadFormat
		}
	}
	ce.Id = id

	// Cricial - no validation needed
	ce.Critical = cej.Critical

	// ValueByteString - must be a valid hex byte string; will try a couple of parsing
	// options (allow bytes to be separated by colons or spaces)
	valueParts := []string{}
	if strings.Contains(cej.ValueHexString, ":") {
		// has colons
		valueParts = strings.Split(cej.ValueHexString, ":")
	} else if strings.Contains(cej.ValueHexString, " ") {
		// has spaces
		valueParts = strings.Split(cej.ValueHexString, " ")
	}

	// if we made value parts, build hex without separator string from them, if we did
	// not, use original hex value
	valueHexNoSep := ""
	if len(valueParts) > 0 {
		for i := range valueParts {
			// each byte must be explicityly two chars long
			if len(valueParts[i]) != 2 {
				// fail if not
				return CertExtension{}, errCertExtValueBad
			}

			// add byte to the no seperation string
			valueHexNoSep += valueParts[i]
		}
	} else {
		// no separator was found, use as-is
		valueHexNoSep = cej.ValueHexString
	}

	// decode hex string
	ce.Value, err = hex.DecodeString(valueHexNoSep)
	if err != nil {
		return CertExtension{}, errCertExtValueBad
	}

	return ce, nil
}
