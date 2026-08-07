package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"fmt"
	"testing"
	"time"
)

func TestPutCertLastAccess(t *testing.T) {
	testCases := []struct {
		certId     int
		lastAccess time.Time

		expectedCert   certificates.Certificate
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id
			-1,
			time.Unix(88888111, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			time.Unix(88888222, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		// do update
		{
			18,
			time.Unix(0, 0),
			certificates.Certificate{
				ID:                 18,
				Name:               "serverdefault",
				Description:        "its a decript",
				Key:                key31,
				Account:            acmeAcct2,
				Subject:            "desk.dude.example.com",
				SubjectAltNames:    []string{"test011.test.example.com"},
				Organization:       "my org",
				OrganizationalUnit: "my ou",
				Country:            "your country",
				State:              "a state",
				City:               "springfield",
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
				PreferredRootCN:             "ISRG Root X1",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(1779386440, 0),
				ApiKey:                      "api-secret-18",
				ApiKeyNew:                   "api-new-secret-18",
				ApiKeyViaUrl:                true,
				PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
				PostProcessingEnvironment:   []string{"asdasdasdsd=asasd"},
				PostProcessingClientAddress: "dude.greg.example.com",
				PostProcessingClientKeyB64:  "aaaaaaaaaaaaaaaaaaaaaaaaaaa-ccccccccccccccc",
				Profile:                     "tlsserver",
			},
			nil,
			nil,
		},
		{
			18,
			time.Unix(1122885, 0),
			certificates.Certificate{
				ID:                 18,
				Name:               "serverdefault",
				Description:        "its a decript",
				Key:                key31,
				Account:            acmeAcct2,
				Subject:            "desk.dude.example.com",
				SubjectAltNames:    []string{"test011.test.example.com"},
				Organization:       "my org",
				OrganizationalUnit: "my ou",
				Country:            "your country",
				State:              "a state",
				City:               "springfield",
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
				PreferredRootCN:             "ISRG Root X1",
				LastAccess:                  time.Unix(1122885, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(1779386440, 0),
				ApiKey:                      "api-secret-18",
				ApiKeyNew:                   "api-new-secret-18",
				ApiKeyViaUrl:                true,
				PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
				PostProcessingEnvironment:   []string{"asdasdasdsd=asasd"},
				PostProcessingClientAddress: "dude.greg.example.com",
				PostProcessingClientKeyB64:  "aaaaaaaaaaaaaaaaaaaaaaaaaaa-ccccccccccccccc",
				Profile:                     "tlsserver",
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putcertlastaccess")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.certId), func(t *testing.T) {
			err := storage.PutCertLastAccess(tc.certId, tc.lastAccess)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			record, err := storage.GetOneCertById(tc.certId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, record, tc.expectedCert)
		})
	}
}
