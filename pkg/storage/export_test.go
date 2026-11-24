package storage

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
	"time"
)

// errors.go
var ErrorWrongRowCount = errorWrongRowCount
var ErrorWrongAffectedRowCount = errorWrongAffectedRowCount

// time.go
var TimeNow = func() func() time.Time { return timeNow } // returns current timeNow function
var SetTimeNow = setTimeNow

// types_json.go
var SliceToJsonString_Strings = sliceToJsonString[[]string]
var SliceToJsonString_CertExtensions = sliceToJsonString[[]certificates.CertExtension]
var StructToNullableJsonString_acmeError = structToNullableJsonString[acme.Error]
var JsonStringToNullableStruct_acmeError = jsonStringToNullableStruct[acme.Error]

// types_null.go
var NullInt64UnixToTime = nullInt64UnixToTime
var TimePointerToNullInt64 = timePointerToNullInt64
var NullStringToString = nullStringToString
