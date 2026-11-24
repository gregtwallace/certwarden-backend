package storage_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/storage"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
)

func TestSliceToJsonString_Strings(t *testing.T) {
	// string slice
	tc := []struct {
		slice          []string
		nullable       bool
		expectedResult *string
	}{
		{
			[]string{"a", "b", "cdef12345"},
			false,
			new(`["a","b","cdef12345"]`),
		},
		{
			[]string{"a", "b", "cdef12345"},
			true,
			new(`["a","b","cdef12345"]`),
		},
		{
			[]string{},
			false,
			new(`[]`),
		},
		{
			[]string{},
			true,
			new(`[]`),
		},
		{
			nil,
			false,
			new(`[]`),
		},
		{
			nil,
			true,
			nil,
		},
	}

	for i := range tc {
		t.Run("%d", func(t *testing.T) {
			result, err := storage.SliceToJsonString_Strings(tc[i].slice, tc[i].nullable)
			if err != nil {
				t.Errorf("error: %s", err)
				return
			}

			if result == nil && tc[i].expectedResult != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedResult == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedResult == nil {
				return
			}

			if *result != *tc[i].expectedResult {
				t.Errorf("result expected '%s' but got '%s'", *tc[i].expectedResult, *result)
			}
		})
	}
}

func TestSliceToJsonString_CertExtensions(t *testing.T) {
	// data
	sliceOne := []certificates.CertExtension{
		{
			Extension: pkix.Extension{
				Id:       asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24},
				Critical: false,
				Value:    []byte{0x30, 0x03, 0x02, 0x01, 0x05},
			},
			Description: "OCSP Must Staple",
		},
		{
			Extension: pkix.Extension{
				Id:       asn1.ObjectIdentifier{4, 6, 5},
				Critical: true,
				Value:    []byte{0x0a, 0x0b},
			},
			Description: "another",
		},
	}
	resultOne := `[{"description":"OCSP Must Staple","oid":"1.3.6.1.5.5.7.1.24","critical":false,"value_hex":"3003020105"},{"description":"another","oid":"4.6.5","critical":true,"value_hex":"0a0b"}]`

	// string slice
	tc := []struct {
		slice          []certificates.CertExtension
		nullable       bool
		expectedResult *string
	}{
		{
			sliceOne,
			false,
			new(resultOne),
		},
		{
			sliceOne,
			true,
			new(resultOne),
		},
		{
			[]certificates.CertExtension{},
			false,
			new(`[]`),
		},
		{
			[]certificates.CertExtension{},
			true,
			new(`[]`),
		},
		{
			nil,
			false,
			new(`[]`),
		},
		{
			nil,
			true,
			nil,
		},
	}

	for i := range tc {
		t.Run("%d", func(t *testing.T) {
			result, err := storage.SliceToJsonString_CertExtensions(tc[i].slice, tc[i].nullable)
			if err != nil {
				t.Errorf("error: %s", err)
				return
			}

			if result == nil && tc[i].expectedResult != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedResult == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedResult == nil {
				return
			}

			if *result != *tc[i].expectedResult {
				t.Errorf("result expected '%s' but got '%s'", *tc[i].expectedResult, *result)
			}
		})
	}
}

func TestStructToNullableJsonString_acmeError(t *testing.T) {
	// string slice
	tc := []struct {
		struc          *acme.Error
		expectedResult *string
	}{
		{
			new(acme.Error{
				Status: 341,
				Type:   "urn:ietf:params:acme:error:someThing1",
				Detail: "whoops",
			}),
			new(`{"status":341,"type":"urn:ietf:params:acme:error:someThing1","detail":"whoops"}`),
		},
		{
			nil,
			nil,
		},
	}

	for i := range tc {
		t.Run("%d", func(t *testing.T) {
			result, err := storage.StructToNullableJsonString_acmeError(tc[i].struc)
			if err != nil {
				t.Errorf("error: %s", err)
				return
			}

			if result == nil && tc[i].expectedResult != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedResult == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedResult == nil {
				return
			}

			if *result != *tc[i].expectedResult {
				t.Errorf("result expected '%s' but got '%s'", *tc[i].expectedResult, *result)
			}
		})
	}
}

func TestJsonStringToNullableStruct_acmeError(t *testing.T) {
	// string slice
	tc := []struct {
		str            *string
		expectedResult *acme.Error
	}{
		{
			new(`{"status":341,"type":"urn:ietf:params:acme:error:someThing1","detail":"whoops"}`),
			new(acme.Error{
				Status: 341,
				Type:   "urn:ietf:params:acme:error:someThing1",
				Detail: "whoops",
			}),
		},
		{
			nil,
			nil,
		},
	}

	for i := range tc {
		t.Run("%d", func(t *testing.T) {
			result, err := storage.JsonStringToNullableStruct_acmeError(tc[i].str)
			if err != nil {
				t.Errorf("error: %s", err)
				return
			}

			if result == nil && tc[i].expectedResult != nil {
				t.Error("result is nil but expected result is non-nil")
				return
			}

			if result != nil && tc[i].expectedResult == nil {
				t.Error("result is non-nil but expected result is nil")
				return
			}

			if result == nil && tc[i].expectedResult == nil {
				return
			}

			// check each value
			if result.Detail != tc[i].expectedResult.Detail {
				t.Errorf("result expected detail '%s' but got '%s'", tc[i].expectedResult.Detail, result.Detail)
			}

			if result.Type != tc[i].expectedResult.Type {
				t.Errorf("result expected tupe '%s' but got '%s'", tc[i].expectedResult.Type, result.Type)
			}

			if result.Status != tc[i].expectedResult.Status {
				t.Errorf("result expected status '%d' but got '%d'", tc[i].expectedResult.Status, result.Status)
			}
		})
	}
}
