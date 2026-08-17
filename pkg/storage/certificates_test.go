package storage_test

import (
	"bytes"
	"certwarden-backend/pkg/domain/certificates"
	"slices"
	"testing"
)

// compareCertificateCSRExtensions is for comparing the special csr extra extensions
func compareCertificateCSRExtensions(t *testing.T, extns, expectedExtns []certificates.CertExtension) {
	if len(extns) != len(expectedExtns) {
		t.Errorf("certificate: csr extra extensions expected length '%d' but got '%d'", len(expectedExtns), len(extns))
	}

	// check all expected
	for i, extn := range expectedExtns {
		if i >= len(extns) {
			t.Errorf("certificate: csr extra extensions expected '%v' but got none", extn)
			continue
		}

		// check each field
		if !slices.Equal(extn.Id, extns[i].Id) {
			t.Errorf("certificate: csr extra extension %d expected oid '%v' but got '%v'", i, extn.Id, extns[i].Id)
		}

		if extn.Critical != extns[i].Critical {
			t.Errorf("certificate: csr extra extension %d expected critical '%t' but got '%t'", i, extn.Critical, extns[i].Critical)
		}

		if !bytes.Equal(extn.Value, extns[i].Value) {
			t.Errorf("certificate: csr extra extension %d expected value '%v' but got '%v'", i, extn.Value, extns[i].Value)
		}

		if extn.Description != extns[i].Description {
			t.Errorf("certificate: csr extra extension %d expected description '%s' but got '%s'", i, extn.Description, extns[i].Description)
		}
	}

	// error for any extras
	if len(extns) > len(expectedExtns) {
		for i := len(expectedExtns); i < len(extns); i++ {
			t.Errorf("certificate: csr extra extensions expected no additional but got '%v'", extns[i])
		}
	}
}

// compareCertificate compares cert to expectedCert and throws appropriate errors for any differences
func compareCertificate(t *testing.T, cert, expectedCert *certificates.Certificate) {
	if cert == nil && expectedCert == nil {
		return
	}
	if cert == nil && expectedCert != nil {
		t.Errorf("certificate: cert is nil but expectedCert is non-nil")
		return
	}
	if cert != nil && expectedCert == nil {
		t.Errorf("certificate: cert is non-nil but expectedCert is nil")
		return
	}

	if cert.ID != expectedCert.ID {
		t.Errorf("certificate: id expected '%d' but got '%d'", expectedCert.ID, cert.ID)
	}

	if cert.Name != expectedCert.Name {
		t.Errorf("certificate: name expected '%s' but got '%s'", expectedCert.Name, cert.Name)
	}

	if cert.Description != expectedCert.Description {
		t.Errorf("certificate: description expected '%s' but got '%s'", expectedCert.Description, cert.Description)
	}

	compareKey(t, &cert.Key, &expectedCert.Key)

	compareAcmeAccount(t, &cert.Account, &expectedCert.Account)

	if cert.Subject != expectedCert.Subject {
		t.Errorf("certificate: subject expected '%s' but got '%s'", expectedCert.Subject, cert.Subject)
	}

	if !slices.Equal(cert.SubjectAltNames, expectedCert.SubjectAltNames) {
		t.Errorf("certificate: subject alt names expected '%v' but got '%v'", expectedCert.SubjectAltNames, cert.SubjectAltNames)
	}

	if cert.Organization != expectedCert.Organization {
		t.Errorf("certificate: organization expected '%s' but got '%s'", expectedCert.Organization, cert.Organization)
	}

	if cert.OrganizationalUnit != expectedCert.OrganizationalUnit {
		t.Errorf("certificate: organizational unit expected '%s' but got '%s'", expectedCert.OrganizationalUnit, cert.OrganizationalUnit)
	}

	if cert.Country != expectedCert.Country {
		t.Errorf("certificate: country expected '%s' but got '%s'", expectedCert.Country, cert.Country)
	}

	if cert.State != expectedCert.State {
		t.Errorf("certificate: state expected '%s' but got '%s'", expectedCert.State, cert.State)
	}

	if cert.City != expectedCert.City {
		t.Errorf("certificate: city expected '%s' but got '%s'", expectedCert.City, cert.City)
	}

	compareCertificateCSRExtensions(t, cert.CSRExtraExtensions, expectedCert.CSRExtraExtensions)

	if cert.PreferredRootCN != expectedCert.PreferredRootCN {
		t.Errorf("certificate: preferred root cn expected '%s' but got '%s'", expectedCert.PreferredRootCN, cert.PreferredRootCN)
	}

	if !cert.LastAccess.Equal(expectedCert.LastAccess) {
		t.Errorf("certificate: last access expected '%s' but got '%s'", expectedCert.LastAccess.UTC(), cert.LastAccess.UTC())
	}

	if !cert.CreatedAt.Equal(expectedCert.CreatedAt) {
		t.Errorf("certificater: created at expected '%s' but got '%s'", expectedCert.CreatedAt.UTC(), cert.CreatedAt.UTC())
	}

	if !cert.UpdatedAt.Equal(expectedCert.UpdatedAt) {
		t.Errorf("certificate: updated at expected '%s' but got '%s'", expectedCert.UpdatedAt.UTC(), cert.UpdatedAt.UTC())
	}

	if cert.ApiKey != expectedCert.ApiKey {
		t.Errorf("certificate: api key expected '%s' but got '%s'", expectedCert.ApiKey, cert.ApiKey)
	}

	if cert.ApiKeyNew != expectedCert.ApiKeyNew {
		t.Errorf("certificate: api key new expected '%s' but got '%s'", expectedCert.ApiKeyNew, cert.ApiKeyNew)
	}

	if cert.ApiKeyViaUrl != expectedCert.ApiKeyViaUrl {
		t.Errorf("acme server: api key via url expected '%t' but got '%t'", expectedCert.ApiKeyViaUrl, cert.ApiKeyViaUrl)
	}

	if cert.PostProcessingCommand != expectedCert.PostProcessingCommand {
		t.Errorf("acme server: post processing command expected '%s' but got '%s'", expectedCert.PostProcessingCommand, cert.PostProcessingCommand)
	}

	if !slices.Equal(cert.PostProcessingEnvironment, expectedCert.PostProcessingEnvironment) {
		t.Errorf("certificate: post processing environment expected '%v' but got '%v'", expectedCert.PostProcessingEnvironment, cert.PostProcessingEnvironment)
	}

	if cert.PostProcessingClientAddress != expectedCert.PostProcessingClientAddress {
		t.Errorf("acme server: post processing client address expected '%s' but got '%s'", expectedCert.PostProcessingClientAddress, cert.PostProcessingClientAddress)
	}

	if cert.PostProcessingClientKeyB64 != expectedCert.PostProcessingClientKeyB64 {
		t.Errorf("acme server: post processing client key base64 expected '%s' but got '%s'", expectedCert.PostProcessingClientKeyB64, cert.PostProcessingClientKeyB64)
	}

	if cert.Profile != expectedCert.Profile {
		t.Errorf("acme server: profile expected '%s' but got '%s'", expectedCert.Profile, cert.Profile)
	}
}
