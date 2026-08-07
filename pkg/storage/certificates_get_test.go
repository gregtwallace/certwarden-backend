package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/test_helpers"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"fmt"
	"testing"
	"time"
)

var (
	cert18 = certificates.Certificate{
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
		UpdatedAt:                   time.Unix(1779386440, 0),
		ApiKey:                      "api-secret-18",
		ApiKeyNew:                   "api-new-secret-18",
		ApiKeyViaUrl:                true,
		PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
		PostProcessingEnvironment:   []string{"asdasdasdsd=asasd"},
		PostProcessingClientAddress: "dude.greg.example.com",
		PostProcessingClientKeyB64:  "aaaaaaaaaaaaaaaaaaaaaaaaaaa-ccccccccccccccc",
		Profile:                     "tlsserver",
	}

	cert26 = certificates.Certificate{
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
		UpdatedAt:                   time.Unix(1765392360, 0),
		ApiKey:                      "api-secret-26",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                false,
		PostProcessingCommand:       "./scripts/windows/post-processing.example.ps1",
		PostProcessingEnvironment:   []string{},
		PostProcessingClientAddress: "test008.test.example.com",
		PostProcessingClientKeyB64:  "",
		Profile:                     "",
	}

	cert27 = certificates.Certificate{
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
		UpdatedAt:                   time.Unix(1746122825, 0),
		ApiKey:                      "api-secret-27",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                false,
		PostProcessingCommand:       "",
		PostProcessingEnvironment:   []string{},
		PostProcessingClientAddress: "",
		PostProcessingClientKeyB64:  "",
		Profile:                     "",
	}
)

func TestGetAllCerts(t *testing.T) {
	testCases := []struct {
		q                 pagination_sort.Query
		expectedTotalCt   int
		expectedResultLen int
		testIndx          int
		expectedAtIndx    certificates.Certificate
	}{
		{pagination_sort.Query{}, 9, 9, 2, cert18},
		{QueryBuilderForTest(1, 1, "id", true), 9, 1, 0, cert26},
		{QueryBuilderForTest(3, 6, "servername", false), 9, 3, 1, cert27},
		{QueryBuilderForTest(4, 1, "accountname", true), 9, 4, 0, cert18},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getallcerts")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedAtIndx.Name), func(t *testing.T) {
			certs, totalCt, err := storage.GetAllCerts(tc.q)
			if err != nil {
				t.Errorf("get all failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(certs) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(certs))
			}
			if tc.testIndx <= len(certs)-1 {
				CompareCertificate(t, certs[tc.testIndx], tc.expectedAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(certs))
			}
		})
	}
}

func TestGetOneCertById(t *testing.T) {
	testCases := []struct {
		id           int
		expectedErr  error
		expectedCert certificates.Certificate
	}{
		{-5, sql.ErrNoRows, certificates.Certificate{}},
		{50, sql.ErrNoRows, certificates.Certificate{}},
		{18, nil, cert18},
		{26, nil, cert26},
		{27, nil, cert27},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getonecertbyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			serv, err := storage.GetOneCertById(tc.id)
			if !test_helpers.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(err))
			}

			CompareCertificate(t, serv, tc.expectedCert)
		})
	}
}

func TestGetOneCertByName(t *testing.T) {
	testCases := []struct {
		name         string
		expectedErr  error
		expectedCert certificates.Certificate
	}{
		{"fake-bad-name", sql.ErrNoRows, certificates.Certificate{}},
		{"", sql.ErrNoRows, certificates.Certificate{}},
		{"serverdefault", nil, cert18},
		{"test008.test.example.com", nil, cert26},
		{"test008.test.example.com-p", nil, cert27},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getonecertbyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			serv, err := storage.GetOneCertByName(tc.name)
			if !test_helpers.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(err))
			}

			CompareCertificate(t, serv, tc.expectedCert)
		})
	}
}
