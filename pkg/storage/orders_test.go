package storage_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"slices"
	"testing"
)

// compareACMERenewalInfo is for comparing order renewal info structs
func compareACMERenewalInfo(t *testing.T, ari, expectedARI *orders.RenewalInfo) {
	if ari == nil && expectedARI == nil {
		return
	}
	if ari == nil && expectedARI != nil {
		t.Errorf("order: ari is nil but expectedARI is non-nil")
		return
	}
	if ari != nil && expectedARI == nil {
		t.Errorf("order: ari is non-nil but expectedARI is nil")
		return
	}

	if !ari.SuggestedWindow.Start.Equal(expectedARI.SuggestedWindow.Start) {
		t.Errorf("order: ari window start expected '%s' but got '%s'", expectedARI.SuggestedWindow.Start, ari.SuggestedWindow.Start)
	}

	if !ari.SuggestedWindow.End.Equal(expectedARI.SuggestedWindow.End) {
		t.Errorf("order: ari window end expected '%s' but got '%s'", expectedARI.SuggestedWindow.End, ari.SuggestedWindow.End)
	}

	err := helpers_test.StringPointerEquals(ari.ExplanationURL, expectedARI.ExplanationURL)
	if err != nil {
		t.Errorf("order: ari explanation url expected '%s' but got '%s'", helpers_test.StringPointerToVal(expectedARI.ExplanationURL), helpers_test.StringPointerToVal(ari.ExplanationURL))
	}

	err = helpers_test.TimePointerEquals(ari.RetryAfter, expectedARI.RetryAfter)
	if err != nil {
		t.Errorf("order: ari retry after expected '%s' but got '%s'", helpers_test.TimeToVal(expectedARI.RetryAfter), helpers_test.TimeToVal(ari.RetryAfter))
	}
}

// compareACMEError is for comparing acme.Error structs
func compareACMEError(t *testing.T, err, expectedErr *acme.Error) {
	if err == nil && expectedErr == nil {
		return
	}
	if err == nil && expectedErr != nil {
		t.Errorf("order: acmeerr is nil but expectedAcmeErr is non-nil")
		return
	}
	if err != nil && expectedErr == nil {
		t.Errorf("order: acmeerr is non-nil but expectedAcmeErr is nil")
		return
	}

	if err.Status != expectedErr.Status {
		t.Errorf("order: acmeerr status expected '%d' but got '%d'", err.Status, expectedErr.Status)
	}

	if err.Type != expectedErr.Type {
		t.Errorf("order: acmeerr type expected '%s' but got '%s'", err.Type, expectedErr.Type)
	}

	if err.Detail != expectedErr.Detail {
		t.Errorf("order: acmeerr detail expected '%s' but got '%s'", err.Detail, expectedErr.Detail)
	}
}

// compareOrder compares ord to expectedOrd and throws appropriate errors for any differences
func compareOrder(t *testing.T, ord, expectedOrd *orders.Order) {
	if ord == nil && expectedOrd == nil {
		return
	}
	if ord == nil && expectedOrd != nil {
		t.Errorf("order: ord is nil but expectedOrd is non-nil")
		return
	}
	if ord != nil && expectedOrd == nil {
		t.Errorf("order: ord is non-nil but expectedOrd is nil")
		return
	}

	if ord.ID != expectedOrd.ID {
		t.Errorf("order: id expected '%d' but got '%d'", expectedOrd.ID, ord.ID)
	}

	compareCertificate(t, &ord.Certificate, &expectedOrd.Certificate)

	if ord.Location != expectedOrd.Location {
		t.Errorf("order: location expected '%s' but got '%s'", expectedOrd.Location, ord.Location)
	}

	if ord.Status != expectedOrd.Status {
		t.Errorf("order: status expected '%s' but got '%s'", expectedOrd.Status, ord.Status)
	}

	if ord.KnownRevoked != expectedOrd.KnownRevoked {
		t.Errorf("order: knownrevoked expected '%t' but got '%t'", expectedOrd.KnownRevoked, ord.KnownRevoked)
	}

	compareACMEError(t, ord.Error, expectedOrd.Error)

	err := helpers_test.TimePointerEquals(ord.Expires, expectedOrd.Expires)
	if err != nil {
		t.Errorf("order: expires '%s'", err)
	}

	if !slices.Equal(ord.DnsIdentifiers, expectedOrd.DnsIdentifiers) {
		t.Errorf("order: dnsidentifiers expected '%s' but got '%s'", expectedOrd.DnsIdentifiers, ord.DnsIdentifiers)
	}

	if !slices.Equal(ord.Authorizations, expectedOrd.Authorizations) {
		t.Errorf("order: authorizations expected '%s' but got '%s'", expectedOrd.Authorizations, ord.Authorizations)
	}

	if ord.Finalize != expectedOrd.Finalize {
		t.Errorf("order: finalize expected '%s' but got '%s'", expectedOrd.Finalize, ord.Finalize)
	}

	compareKey(t, ord.FinalizedKey, expectedOrd.FinalizedKey)

	err = helpers_test.StringPointerEquals(ord.CertificateUrl, expectedOrd.CertificateUrl)
	if err != nil {
		t.Errorf("order: certificateurl expected '%s' but got '%s'", helpers_test.StringPointerToVal(expectedOrd.CertificateUrl), helpers_test.StringPointerToVal(ord.CertificateUrl))
	}

	err = helpers_test.StringPointerEquals(ord.Pem, expectedOrd.Pem)
	if err != nil {
		t.Errorf("order: pem expected '%s' but got '%s'", helpers_test.StringPointerToVal(expectedOrd.Pem), helpers_test.StringPointerToVal(ord.Pem))
	}

	err = helpers_test.TimePointerEquals(ord.ValidFrom, expectedOrd.ValidFrom)
	if err != nil {
		t.Errorf("order: validfrom '%s'", err)
	}

	err = helpers_test.TimePointerEquals(ord.ValidTo, expectedOrd.ValidTo)
	if err != nil {
		t.Errorf("order: validto '%s'", err)
	}

	err = helpers_test.StringPointerEquals(ord.ChainRootCN, expectedOrd.ChainRootCN)
	if err != nil {
		t.Errorf("order: chainrootcn expected '%s' but got '%s'", helpers_test.StringPointerToVal(expectedOrd.ChainRootCN), helpers_test.StringPointerToVal(ord.ChainRootCN))
	}

	if !ord.CreatedAt.Equal(expectedOrd.CreatedAt) {
		t.Errorf("order: createdat expected '%s' but got '%s'", expectedOrd.CreatedAt.UTC(), ord.CreatedAt.UTC())
	}

	if !ord.UpdatedAt.Equal(expectedOrd.UpdatedAt) {
		t.Errorf("order: updatedat expected '%s' but got '%s'", expectedOrd.UpdatedAt.UTC(), ord.UpdatedAt.UTC())
	}

	err = helpers_test.StringPointerEquals(ord.Profile, expectedOrd.Profile)
	if err != nil {
		t.Errorf("order: profile expected '%s' but got '%s'", helpers_test.StringPointerToVal(expectedOrd.Profile), helpers_test.StringPointerToVal(ord.Profile))
	}

	compareACMERenewalInfo(t, ord.RenewalInfo, expectedOrd.RenewalInfo)
}
