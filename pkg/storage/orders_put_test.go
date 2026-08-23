package storage_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// TODO:
// PutOrderAcme
// PutRenewalInfo
// PutOrderInvalid
// UpdateFinalizedKey
// PutOrderPemData // TODO: need to rewrite some logic to do proper testing

// Useful for removing CR chars from pem
// UPDATE acme_orders
// SET
//   pem = REPLACE(pem, CHAR(13), '')
// WHERE id = [insert id here];

func TestPutOrderPemData(t *testing.T) {
	testCases := []struct {
		ordID   int
		payload orders.CertPayload

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id -2
		{
			-2,
			orders.CertPayload{
				AcmeCert:    new(acme.Certificate{}),
				RenewalInfo: new(orders.RenewalInfo{}),
				UpdatedAt:   time.Unix(2, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id 444
		{
			444,
			orders.CertPayload{
				AcmeCert:    new(acme.Certificate{}),
				RenewalInfo: new(orders.RenewalInfo{}),
				UpdatedAt:   time.Unix(2, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// TODO: Addl tests
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderpemdata")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordID), func(t *testing.T) {
			err := store.PutOrderPemData(tc.ordID, tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order revoke error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

func TestPutOrderRevoke(t *testing.T) {
	testCases := []struct {
		ordId     int
		updatedAt time.Time

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id -1
		{
			-1,
			time.Unix(77755552, 0),
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id 555
		{
			555,
			time.Unix(76655552, 0),
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// valid starting as known revoked false
		{
			204,
			time.Unix(45345345, 0),
			nil,
			orders.Order{
				ID:           204,
				Certificate:  cert35,
				Location:     "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-204",
				Status:       "invalid",
				KnownRevoked: true,
				Error: new(acme.Error{
					Status: 522,
					Type:   "urn:ietf:params:acme:error:badNonce",
					Detail: "some kind of error info",
				}),
				Expires:        nil,
				DnsIdentifiers: []string{"test005.test.example2.com"},
				Authorizations: []string{"https://acme-staging-v02.api.letsencrypt.org/acme/authz/red-1/r-9773"},
				Finalize:       "https://acme-staging-v02.api.letsencrypt.org/acme/finalize/red-1/red-204",
				FinalizedKey:   nil,
				CertificateUrl: nil,
				Pem:            nil,
				ValidFrom:      nil,
				ValidTo:        nil,
				CreatedAt:      time.Unix(1779386773, 0),
				UpdatedAt:      time.Unix(45345345, 0),
				ChainRootCN:    nil,
				Profile:        nil,
				RenewalInfo:    nil,
			},
			nil,
		},
		// valid starting as known revoked true
		{
			206,
			time.Unix(55663333, 0),
			nil,
			orders.Order{
				ID:             206,
				Certificate:    cert33,
				Location:       "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-206",
				Status:         "valid",
				KnownRevoked:   true,
				Error:          nil,
				Expires:        new(time.Unix(1779389990, 0)),
				DnsIdentifiers: []string{"test007.test.example2.com"},
				Authorizations: []string{"https://acme-staging-v02.api.letsencrypt.org/acme/authz/red-1/r-4643"},
				Finalize:       "https://acme-staging-v02.api.letsencrypt.org/acme/finalize/red-1/red-206",
				FinalizedKey:   &key69,
				CertificateUrl: new("https://acme-staging-v02.api.letsencrypt.org/acme/cert/red-206"),
				Pem: new(`-----BEGIN CERTIFICATE-----
red-206
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIICxTCCAkugAwIBAgIQCsMuCNHGfSpvkgUcBb9x8DAKBggqhkjOPQQDAzBHMQsw
CQYDVQQGEwJVUzENMAsGA1UEChMESVNSRzEpMCcGA1UEAxMgKFNUQUdJTkcpIFll
YXJuaW5nIFl1Y2NhIFJvb3QgWUUwHhcNMjUwOTAzMDAwMDAwWhcNMjgwOTAyMjM1
OTU5WjBMMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3MgRW5jcnlwdDElMCMG
A1UEAxMcKFNUQUdJTkcpIEJhbG9uZXkgQnVsZ3VyIFlFMjB2MBAGByqGSM49AgEG
BSuBBAAiA2IABGSNfIaSqN+sgsY9R8XDlo20ge3DDVKsB2H7MXDMrH6QHI7+8P12
FzUmDLlRHgTv5upCL9yM3Tm4/Fc8F+eizVRGS/nqjDXy3f4stt30AojDmT0ck06b
6L/5FZcONk8/06OB9jCB8zAOBgNVHQ8BAf8EBAMCAYYwEwYDVR0lBAwwCgYIKwYB
BQUHAwEwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU3tpCJfI0wCebZhbP
ZWqQrCHVMIcwHwYDVR0jBBgwFoAUkzsblChigfkfJ+uq35owDVquf+wwNgYIKwYB
BQUHAQEEKjAoMCYGCCsGAQUFBzAChhpodHRwOi8vc3RnLXllLmkubGVuY3Iub3Jn
LzATBgNVHSAEDDAKMAgGBmeBDAECATArBgNVHR8EJDAiMCCgHqAchhpodHRwOi8v
c3RnLXllLmMubGVuY3Iub3JnLzAKBggqhkjOPQQDAwNoADBlAjADQTdMtsV/BzKj
F6Wmw5YUqTT1TPFTEIul3UbsmlO688CVDpiv8Jxd35DijyFFqHECMQCr8jCqJ7uk
Aff/OJIClp0VybB1EM5JCAi/RnKleVVC6Ot0Erzxl9QuMu/vEiGbaw4=
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIIC3zCCAmWgAwIBAgIRAJi0CtKeMRLsmzQ3Qmq2GdkwCgYIKoZIzj0EAwMwaDEL
MAkGA1UEBhMCVVMxMzAxBgNVBAoTKihTVEFHSU5HKSBJbnRlcm5ldCBTZWN1cml0
eSBSZXNlYXJjaCBHcm91cDEkMCIGA1UEAxMbKFNUQUdJTkcpIEJvZ3VzIEJyb2Nj
b2xpIFgyMB4XDTI2MDUxMzAwMDAwMFoXDTMyMDkwMjIzNTk1OVowRzELMAkGA1UE
BhMCVVMxDTALBgNVBAoTBElTUkcxKTAnBgNVBAMTIChTVEFHSU5HKSBZZWFybmlu
ZyBZdWNjYSBSb290IFlFMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE9LGxyXLRIUhN
aCaufjztgLQYSwTDaqNhkD9nrf3IOkM1o5hqKUTGOm3SuJVqzzLrRXT7JL+UCtsh
YPi5VSbyaH8JXgogE+HlzMPK/jQEsjYo91J9mEjnl1RQyhMyI3wlo4HzMIHwMA4G
A1UdDwEB/wQEAwIBBjATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTAD
AQH/MB0GA1UdDgQWBBSTOxuUKGKB+R8n66rfmjANWq5/7DAfBgNVHSMEGDAWgBTe
0aNZZA7BmjajRu6wEHbvrpeVZjA2BggrBgEFBQcBAQQqMCgwJgYIKwYBBQUHMAKG
Gmh0dHA6Ly9zdGcteDIuaS5sZW5jci5vcmcvMBMGA1UdIAQMMAowCAYGZ4EMAQIB
MCsGA1UdHwQkMCIwIKAeoByGGmh0dHA6Ly9zdGcteDIuYy5sZW5jci5vcmcvMAoG
CCqGSM49BAMDA2gAMGUCMG0EVdwWYPCvcBoJuLUgSpZ/hIE37GfxpdNfXfaBwK3D
69xeQzdVoKN7mwDSCo6L9wIxALKTj6tmQz72ROfDSrD9mnwGsjqbxQ+r+Kg+Bf5a
EDlktNJNa0efiasw83EdIu3YIA==
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIIEqTCCApGgAwIBAgIRAKwx/YXsWhhVhStoMbbK6YwwDQYJKoZIhvcNAQELBQAw
ZjELMAkGA1UEBhMCVVMxMzAxBgNVBAoTKihTVEFHSU5HKSBJbnRlcm5ldCBTZWN1
cml0eSBSZXNlYXJjaCBHcm91cDEiMCAGA1UEAxMZKFNUQUdJTkcpIFByZXRlbmQg
UGVhciBYMTAeFw0yNjA1MTMwMDAwMDBaFw0zMjA5MDIyMzU5NTlaMGgxCzAJBgNV
BAYTAlVTMTMwMQYDVQQKEyooU1RBR0lORykgSW50ZXJuZXQgU2VjdXJpdHkgUmVz
ZWFyY2ggR3JvdXAxJDAiBgNVBAMTGyhTVEFHSU5HKSBCb2d1cyBCcm9jY29saSBY
MjB2MBAGByqGSM49AgEGBSuBBAAiA2IABDr0vsNZAswMWDiWwNOgMNBxT9rSwSyj
6BUKkfQDLJJdZwtve+XkKsnEfgAr2HpQPK38BVzmzB2Fydt1ywfnQIzyVTidjnLI
01ajuHXA1rvq0NlSC3ZyUWMqZ1dTDE4VcaOB/TCB+jAOBgNVHQ8BAf8EBAMCAQYw
HQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMA8GA1UdEwEB/wQFMAMBAf8w
HQYDVR0OBBYEFN7Ro1lkDsGaNqNG7rAQdu+ul5VmMB8GA1UdIwQYMBaAFLXzZfL+
sAqSH/s8ffNEoKxjJcMUMDYGCCsGAQUFBwEBBCowKDAmBggrBgEFBQcwAoYaaHR0
cDovL3N0Zy14MS5pLmxlbmNyLm9yZy8wEwYDVR0gBAwwCjAIBgZngQwBAgEwKwYD
VR0fBCQwIjAgoB6gHIYaaHR0cDovL3N0Zy14MS5jLmxlbmNyLm9yZy8wDQYJKoZI
hvcNAQELBQADggIBACJf2Il+hLJ3JTdneWO06d+UwqghsR1L1tRWRYP6SN490I8m
vE26XRTqZR80nsRJcpY6UmrUCxuWHUa8kmzyWB9RLpxPYXuVAYFYFrcmCabbuhC+
Po+hYazv6Q05O1/Fq0f6lbU1aXA0rL8Xe9U6oZkhRaQ3foij3J/0Y6LXSuhQoBLK
3QN3wAmwsgzscfFEabQwIUFIqDg7/kCyHPEmV19i3rDkLCDpyJTgfS3vW6024777
/ULBcMVNwikrY02NrYR3q6IsBP5oLjsiulDyJ51x9dnzJgKlK4vqEqJVhe54h4VW
mjvY/5rc8FWAb9HX2Bce065RC+iaURnrcIZe3CMN2lta90AiCJ2ObQlt6b6G8csL
N1hvCHYdKIPdz3Oz3d7+5VagAqdfsft8ZIGYWIkiOw2Urf9VFEU+JhysdSquBfWU
QcQc7wAswQX7SrDh4mirdBxoWZE8wyQSJnj2ckY5+iZN2F332potz6QdueKNzcKs
WlqjWBJWBgXpR+Zx4qStv/Czj1Fb4lfJhM0YHyxbjVQWZYGQOJKw7779+nt6hFBT
n8C1vrB+gAHz7ngiivApEFaRJP4u2TJ+E6rvo5b2M/B6e3sY/owLkfjK0iWnQkhx
LORExLpyf4W2494Zil1LsqWq+54a+RNzTBRfeazeXU5vXTITzvjaa5at42Ui
-----END CERTIFICATE-----
`),
				ValidFrom:   new(time.Unix(1779382890, 0)),
				ValidTo:     new(time.Unix(1783270890, 0)),
				CreatedAt:   time.Unix(1779386390, 0),
				UpdatedAt:   time.Unix(55663333, 0),
				ChainRootCN: new("(STAGING) Pretend Pear X1"),
				Profile:     new("tlsserver"),
				RenewalInfo: &orders.RenewalInfo{
					SuggestedWindow: struct {
						Start time.Time "json:\"start\""
						End   time.Time "json:\"end\""
					}{
						Start: time.Unix(1781937207, 0),
						End:   time.Unix(1782014897, 0),
					},
					ExplanationURL: new("https://example.com/explain"),
					RetryAfter:     new(time.Unix(1779820852, 0)), //
				},
			},
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putrevokeorder")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordId), func(t *testing.T) {
			err := store.PutOrderRevoke(tc.ordId, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order revoke error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}
