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

func TestPutDetailsCert(t *testing.T) {
	testCases := []struct {
		payload certificates.UpdatePayload

		expectedPutResult certificates.Certificate
		expectedPutErr    error

		getId             int
		expectedGetResult certificates.Certificate
		expectedGetErr    error
	}{
		{ // invalid cert
			certificates.UpdatePayload{
				ID: -1,
			},
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			-1,
			certificates.Certificate{},
			sql.ErrNoRows,
		},
		{ // invalid key
			certificates.UpdatePayload{
				ID: 522,
			},
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			522,
			certificates.Certificate{},
			sql.ErrNoRows,
		},
		{ // update all things
			certificates.UpdatePayload{
				ID:                 18,
				Name:               new("somenewNameHere"),
				Description:        new("some new desc goes here"),
				PrivateKeyId:       new(58),
				SubjectAltNames:    []string{"new.example.com"},
				Organization:       new("new org 1"),
				OrganizationalUnit: new("new orgu 1"),
				Country:            new("new orgco 1"),
				State:              new("new orgst 1"),
				City:               new("new orgci 1"),
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 2, 4},
							Critical: true,
							Value:    []byte{0xaa, 0xbb, 0x11},
						},
						Description: "a",
					},
				},
				PreferredRootCN:             new("different new cn"),
				PostProcessingCommand:       new("./app.exe"),
				PostProcessingEnvironment:   []string{"a=123", "b=zba"},
				PostProcessingClientAddress: new("xyz.com"),
				PostProcessingClientKeyB64:  new("aaa888aaabbbccc"),
				Profile:                     new("new prof 2"),
				ApiKey:                      new("api-key---"),
				ApiKeyNew:                   new("api-key-new---"),
				ApiKeyViaUrl:                new(false),
				UpdatedAt:                   time.Unix(222223333, 0),
			},
			certificates.Certificate{
				ID:                 18,
				Name:               "somenewNameHere",
				Description:        "some new desc goes here",
				Key:                key58,
				Account:            acmeAcct2,
				Subject:            "desk.dude.example.com",
				SubjectAltNames:    []string{"new.example.com"},
				Organization:       "new org 1",
				OrganizationalUnit: "new orgu 1",
				Country:            "new orgco 1",
				State:              "new orgst 1",
				City:               "new orgci 1",
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 2, 4},
							Critical: true,
							Value:    []byte{0xaa, 0xbb, 0x11},
						},
						Description: "a",
					},
				},
				PreferredRootCN:             "different new cn",
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(222223333, 0),
				ApiKey:                      "api-key---",
				ApiKeyNew:                   "api-key-new---",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "./app.exe",
				PostProcessingEnvironment:   []string{"a=123", "b=zba"},
				PostProcessingClientAddress: "xyz.com",
				PostProcessingClientKeyB64:  "aaa888aaabbbccc",
				Profile:                     "new prof 2",
			},
			nil,
			18,
			certificates.Certificate{
				ID:                 18,
				Name:               "somenewNameHere",
				Description:        "some new desc goes here",
				Key:                key58,
				Account:            acmeAcct2,
				Subject:            "desk.dude.example.com",
				SubjectAltNames:    []string{"new.example.com"},
				Organization:       "new org 1",
				OrganizationalUnit: "new orgu 1",
				Country:            "new orgco 1",
				State:              "new orgst 1",
				City:               "new orgci 1",
				CSRExtraExtensions: []certificates.CertExtension{
					{
						Extension: pkix.Extension{
							Id:       asn1.ObjectIdentifier{1, 2, 4},
							Critical: true,
							Value:    []byte{0xaa, 0xbb, 0x11},
						},
						Description: "a",
					},
				},
				PreferredRootCN:             "different new cn",
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(222223333, 0),
				ApiKey:                      "api-key---",
				ApiKeyNew:                   "api-key-new---",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "./app.exe",
				PostProcessingEnvironment:   []string{"a=123", "b=zba"},
				PostProcessingClientAddress: "xyz.com",
				PostProcessingClientKeyB64:  "aaa888aaabbbccc",
				Profile:                     "new prof 2",
			},
			nil,
		},
		// update nothing (except mandatory update time)
		{
			certificates.UpdatePayload{
				ID:        27,
				UpdatedAt: time.Unix(15151515, 0),
			},
			certificates.Certificate{
				ID:                          27,
				Name:                        "test008.test.example.com-p",
				Description:                 "",
				Key:                         key56,
				Account:                     acmeAcct2,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{"test011.test.example.com", "*.test011.test.example.com"},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743171060, 0),
				UpdatedAt:                   time.Unix(15151515, 0),
				ApiKey:                      "api-secret-27",
				ApiKeyNew:                   "",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
			27,
			certificates.Certificate{
				ID:                          27,
				Name:                        "test008.test.example.com-p",
				Description:                 "",
				Key:                         key56,
				Account:                     acmeAcct2,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{"test011.test.example.com", "*.test011.test.example.com"},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743171060, 0),
				UpdatedAt:                   time.Unix(15151515, 0),
				ApiKey:                      "api-secret-27",
				ApiKeyNew:                   "",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
		},
		// update a couple things (including empty slice)
		{
			certificates.UpdatePayload{
				ID:                          27,
				SubjectAltNames:             []string{},
				PostProcessingClientAddress: new("someaddr.example.com"),
				UpdatedAt:                   time.Unix(151222215, 0),
			},
			certificates.Certificate{
				ID:                          27,
				Name:                        "test008.test.example.com-p",
				Description:                 "",
				Key:                         key56,
				Account:                     acmeAcct2,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743171060, 0),
				UpdatedAt:                   time.Unix(151222215, 0),
				ApiKey:                      "api-secret-27",
				ApiKeyNew:                   "",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "someaddr.example.com",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
			27,
			certificates.Certificate{
				ID:                          27,
				Name:                        "test008.test.example.com-p",
				Description:                 "",
				Key:                         key56,
				Account:                     acmeAcct2,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743171060, 0),
				UpdatedAt:                   time.Unix(151222215, 0),
				ApiKey:                      "api-secret-27",
				ApiKeyNew:                   "",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "someaddr.example.com",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putcertupdate")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			c, err := storage.PutCertUpdate(tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, c, tc.expectedPutResult)

			c, err = storage.GetOneCertById(tc.getId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, c, tc.expectedGetResult)
		})
	}
}

