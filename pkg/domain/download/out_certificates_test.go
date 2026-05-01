package download

import (
	"certwarden-backend/pkg/output"
	"testing"
)

func TestOutCertificatesViaHeader(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No header provided
	oneTest(t, service.DownloadCertViaHeader, nil, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, nil, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, nil, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, nil, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, nil, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: key apikey provided instead of cert apikey
	apiKey = "k-123"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: key apikey variants
	apiKey = ".k-123"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123."
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided but via url
	apiKey = "c-abc"
	oneTest(t, service.DownloadCertViaHeader, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadCertViaHeader, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	// `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadCertViaHeader, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct new apikey provided but via url
	apiKey = "c-abc-new"
	oneTest(t, service.DownloadCertViaHeader, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaHeader, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-a", `-----BEGIN CERTIFICATE-----
MIIBOzCB5qADAgECAgFvMA0GCSqGSIb3DQEBCwUAMB4xHDAaBgNVBAoTE0ludGVy
bWVkaWF0ZS1UZXN0LWEwHhcNMjYwNTAxMjIwNDMxWhcNMzYwNTAxMjIwNDMxWjAW
MRQwEgYDVQQKEwtMZWFmLVRlc3QtYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQCv
ALveirKCCH4P0VX2ALGIu+rLoFPTuIoe94iPSdUVyiZYhtU94CwInFm1NiK0SJ88
hRebXl34ueNysTUqBomnAgMBAAGjFzAVMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA0G
CSqGSIb3DQEBCwUAA0EApQ9eotS7RB9V8N06RTE6KJXtoTt5niwG0janr3fOYdk2
6VVG1yS6fkb/Lyn35W/6LnIoGOwB1C8xntj5yvzTcw==
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBCzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYTAeFw0yNjA1MDEyMjA0MzFaFw0zNjA1MDEyMjA0MzFaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWEwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rwC73oqyggh+D9FV9gCxiLvqy6BT07iKHveIj0nVFcomWIbVPeAsCJxZtTYitEif
PIUXm15d+LnjcrE1KgaJpwIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUnzsNoqfuhd+iva9d5im/++NPU5gwDQYJKoZI
hvcNAQELBQADQQBOwWQzYp8fj1Dk9TzikyoYdmONKWYknSSNjzD/kqMMI4fr51qM
sp5CbPfXs9p7/ovby8Fmd4QY+CVXwlYrZZnI
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYTAeFw0yNjA1MDEyMjA0MzFaFw0zNjA1MDEyMjA0MzFaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1hMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMR8JRKxAetS
CF7wo/1pSnNDXz6ox/5AH+FXVVBy0LmVayxrgPxr5OVDAcgDwUx2gPB77XmtyCq8
dmOfwm+vf3sCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFH+dsDe3QdO7VwziDWsg7E7VScrtMA0GCSqGSIb3DQEBCwUA
A0EAT/w/FWrwVzV6uau8H2qIhBuM+fbVAlVpVuvg5AxVwAOv+YhuTc83BTakfXgH
8Joh9m0OL1x+zBvxoby5o9Xg3Q==
-----END CERTIFICATE-----`, nil)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-c", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgIBTTANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1jMB4XDTI2MDUwMTIyMTM0MloXDTM2MDUwMTIyMTM0Mlow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWMwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
pwYElBHs6irbGuWJFiQ0ydyiU4m1CrAcH53nE3+1vrIUS3k75QxGgndf4A8L5iIV
mmLJWCTF4S0lLF5DqBYEMwIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAKUgPgdxletP9b5RP8LOmGifDVDI+TQHHbSFdpkeNXm9
mPv9L5+jK4g+0dedMKEy3Yk1FmfSIEXS1Xohk4BMNW8=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBITANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYzAeFw0yNjA1MDEyMjEzNDJaFw0zNjA1MDEyMjEzNDJaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWMwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
pwYElBHs6irbGuWJFiQ0ydyiU4m1CrAcH53nE3+1vrIUS3k75QxGgndf4A8L5iIV
mmLJWCTF4S0lLF5DqBYEMwIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUlFkGfdyfNVmhMQajsJA8JNp7sEgwDQYJKoZI
hvcNAQELBQADQQAo7jCPam5K2xZqcvjVQBGhIi5Zxc+++XDbsHttiwWVWdE1DkqV
roblaZUgYh2sf2XTG1U12LfTJ7LS9X3mC8HO
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBAzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYzAeFw0yNjA1MDEyMjEzNDJaFw0zNjA1MDEyMjEzNDJaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1jMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANSk+BJk25sr
O2XBe2yWlvqbQB817vJKOq8WEGCsj25W3XSDyGByvMTLUGM9nGH4mlE0moePVnf7
gQek+T7syC8CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFKBtmH4P8sXj3S55qI77ZL62ssufMA0GCSqGSIb3DQEBCwUA
A0EATsMeVCPlMKlPADgnH6XLiL84ofGJNtUQ8YgO3EY2Jgpr0HELG1oSSizWNRes
gGAROyiIETRshG2BbfFolQxEUw==
-----END CERTIFICATE-----`, nil)
	// `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBNzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUUt4WDgWCFKQBP5pUgeJoaZ8OXUEwDQYJKoZI
hvcNAQELBQADQQB9FrEVhfLtmX502bzUwWdThLgrnDpbUvAwi5sn6pD3bK3BFmkl
g1/KbwHRr8PkIHTFtGbFNvZzjbrYb/FhHbSA
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1lMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKBp48fOFsbg
lz3dPzfZAwn6VPaDHkSBf8hBl3WHsd+XAErWYCzFLbDPrwTYLeoKTKBMvN6h36S6
q528WfMRsAkCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFPggvDzOv1cAig6XMGB8mxeogmguMA0GCSqGSIb3DQEBCwUA
A0EAEKXN9SngV8TnwoWVtZjHlZy3Y/kaY3vRfnRQB4cLeE6/UD4VshDFC9LbvH9P
BcJ3umu+SsvpqDQvQagT/D/bZg==
-----END CERTIFICATE-----`, nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new"
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-b", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgIA3jANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1iMB4XDTI2MDUwMTIyMTEwMFoXDTM2MDUwMTIyMTEwMFow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWIwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
wAccmz5RPyTjdQnpTDUDuFwfo93bK8BfRqSd6+5bk7P3Qrb7pHfF3IZ/G4/VQKik
SvcZsBXvUs/89Z6CrX9k0wIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAJ2RTakwU/3dR3mZfhHBHairYA33pWtWNwBZK1GafOAb
tQIixzXryAJo0bty01ipOrwViK8QAruggUFDTQ7trfo=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBFjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYjAeFw0yNjA1MDEyMjExMDBaFw0zNjA1MDEyMjExMDBaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWIwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
wAccmz5RPyTjdQnpTDUDuFwfo93bK8BfRqSd6+5bk7P3Qrb7pHfF3IZ/G4/VQKik
SvcZsBXvUs/89Z6CrX9k0wIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUEI05lyluyXVvqiWPgeCKTFENnXEwDQYJKoZI
hvcNAQELBQADQQAPrSh4VGjediCq2kWtkkQciRrz+nyRKKH50vbpTmImNLsxo4DW
lu3c8NbPDv+Dl+nlovC7VvHm6Uq8aKcPpCuh
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBAjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtYjAeFw0yNjA1MDEyMjExMDBaFw0zNjA1MDEyMjExMDBaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1iMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANeEKvEOGie+
J73rLRWN8fblF6pZ31tMeVAGXWvi1J6ZiLtDU2Dj65+9zWo/8TfIUtj9VwJRuPeM
DNoVA5fG8wkCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFCqm2NbHYf7QY4LEPLZVa02U2lT9MA0GCSqGSIb3DQEBCwUA
A0EAq7awvHcloElKaKXjHwnYADpmGHlzKxWXCYpbYewq+MWjkyNOz5+mtUPXR/Jk
abU3C9xvwy61nVOjwhJpLlOSkA==
-----END CERTIFICATE-----`, nil)
	oneTest(t, service.DownloadCertViaHeader, &apiKey, nil, "test-e", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBNzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUUt4WDgWCFKQBP5pUgeJoaZ8OXUEwDQYJKoZI
hvcNAQELBQADQQB9FrEVhfLtmX502bzUwWdThLgrnDpbUvAwi5sn6pD3bK3BFmkl
g1/KbwHRr8PkIHTFtGbFNvZzjbrYb/FhHbSA
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1lMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKBp48fOFsbg
lz3dPzfZAwn6VPaDHkSBf8hBl3WHsd+XAErWYCzFLbDPrwTYLeoKTKBMvN6h36S6
q528WfMRsAkCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFPggvDzOv1cAig6XMGB8mxeogmguMA0GCSqGSIb3DQEBCwUA
A0EAEKXN9SngV8TnwoWVtZjHlZy3Y/kaY3vRfnRQB4cLeE6/UD4VshDFC9LbvH9P
BcJ3umu+SsvpqDQvQagT/D/bZg==
-----END CERTIFICATE-----`, nil)
}

func TestOutCertificatesViaURL(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No url value provided
	oneTest(t, service.DownloadCertViaUrl, nil, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: key apikey provided instead of cert apikey
	apiKey = "k-123"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: key apikey variants
	apiKey = ".k-123"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123."
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	// `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBNzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUUt4WDgWCFKQBP5pUgeJoaZ8OXUEwDQYJKoZI
hvcNAQELBQADQQB9FrEVhfLtmX502bzUwWdThLgrnDpbUvAwi5sn6pD3bK3BFmkl
g1/KbwHRr8PkIHTFtGbFNvZzjbrYb/FhHbSA
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1lMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKBp48fOFsbg
lz3dPzfZAwn6VPaDHkSBf8hBl3WHsd+XAErWYCzFLbDPrwTYLeoKTKBMvN6h36S6
q528WfMRsAkCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFPggvDzOv1cAig6XMGB8mxeogmguMA0GCSqGSIb3DQEBCwUA
A0EAEKXN9SngV8TnwoWVtZjHlZy3Y/kaY3vRfnRQB4cLeE6/UD4VshDFC9LbvH9P
BcJ3umu+SsvpqDQvQagT/D/bZg==
-----END CERTIFICATE-----`, nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new"
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadCertViaUrl, nil, &apiKey, "test-e", `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBNzANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUUt4WDgWCFKQBP5pUgeJoaZ8OXUEwDQYJKoZI
hvcNAQELBQADQQB9FrEVhfLtmX502bzUwWdThLgrnDpbUvAwi5sn6pD3bK3BFmkl
g1/KbwHRr8PkIHTFtGbFNvZzjbrYb/FhHbSA
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZTAeFw0yNjA1MDEyMjE0NDdaFw0zNjA1MDEyMjE0NDdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1lMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKBp48fOFsbg
lz3dPzfZAwn6VPaDHkSBf8hBl3WHsd+XAErWYCzFLbDPrwTYLeoKTKBMvN6h36S6
q528WfMRsAkCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFPggvDzOv1cAig6XMGB8mxeogmguMA0GCSqGSIb3DQEBCwUA
A0EAEKXN9SngV8TnwoWVtZjHlZy3Y/kaY3vRfnRQB4cLeE6/UD4VshDFC9LbvH9P
BcJ3umu+SsvpqDQvQagT/D/bZg==
-----END CERTIFICATE-----`, nil)
}
