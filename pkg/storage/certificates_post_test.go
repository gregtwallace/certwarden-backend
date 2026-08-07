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
	testCases := []struct {
		newPayload      certificates.NewPayload
		expectedPostErr error
		expectedNew     certificates.Certificate
		expectedGetErr  error
	}{
		{ // valid insertion
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
			nil,
			certificates.Certificate{
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
			},
			nil,
		},
		{ // duplicate name (non-case sensitive)
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
			helpers_test.MakeTestErrorStringComp("UNIQUE constraint failed: certificates.name"),
			certificates.Certificate{},
			sql.ErrNoRows,
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
			helpers_test.MakeTestErrorStringComp("NOT NULL constraint failed: certificates.acme_account_id"),
			certificates.Certificate{},
			sql.ErrNoRows,
		},
		{ // incomplete payload 2
			certificates.NewPayload{
				Name:                 new("NewCertHerexxxxyyy"),
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
			helpers_test.MakeTestErrorStringComp("NOT NULL constraint failed: certificates.subject"),
			certificates.Certificate{},
			sql.ErrNoRows,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "postnewcert")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", helpers_test.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			record, err := storage.PostNewCert(tc.newPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, record, tc.expectedNew)

			record, err = storage.GetOneCertByName(record.Name)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, record, tc.expectedNew)
		})
	}
}