func TestPutCertApiKey(t *testing.T) {
	testCases := []struct {
		certId     int
		apiKey     string
		updateTime time.Time

		expectedCert   certificates.Certificate
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid id
			-1,
			"fake",
			time.Unix(100005, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid id
			500,
			"anotherfake",
			time.Unix(10005000, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		// do some updates
		{
			18,
			"somekey31cert",
			time.Unix(101300522, 0),
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
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(101300522, 0),
				ApiKey:                      "somekey31cert",
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
			26,
			"",
			time.Unix(0, 0),
			certificates.Certificate{
				ID:                          26,
				Name:                        "test008.test.example.com",
				Description:                 "",
				Key:                         key55,
				Account:                     acmeAcct1,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743170701, 0),
				UpdatedAt:                   time.Unix(0, 0),
				ApiKey:                      "",
				ApiKeyNew:                   "",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "test008.test.example.com",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putcertapikey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.certId), func(t *testing.T) {
			err := storage.PutCertApiKey(tc.certId, tc.apiKey, tc.updateTime)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put cert api key error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			cert, err := storage.GetOneCertById(tc.certId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get cert error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, cert, tc.expectedCert)
		})
	}
}

func TestCertApiKeyNew(t *testing.T) {
	testCases := []struct {
		certId    int
		apiKeyNew string
		updatedAt time.Time

		expectedCert   certificates.Certificate
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid id
			-1,
			"",
			time.Unix(10999051, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid id
			500,
			"anotherfake",
			time.Unix(10445000, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		// do some updates
		{
			27,
			"certkey27",
			time.Unix(102202305, 0),
			certificates.Certificate{
				ID:                          27,
				Name:                        "test008.test.example.com-p",
				Description:                 "",
				Key:                         key56,
				Account:                     acmeAcct2,
				Subject:                     "test008.test.example.com",
				SubjectAltNames:             []string{"test011.test.example.com", "*.test011.test.example.com"},
				Organization:                "",
				OrganizationalUnit:          "",
				Country:                     "",
				State:                       "",
				City:                        "",
				CSRExtraExtensions:          []certificates.CertExtension{},
				PreferredRootCN:             "",
				LastAccess:                  time.Unix(0, 0),
				CreatedAt:                   time.Unix(1743171060, 0),
				UpdatedAt:                   time.Unix(102202305, 0),
				ApiKey:                      "api-secret-27",
				ApiKeyNew:                   "certkey27",
				ApiKeyViaUrl:                false,
				PostProcessingCommand:       "",
				PostProcessingEnvironment:   []string{},
				PostProcessingClientAddress: "",
				PostProcessingClientKeyB64:  "",
				Profile:                     "",
			},
			nil,
			nil,
		},
		{
			18,
			"",
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
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(0, 0),
				ApiKey:                      "api-secret-18",
				ApiKeyNew:                   "",
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
	storage, err := openStorageWithTestData(t, "putkeyapikeynew")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.certId), func(t *testing.T) {
			err := storage.PutCertApiKeyNew(tc.certId, tc.apiKeyNew, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected apikeynew put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			cert, err := storage.GetOneCertById(tc.certId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected cert get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, cert, tc.expectedCert)
		})
	}

}

func TestPutCertUpdatedAt(t *testing.T) {
	testCases := []struct {
		certId    int
		updatedAt time.Time

		expectedCert   certificates.Certificate
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id
			-1,
			time.Unix(28888111, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			time.Unix(28888222, 0),
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
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(0, 0),
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
			time.Unix(333444442, 0),
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
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(333444442, 0),
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
	storage, err := openStorageWithTestData(t, "putcertupdatedat")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.certId), func(t *testing.T) {
			err := storage.PutCertUpdatedAt(tc.certId, tc.updatedAt)
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

func TestPutCertClientKey(t *testing.T) {
	testCases := []struct {
		certId    int
		newKey    string
		updatedAt time.Time

		expectedCert   certificates.Certificate
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id 1
			-1,
			"new-b64-key",
			time.Unix(8883333, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id 2
			8888,
			"new-b64-key",
			time.Unix(88833331, 0),
			certificates.Certificate{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // valid update
			18,
			"new-b64-key-xxx111yyyy",
			time.Unix(88832231, 0),
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
				LastAccess:                  time.Unix(1745952074, 0),
				CreatedAt:                   time.Unix(1709327717, 0),
				UpdatedAt:                   time.Unix(88832231, 0),
				ApiKey:                      "api-secret-18",
				ApiKeyNew:                   "api-new-secret-18",
				ApiKeyViaUrl:                true,
				PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
				PostProcessingEnvironment:   []string{"asdasdasdsd=asasd"},
				PostProcessingClientAddress: "dude.greg.example.com",
				PostProcessingClientKeyB64:  "new-b64-key-xxx111yyyy",
				Profile:                     "tlsserver",
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putcertclientkey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.certId), func(t *testing.T) {
			err := storage.PutCertClientKey(tc.certId, tc.newKey, tc.updatedAt)
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
