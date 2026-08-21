package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
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

	cert28 = certificates.Certificate{
		ID:                          28,
		Name:                        "a0.alias.test.example.com",
		Description:                 "",
		Key:                         key57,
		Account:                     acmeAcct1,
		Subject:                     "a1.alias.test.example.com",
		SubjectAltNames:             []string{"q1.alias.test.example.com", "*.q1.alias.test.example.com", "q2.alias.test.example.com", "*.q2.alias.test.example.com", "q3.alias.test.example.com", "*.q3.alias.test.example.com", "q4.alias.test.example.com", "*.q4.alias.test.example.com", "q5.alias.test.example.com", "*.q5.alias.test.example.com", "q6.alias.test.example.com", "*.q6.alias.test.example.com"},
		Organization:                "",
		OrganizationalUnit:          "",
		Country:                     "",
		State:                       "",
		City:                        "",
		CSRExtraExtensions:          []certificates.CertExtension{},
		PreferredRootCN:             "",
		LastAccess:                  time.Unix(0, 0),
		CreatedAt:                   time.Unix(1743173262, 0),
		UpdatedAt:                   time.Unix(1777494241, 0),
		ApiKey:                      "api-secret-28",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                false,
		PostProcessingCommand:       "",
		PostProcessingEnvironment:   []string{},
		PostProcessingClientAddress: "asdasdas.com",
		PostProcessingClientKeyB64:  "aaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaa",
		Profile:                     "",
	}

	cert33 = certificates.Certificate{
		ID:                          33,
		Name:                        "STAGING_persist--test007.test.example2.com",
		Description:                 "",
		Key:                         key64,
		Account:                     acmeAcct1,
		Subject:                     "test007.test.example2.com",
		SubjectAltNames:             []string{},
		Organization:                "",
		OrganizationalUnit:          "",
		Country:                     "",
		State:                       "",
		City:                        "",
		CSRExtraExtensions:          []certificates.CertExtension{},
		PreferredRootCN:             "",
		LastAccess:                  time.Unix(1777555692, 0),
		CreatedAt:                   time.Unix(1775761592, 0),
		UpdatedAt:                   time.Unix(1779386476, 0),
		ApiKey:                      "api-secret-33",
		ApiKeyNew:                   "",
		ApiKeyViaUrl:                false,
		PostProcessingCommand:       "",
		PostProcessingEnvironment:   []string{},
		PostProcessingClientAddress: "",
		PostProcessingClientKeyB64:  "",
		Profile:                     "tlsserver",
	}

	cert35 = certificates.Certificate{
		ID:                          35,
		Name:                        "STAGING_persist--test005.test.example2.com",
		Description:                 "",
		Key:                         key70,
		Account:                     acmeAcct1,
		Subject:                     "test005.test.example2.com",
		SubjectAltNames:             []string{},
		Organization:                "",
		OrganizationalUnit:          "",
		Country:                     "",
		State:                       "",
		City:                        "",
		CSRExtraExtensions:          []certificates.CertExtension{},
		PreferredRootCN:             "",
		LastAccess:                  time.Unix(0, 0),
		CreatedAt:                   time.Unix(1779386771, 0),
		UpdatedAt:                   time.Unix(1779386798, 0),
		ApiKey:                      "api-secret-35",
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
		{queryBuilderForTest(1, 1, "id", false), 9, 1, 0, cert26},
		{queryBuilderForTest(3, 6, "servername", true), 9, 3, 1, cert27},
		{queryBuilderForTest(4, 1, "accountname", false), 9, 4, 0, cert18},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getallcerts")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedAtIndx.Name), func(t *testing.T) {
			certs, totalCt, err := store.GetAllCerts(tc.q)
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
				compareCertificate(t, certs[tc.testIndx], &tc.expectedAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(certs))
			}
		})
	}
}

func TestGetOneCertById(t *testing.T) {
	testCases := []struct {
		id           int
		expectedCert *certificates.Certificate
		expectedErr  error
	}{
		{-5, nil, sql.ErrNoRows},
		{50, nil, sql.ErrNoRows},
		{18, &cert18, nil},
		{26, &cert26, nil},
		{27, &cert27, nil},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getonecertbyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			serv, err := store.GetOneCertById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareCertificate(t, serv, tc.expectedCert)
		})
	}
}

func TestGetOneCertByName(t *testing.T) {
	testCases := []struct {
		name         string
		expectedCert *certificates.Certificate
		expectedErr  error
	}{
		{"fake-bad-name", nil, sql.ErrNoRows},
		{"", nil, sql.ErrNoRows},
		{"serverDEFault", &cert18, nil}, // case is wrong
		{"test008.test.example.com", &cert26, nil},
		{"test008.test.example.com-p", &cert27, nil},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getonecertbyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			serv, err := store.GetOneCertByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareCertificate(t, serv, tc.expectedCert)
		})
	}
}
