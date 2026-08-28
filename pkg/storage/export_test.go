package storage

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
)

// errors.go
var ErrorWrongUpdateRowCount = errorWrongUpdateRowCount

// types_json.go
var SliceToJsonString_Strings = sliceToJsonString[[]string]
var SliceToJsonString_CertExtensions = sliceToJsonString[[]certificates.CertExtension]
var StructToNullableJsonString_acmeError = structToNullableJsonString[acme.Error]
var JsonStringToNullableStruct_acmeError = jsonStringToNullableStruct[acme.Error]

// types_null.go
var NullInt64UnixToTime = nullInt64UnixToTime
var TimePointerToNullInt64 = timePointerToNullInt64
var NullStringToString = nullStringToString
