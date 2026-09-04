package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/helpers_test"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"fmt"
	"testing"
	"time"
)

func TestPostNewCert(t *testing.T) {
	newOrd36 := certificates.Certificate{
		ID:                 36,
		Name:               "NewCertHere",
		Description:        "some cert ins",
		Key:                key58,
		Account:            acmeAcct1,
		Subject:            "some.example.com",
		SubjectAltNames:    []string{"some1.example.com", "some2.example.com"},
		Organization:       "an org",
		OrganizationalUnit: "an ou",
		Country:            "usa",
		State:              "Ca",
		City:               "los santos",
		CSRExtraExtensions: []certificates.CertExtension{
			{
				Extension: pkix.Extension{
					Id:       asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24},
					Critical: false,
					Value:    []byte{0x30, 0x03, 0x02, 0x01, 0x05},
				},
				Description: "OCSP Must Staple",
			},
		},
		PreferredRootCN:             "Root xyz",
		LastAccess:                  time.Unix(0, 0),
		CreatedAt:                   time.Unix(770337479, 0),
		UpdatedAt:                   time.Unix(770338000, 0),
		ApiKey:                      "12345fffff",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                true,
		PostProcessingCommand:       "./run-me.py",
		PostProcessingEnvironment:   []string{"a=123", "b=456"},
		PostProcessingClientAddress: "endpoint.example.com",
		PostProcessingClientKeyB64:  "an aes key",
		Profile:                     "test-prof",
	}
	newOrd37 := certificates.Certificate{
		ID:                          37,
		Name:                        "NewCertHere2",
		Description:                 "some cert ins2",
		Key:                         key62,
		Account:                     acmeAcct1,
		Subject:                     "some.example2.com",
		SubjectAltNames:             []string{},
		Organization:                "an org2",
		OrganizationalUnit:          "an ou2",
		Country:                     "usa2",
		State:                       "Ca2",
		City:                        "los santos2",
		CSRExtraExtensions:          []certificates.CertExtension{},
		PreferredRootCN:             "Root xyz2",
		LastAccess:                  time.Unix(0, 0),
		CreatedAt:                   time.Unix(1234, 0),
		UpdatedAt:                   time.Unix(5678, 0),
		ApiKey:                      "12345fffff2",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                false,
		PostProcessingCommand:       "./run-me.py2",
		PostProcessingEnvironment:   []string{},
		PostProcessingClientAddress: "endpoint.example2.com",
		PostProcessingClientKeyB64:  "an aes ke2y",
		Profile:                     "test-prof2",
	}

	testCases := []struct {
		newPayload       certificates.NewPayload
		expectedPostCert *certificates.Certificate
		expectedPostErr  error

		getName         string
		expectedGetCert *certificates.Certificate
		expectedGetErr  error
	}{
		// valid insertions
		{
			certificates.NewPayload{
				Name:                 new("NewCertHere"),
				Description:          new("some cert ins"),
				PrivateKeyID:         new(58),
				NewKeyAlgorithmValue: new("some-alg"), // should be ignored
				AcmeAccountID:        new(1),
				Subject:              new("some.example.com"),
				SubjectAltNames:      []string{"some1.example.com", "some2.example.com"},
				Organization:         new("an org"),
				OrganizationalUnit:   new("an ou"),
				Country:              new("usa"),
				State:                new("Ca"),
				City:                 new("los santos"),
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24},
							Critical: false,
							Value:    []byte{0x30, 0x03, 0x02, 0x01, 0x05},
						},
						Description: "OCSP Must Staple",
					},
				},
				PreferredRootCN:             new("Root xyz"),
				PostProcessingCommand:       new("./run-me.py"),
				PostProcessingEnvironment:   []string{"a=123", "b=456"},
				PostProcessingClientAddress: new("endpoint.example.com"),
				PostProcessingClientKeyB64:  new("an aes key"),
				Profile:                     new("test-prof"),
				ApiKey:                      "12345fffff",
				ApiKeyViaUrl:                true,
				CreatedAt:                   time.Unix(770337479, 0),
				UpdatedAt:                   time.Unix(770338000, 0),
			},
			&newOrd36,
			nil,
			"NewCertHere",
			&newOrd36,
			nil,
		},
		// new with nil slices
		{
			certificates.NewPayload{
				Name:                        new("NewCertHere2"),
				Description:                 new("some cert ins2"),
				PrivateKeyID:                new(62),
				NewKeyAlgorithmValue:        new("some-alg-2"), // should be ignored
				AcmeAccountID:               new(1),
				Subject:                     new("some.example2.com"),
				SubjectAltNames:             nil,
				Organization:                new("an org2"),
				OrganizationalUnit:          new("an ou2"),
				Country:                     new("usa2"),
				State:                       new("Ca2"),
				City:                        new("los santos2"),
				CSRExtraExtensions:          nil,
				PreferredRootCN:             new("Root xyz2"),
				PostProcessingCommand:       new("./run-me.py2"),
				PostProcessingEnvironment:   nil,
				PostProcessingClientAddress: new("endpoint.example2.com"),
				PostProcessingClientKeyB64:  new("an aes ke2y"),
				Profile:                     new("test-prof2"),
				ApiKey:                      "12345fffff2",
				ApiKeyViaUrl:                false,
				CreatedAt:                   time.Unix(1234, 0),
				UpdatedAt:                   time.Unix(5678, 0),
			},
			&newOrd37,
			nil,
			"NewCertHere2",
			&newOrd37,
			nil,
		},
		// duplicate name (non-case sensitive)
		{
			certificates.NewPayload{
				Name:                        new("test008.TEST.example.com"),
				Description:                 new("some cert ins"),
				PrivateKeyID:                new(58),
				NewKeyAlgorithmValue:        new("some-alg"),
				AcmeAccountID:               new(1),
				Subject:                     new("some.example.com"),
				SubjectAltNames:             []string{},
				Organization:                new("an org"),
				OrganizationalUnit:          new("an ou"),
				Country:                     new("usa"),
				State:                       new("Ca"),
				City:                        new("los santos"),
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             new("Root xyz"),
				PostProcessingCommand:       new("./run-me.py"),
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: new("endpoint.example.com"),
				PostProcessingClientKeyB64:  new("an aes key"),
				Profile:                     new("test-prof"),
				ApiKey:                      "12345fffff",
				ApiKeyViaUrl:                true,
				CreatedAt:                   time.Unix(770337479, 0),
				UpdatedAt:                   time.Unix(770338000, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed: certificates.name"),
			"test008.TEST.example.com",
			&cert26,
			nil,
		},
		{ // incomplete payload 1
			certificates.NewPayload{
				Name:                 new("NewCertHerexxxxy"),
				Description:          new("some cert ins"),
				PrivateKeyID:         new(58),
				NewKeyAlgorithmValue: new("some-alg"), // should be ignored
				// AcmeAccountID:        new(1),
				Subject:            new("some.example.com"),
				SubjectAltNames:    []string{"some1.example.com", "some2.example.com"},
				Organization:       new("an org"),
				OrganizationalUnit: new("an ou"),
				Country:            new("usa"),
				State:              new("Ca"),
				City:               new("los santos"),
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24},
							Critical: false,
							Value:    []byte{0x30, 0x03, 0x02, 0x01, 0x05},
						},
						Description: "OCSP Must Staple",
					},
				},
				PreferredRootCN:             new("Root xyz"),
				PostProcessingCommand:       new("./run-me.py"),
				PostProcessingEnvironment:   []string{"a=123", "b=456"},
				PostProcessingClientAddress: new("endpoint.example.com"),
				PostProcessingClientKeyB64:  new("an aes key"),
				Profile:                     new("test-prof"),
				ApiKey:                      "12345fffff",
				ApiKeyViaUrl:                true,
				CreatedAt:                   time.Unix(770337479, 0),
				UpdatedAt:                   time.Unix(770338000, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed: certificates.acme_account_id"),
			"NewCertHerexxxxy",
			nil,
			sql.ErrNoRows,
		},
		{ // incomplete payload 2
			certificates.NewPayload{
				Name:                 new("NewCertHerexxxxyyyzzzz"),
				Description:          new("some cert ins"),
				PrivateKeyID:         new(58),
				NewKeyAlgorithmValue: new("some-alg"), // should be ignored
				AcmeAccountID:        new(1),
				// Subject:              new("some.example.com"),
				SubjectAltNames:    []string{"some1.example.com", "some2.example.com"},
				Organization:       new("an org"),
				OrganizationalUnit: new("an ou"),
				Country:            new("usa"),
				State:              new("Ca"),
				City:               new("los santos"),
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24},
							Critical: false,
							Value:    []byte{0x30, 0x03, 0x02, 0x01, 0x05},
						},
						Description: "OCSP Must Staple",
					},
				},
				PreferredRootCN:             new("Root xyz"),
				PostProcessingCommand:       new("./run-me.py"),
				PostProcessingEnvironment:   []string{"a=123", "b=456"},
				PostProcessingClientAddress: new("endpoint.example.com"),
				PostProcessingClientKeyB64:  new("an aes key"),
				Profile:                     new("test-prof"),
				ApiKey:                      "12345fffff",
				ApiKeyViaUrl:                true,
				CreatedAt:                   time.Unix(770337479, 0),
				UpdatedAt:                   time.Unix(770338000, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed: certificates.subject"),
			"NewCertHerexxxxyyyzzzz",
			nil,
			sql.ErrNoRows,
		},
	}

	// create testing service
	store := openStorageWithTestData(t, "postnewcert")

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", helpers_test.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			cert, err := store.PostNewCert(&tc.newPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			compareCertificate(t, cert, tc.expectedPostCert)

			cert, err = store.GetOneCertByName(tc.getName)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareCertificate(t, cert, tc.expectedGetCert)
		})
	}
}
