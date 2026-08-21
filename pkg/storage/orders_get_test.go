package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"fmt"
	"testing"
	"time"
)

// GetAllValidCurrentOrders
// GetOrdersByCert
// GetAllIncompleteOrderIds
// GetNewestIncompleteCertOrderId
// GetOrders
// GetOneOrder
// GetCertNewestValidOrderById
// GetCertNewestValidOrderByName

var (
	ord203 = orders.Order{
		ID: 203,
		Certificate: certificates.Certificate{
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
		},
		Location:     "https://acme-v02.api.letsencrypt.org/acme/order/red-2/red-203",
		Status:       "valid",
		KnownRevoked: false,
		Error:        nil,
		Expires:      new(1779411544),
		DnsIdentifiers: []string{
			"desk.dude.example.com",
		},
		Authorizations: []string{
			"https://acme-v02.api.letsencrypt.org/acme/authz/red-2/r-6876",
		},
		Finalize:       "https://acme-v02.api.letsencrypt.org/acme/finalize/red-2/red-203",
		FinalizedKey:   &key31,
		CertificateUrl: new("https://acme-v02.api.letsencrypt.org/acme/cert/red-203"),
		Pem: new(`-----BEGIN CERTIFICATE-----
red-203
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIICizCCAhGgAwIBAgIQXd1w3TH4AchcGGp6BLgK/jAKBggqhkjOPQQDAzAuMQsw
CQYDVQQGEwJVUzENMAsGA1UEChMESVNSRzEQMA4GA1UEAxMHUm9vdCBZRTAeFw0y
NTA5MDMwMDAwMDBaFw0yODA5MDIyMzU5NTlaMDMxCzAJBgNVBAYTAlVTMRYwFAYD
VQQKEw1MZXQncyBFbmNyeXB0MQwwCgYDVQQDEwNZRTEwdjAQBgcqhkjOPQIBBgUr
gQQAIgNiAAQHZVB1/mimla2hfSurylScjPMZaOJXLz/NnAc2sylm8WDyhU9Ccp+z
ASQi5vSwGGJjSGklkD9fdPR8GpyDIOIjCEfrnbt/v+ZSEPLLEGbaM6EccDbN7p9x
teIm2Avf+ryjge4wgeswDgYDVR0PAQH/BAQDAgGGMBMGA1UdJQQMMAoGCCsGAQUF
BwMBMBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFLsgykcL/tflnPmPCSqj
jDdFsbzYMB8GA1UdIwQYMBaAFKPIJlqOoUzQNWP8myPIOq5W809WMDIGCCsGAQUF
BwEBBCYwJDAiBggrBgEFBQcwAoYWaHR0cDovL3llLmkubGVuY3Iub3JnLzATBgNV
HSAEDDAKMAgGBmeBDAECATAnBgNVHR8EIDAeMBygGqAYhhZodHRwOi8veWUuYy5s
ZW5jci5vcmcvMAoGCCqGSM49BAMDA2gAMGUCMQDgjUEahFT/h3DRakqiPZpLvPgf
Zwkt6K2EOMmh1nvEzl83eMLYcod4GCl3b0J1Nn0CMBNYmEQJb4CEG5WoOe7aRn/L
VKu6saHmHEynI7ysIPd8zQsK1HdmhlHKlw9Z5GpGvA==
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIICpjCCAiugAwIBAgIRAIchZfw0tuX7qK3Vs3BftTowCgYIKoZIzj0EAwMwTzEL
MAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2VhcmNo
IEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDIwHhcNMjYwNTEzMDAwMDAwWhcN
MzIwOTAyMjM1OTU5WjAuMQswCQYDVQQGEwJVUzENMAsGA1UEChMESVNSRzEQMA4G
A1UEAxMHUm9vdCBZRTB2MBAGByqGSM49AgEGBSuBBAAiA2IABDwS/6vhrcVqcbBo
+wgdI3fwn9x7DNJJOY/lTOti0vkwuRN87RhEhTH17E7XyFjWsPYhIPt/wzOqxTd2
b+4ZJNy9ID04YywF9U5zasDVyGSNErVNtz8uSGh5izW87j77GaOB6zCB6DAOBgNV
HQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB
/zAdBgNVHQ4EFgQUo8gmWo6hTNA1Y/ybI8g6rlbzT1YwHwYDVR0jBBgwFoAUfEKW
rt5LSDv6kviejM9ti6lyN5UwMgYIKwYBBQUHAQEEJjAkMCIGCCsGAQUFBzAChhZo
dHRwOi8veDIuaS5sZW5jci5vcmcvMBMGA1UdIAQMMAowCAYGZ4EMAQIBMCcGA1Ud
HwQgMB4wHKAaoBiGFmh0dHA6Ly94Mi5jLmxlbmNyLm9yZy8wCgYIKoZIzj0EAwMD
aQAwZgIxAMU19WCtmxVND8UHBZRoma49Z7jPs64Dma0eTu1OChVbB/2J7GV3nvYK
Ax54uk1G9QIxAO0miLVJu8PLNiXXXkiE/gsK3CTRTF/aeo4bMX42Zw40csRU6AC2
6hSW1/IWaas6dg==
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIIEcDCCAligAwIBAgIQbI8dxyfHEX97r4U6yYD5zTANBgkqhkiG9w0BAQsFADBP
MQswCQYDVQQGEwJVUzEpMCcGA1UEChMgSW50ZXJuZXQgU2VjdXJpdHkgUmVzZWFy
Y2ggR3JvdXAxFTATBgNVBAMTDElTUkcgUm9vdCBYMTAeFw0yNjA1MTMwMDAwMDBa
Fw0zMjA5MDIyMzU5NTlaME8xCzAJBgNVBAYTAlVTMSkwJwYDVQQKEyBJbnRlcm5l
dCBTZWN1cml0eSBSZXNlYXJjaCBHcm91cDEVMBMGA1UEAxMMSVNSRyBSb290IFgy
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEzZvVn4CDCuwJSvMWSj5cz3es3mcFDR0H
ttwW+1qLFNvicWDEukWVEYmO6gbf9yoWHKS5xcUy4APgHoIYOIvXRdgKam7mAHf7
AlF9ItgKbppbd9/w+kHsOdx1ymgHDB/qo4H1MIHyMA4GA1UdDwEB/wQEAwIBBjAd
BgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zAd
BgNVHQ4EFgQUfEKWrt5LSDv6kviejM9ti6lyN5UwHwYDVR0jBBgwFoAUebRZ5nu2
5eQBc4AIiMgaWPbpm24wMgYIKwYBBQUHAQEEJjAkMCIGCCsGAQUFBzAChhZodHRw
Oi8veDEuaS5sZW5jci5vcmcvMBMGA1UdIAQMMAowCAYGZ4EMAQIBMCcGA1UdHwQg
MB4wHKAaoBiGFmh0dHA6Ly94MS5jLmxlbmNyLm9yZy8wDQYJKoZIhvcNAQELBQAD
ggIBAD2/e9frmMxNpCV03qUHegg+MV2wz9644YoXdqtH8RyWYcBO7xfjjGEXdU1e
/o0OkEFiynUCOSIk/vLLo7ttz6CPAeNlWfC0XNkoGeWgK6jjXvozBaGuGH5n0Ufo
shMeWTuURqNN5G00sSXDTBrpp2+mgvdZQjb8K11TYMA25QA+YHNfbIEL0BniAhKS
2gsnJjSzrdZLI+EZ7SEyqdR2rkjd1KutLDU+n3TFyxjniZVGur4YlhMP3mY/dV95
IruAkkjOZier6hGBdEgZXXvaCz9u9iVEadsIE75pAGL8oHV5vxdARDiotRpul1IN
/UZwzAbrfUFcw1HkAcYD/mlZfnQ2ieCF2MS7j3Vhv7JPDKp45fmykmzYNSrumRW0
upFFKDBOoF7hsOb7oLyHS+Uft6jOUfOrogj8YUx38hKb2K20r42OgsSdDdxdeYWc
MS3Sb6mwJeSZEYxJ2gaXnDSPaKhhrNkYwljyVQyr4Nq+MEJytXNTnHqaAcrNwZlV
pcJL1KBnMrMjP7eanvUwL3FYj3cF17jtboLt7gLoi4+2rWZFvn+w54jmd/FIuhhZ
cEaU/wvU6BUNMtcVquVGHp7itQeDth5j+XL3j4WJ2SABwzUl6OeYdgpIt/ITZa+p
TT0mQ/r5XyA4MEAiabn7XJjvCERlF2dcn2wqJw+CreTkkQ2R
-----END CERTIFICATE-----
`),
		ValidFrom:   new(time.Unix(1779382921, 0)),
		ValidTo:     new(time.Unix(1783270920, 0)),
		CreatedAt:   time.Unix(1779386425, 0),
		UpdatedAt:   time.Unix(1779386440, 0),
		ChainRootCN: new("ISRG Root X1"),
		Profile:     new("tlsserver"),
		RenewalInfo: &orders.RenewalInfo{
			SuggestedWindow: struct {
				Start time.Time "json:\"start\""
				End   time.Time "json:\"end\""
			}{
				Start: time.Unix(1781937245, 0),
				End:   time.Unix(1782014934, 0),
			},
			ExplanationURL: nil,
			RetryAfter:     new(time.Unix(1779410250, 0)), //
		},
	}

	ord198 = orders.Order{
		ID:             198,
		Certificate:    cert33,
		Location:       "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-198",
		Status:         "valid",
		KnownRevoked:   false,
		Error:          nil,
		Expires:        new(1779389985),
		DnsIdentifiers: []string{"test007.test.example2.com"},
		Authorizations: []string{"https://acme-staging-v02.api.letsencrypt.org/acme/authz/red-1/r-4643"},
		Finalize:       "https://acme-staging-v02.api.letsencrypt.org/acme/finalize/red-1/red-198",
		FinalizedKey:   &key69,
		CertificateUrl: new("https://acme-staging-v02.api.letsencrypt.org/acme/cert/red-198"),
		Pem: new(`-----BEGIN CERTIFICATE-----
red-198
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
		ValidFrom:   new(time.Unix(1779382883, 0)),
		ValidTo:     new(time.Unix(1783270882, 0)),
		ChainRootCN: new("(STAGING) Pretend Pear X1"),
		CreatedAt:   time.Unix(1779386387, 0),
		UpdatedAt:   time.Unix(1779802225, 0),
		Profile:     new("tlsserver"),
		RenewalInfo: &orders.RenewalInfo{
			SuggestedWindow: struct {
				Start time.Time "json:\"start\""
				End   time.Time "json:\"end\""
			}{
				Start: time.Unix(1781937207, 0),
				End:   time.Unix(1782014897, 0),
			},
			ExplanationURL: nil,
			RetryAfter:     new(time.Unix(1779820852, 0)), // note: this was manually converted to zulu in the test db to ensure compatibility
		},
	}
)

func TestGetCertNewestValidOrderById(t *testing.T) {
	testCases := []struct {
		id int

		expectedOrd orders.Order
		expectedErr error
	}{
		{-1, orders.Order{}, sql.ErrNoRows},
		{666, orders.Order{}, sql.ErrNoRows},
		{18, ord203, nil},                   // 18: newest is valid, case is wrong (also has a createdAt tie that must be broken by order.id)
		{35, orders.Order{}, sql.ErrNoRows}, // 35: no valid order
		{28, orders.Order{}, sql.ErrNoRows}, // 28: newest valid is expired
		{31, orders.Order{}, sql.ErrNoRows}, // 31: all valid orders but expired
		{33, ord198, nil},                   // 33: newest is valid but revoked, drop back to next newest valid
		{26, orders.Order{}, sql.ErrNoRows}, // 26: newest valid is expired
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getcertnewestvalidorderbyid")
	if err != nil {
		t.Fatal(err)
	}

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedOrd)
		})
	}
}

func TestGetCertNewestValidOrderByName(t *testing.T) {
	testCases := []struct {
		name string

		expectedOrd orders.Order
		expectedErr error
	}{
		{"fake-bad-name", orders.Order{}, sql.ErrNoRows},
		{"", orders.Order{}, sql.ErrNoRows},
		{"serverDEFault", ord203, nil},                                                // 18: newest is valid, case is wrong
		{"STAGING_persist--test005.test.example2.com", orders.Order{}, sql.ErrNoRows}, // 35: no valid order
		{"a0.alias.test.example.com", orders.Order{}, sql.ErrNoRows},                  // 28: newest valid is expired
		{"SomeSmallTest", orders.Order{}, sql.ErrNoRows},                              // 31: all valid orders but expired
		{"STAGING_persist--test007.test.example2.com", ord198, nil},                   // 33: newest is valid but revoked, drop back to next newest valid
		{"test008.test.example.com", orders.Order{}, sql.ErrNoRows},                   // 26: newest valid is expired
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getcertnewestvalidorderbyname")
	if err != nil {
		t.Fatal(err)
	}

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedOrd)
		})
	}
}
