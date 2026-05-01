package download

import (
	"certwarden-backend/pkg/output"
	"testing"
)

func TestOutPrivateCertViaHeader(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No header provided
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "something.something"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: just one of the apikeys
	apiKey = "c-abc"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: apikey variants
	apiKey = ".k-123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = ".c-abc"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123."
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "c-abc."
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.c-abc"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "c-abc.123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided but via url
	apiKey = "c-abc.k-123"
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	// `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct new apikey provided but via url
	apiKey = "c-abc-new.k-123-new"
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaHeader, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc.k-123"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-a", `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBAMLwirxhhFBmtzbKk0+m+MBRBUPcj1CrDmNmvVlkmTTKCzY1RNVk
tOpgN6szMRlX1VRb+v5j7lJ5r2gJZrDNs8kCAwEAAQJACBwGRCSbuCszD1DJZLSM
f+ue7XNydCekN0G3OiMeNdI92AUYEb+Yh8meJIYGob8wcAYCt3pp/WhhoM8Qw8kf
BQIhAMaX+Dhswwehcf2hhO+eS0KNdB8i8demjJGLap+W/eZbAiEA+0osY3/LH24Z
xWboFT6ISJuriZZK24AqbeiS/IsYj6sCIQC+mAEInhE7FI2i/k3n7kKKd9l3PIFg
Fx6XXHcS/MVmOwIhAJi2lwtQ2oybSKYix+BBRGl70V+oKo4C8cYhlVJM5fxJAiBu
N4JkNXHxXM7m8/ItFqWJtKH2DCTDl5SSt64qUnEEbw==
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBOzCB5qADAgECAgFvMA0GCSqGSIb3DQEBCwUAMB4xHDAaBgNVBAoTE0ludGVy
bWVkaWF0ZS1UZXN0LWEwHhcNMjYwNTAxMjIwNDMxWhcNMzYwNTAxMjIwNDMxWjAW
MRQwEgYDVQQKEwtMZWFmLVRlc3QtYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQCv
ALveirKCCH4P0VX2ALGIu+rLoFPTuIoe94iPSdUVyiZYhtU94CwInFm1NiK0SJ88
hRebXl34ueNysTUqBomnAgMBAAGjFzAVMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA0G
CSqGSIb3DQEBCwUAA0EApQ9eotS7RB9V8N06RTE6KJXtoTt5niwG0janr3fOYdk2
6VVG1yS6fkb/Lyn35W/6LnIoGOwB1C8xntj5yvzTcw==
-----END CERTIFICATE-----
`, nil)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-c", `-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAN1VxmTOxWUcQjr7MNUxwkyJT5TuTQJU734/b9f8wKBVy87ikjFe
UGJLzYRJYutBwJBztdXnlhOS/bXRBs1szUsCAwEAAQJACD2RTV+FaeZLcPa5MrbP
jRnvpJPauiN/Zyvldh0q7s0xMQEZHVRmYUpsXZM4fFmSUvq3npBFptA3gNzOv8Hs
AQIhAOFkEtCJ3TyXRb0/pdyY8wQijWaRvOgGgYvKLjy7YPvBAiEA+2SyKDgh15pB
5yPyoGLg68tglMwm4VjVMFaeoiRw/AsCIQCIQVpKbX2senqzfL3FTUVkU4sN3b7I
ud4o5vHqzxBDQQIhANggRwZK09V/Gf90qUf4GjS9wYfLR/XeoFIRdgoh2DznAiEA
j2MbNxUDruMAHCN7tfis2cinNPVYxPGhPxrNWdikuoY=
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgIBTTANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1jMB4XDTI2MDUwMTIyMTM0MloXDTM2MDUwMTIyMTM0Mlow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWMwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
pwYElBHs6irbGuWJFiQ0ydyiU4m1CrAcH53nE3+1vrIUS3k75QxGgndf4A8L5iIV
mmLJWCTF4S0lLF5DqBYEMwIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAKUgPgdxletP9b5RP8LOmGifDVDI+TQHHbSFdpkeNXm9
mPv9L5+jK4g+0dedMKEy3Yk1FmfSIEXS1Xohk4BMNW8=
-----END CERTIFICATE-----
`, nil)
	// // `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL8nKLmviCA2UqLqEwTJ8mPnuN5k1iKw4W1RlxYYtTpvyDDkz5v6
C25MeHrbJqj0tzsBwXNZI2OHrBcSl3qWum0CAwEAAQJAKE4uu2Z8dsRVjNX74CLf
2fp6g+QxtbtjsQwG5kkb73/Um0phhoLDtMzgNg+MncKKgjx2WmVkX45LTP/TN8Pv
iQIhANbGnhTb6PD+ja1lTHuoxb16cKe3YjWRzu5QcqWwziE5AiEA49fKUxVQ5+gn
vYeAzr1zmAnmlvSpgEYiFIx8ENN7xtUCICN0qHYjE6JtM3BTj7u+Ud6EzwIw+OqF
Bpc6+qI1vOGpAiAZg7HBihKMVcAVhlYTUL3gGcO7xdwxZCku2eiOzc//nQIhALuF
N/VC8wvORy4lKBkAeRy0oK+9o6R5mdzP3Fjpz9I4
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
`, nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new.k-123-new"
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-b", `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALfdFbvTuf1r6Mk80ZLTfInivcfu9hF/JcRLnV+EOd4Z4/28zGBD
IwlnYkWDD7gBhuBRJoUPLnyXJ7Rp84+bp10CAwEAAQJAKJHNK6Bse+VdXFgB4zys
kG06VH0fCR3N2soXe728mguq9D3E3PyyFW/OyLUwWgXI3JXFC0+anu7oehFcE3o1
aQIhAOkYh4WJiyP7eBPcuuRNaZUweBsmZMkoW80B3W/RsGn5AiEAye4hhcuWPVuu
CjicvivY0I/y7tJ2nY/vXYfG1JqHoYUCIQDGWMghOpw6vyY7iI1D7heVCsx5Fd+X
SI9tUFP0bbM3SQIgFZuy4KNhh11ZKWTXeQ4uHFtbDq1c3g15+tM9tqB2pRUCIBT8
5URzNI/wCwqQD6D98UNKRJhD4MrDQlBBA9PYqnab
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgIA3jANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1iMB4XDTI2MDUwMTIyMTEwMFoXDTM2MDUwMTIyMTEwMFow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWIwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
wAccmz5RPyTjdQnpTDUDuFwfo93bK8BfRqSd6+5bk7P3Qrb7pHfF3IZ/G4/VQKik
SvcZsBXvUs/89Z6CrX9k0wIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAJ2RTakwU/3dR3mZfhHBHairYA33pWtWNwBZK1GafOAb
tQIixzXryAJo0bty01ipOrwViK8QAruggUFDTQ7trfo=
-----END CERTIFICATE-----
`, nil)
	oneTest(t, service.DownloadPrivateCertViaHeader, &apiKey, nil, "test-e", `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL8nKLmviCA2UqLqEwTJ8mPnuN5k1iKw4W1RlxYYtTpvyDDkz5v6
C25MeHrbJqj0tzsBwXNZI2OHrBcSl3qWum0CAwEAAQJAKE4uu2Z8dsRVjNX74CLf
2fp6g+QxtbtjsQwG5kkb73/Um0phhoLDtMzgNg+MncKKgjx2WmVkX45LTP/TN8Pv
iQIhANbGnhTb6PD+ja1lTHuoxb16cKe3YjWRzu5QcqWwziE5AiEA49fKUxVQ5+gn
vYeAzr1zmAnmlvSpgEYiFIx8ENN7xtUCICN0qHYjE6JtM3BTj7u+Ud6EzwIw+OqF
Bpc6+qI1vOGpAiAZg7HBihKMVcAVhlYTUL3gGcO7xdwxZCku2eiOzc//nQIhALuF
N/VC8wvORy4lKBkAeRy0oK+9o6R5mdzP3Fjpz9I4
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
`, nil)
}

func TestOutPrivateCertViaURL(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No header provided
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, nil, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, nil, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, nil, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, nil, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, nil, "test-e", "", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "something.something"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: just one of the apikeys
	apiKey = "c-abc"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: apikey variants
	apiKey = ".k-123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = ".c-abc"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123."
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "c-abc."
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "123.c-abc"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)
	apiKey = "c-abc.123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-d", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", "", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc.k-123"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-a", "", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-c", "", output.JsonErrUnauthorized)
	// // `d` doesnt have a any correct apikey
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL8nKLmviCA2UqLqEwTJ8mPnuN5k1iKw4W1RlxYYtTpvyDDkz5v6
C25MeHrbJqj0tzsBwXNZI2OHrBcSl3qWum0CAwEAAQJAKE4uu2Z8dsRVjNX74CLf
2fp6g+QxtbtjsQwG5kkb73/Um0phhoLDtMzgNg+MncKKgjx2WmVkX45LTP/TN8Pv
iQIhANbGnhTb6PD+ja1lTHuoxb16cKe3YjWRzu5QcqWwziE5AiEA49fKUxVQ5+gn
vYeAzr1zmAnmlvSpgEYiFIx8ENN7xtUCICN0qHYjE6JtM3BTj7u+Ud6EzwIw+OqF
Bpc6+qI1vOGpAiAZg7HBihKMVcAVhlYTUL3gGcO7xdwxZCku2eiOzc//nQIhALuF
N/VC8wvORy4lKBkAeRy0oK+9o6R5mdzP3Fjpz9I4
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
`, nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new.k-123-new"
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-b", "", output.JsonErrUnauthorized)
	oneTest(t, service.DownloadPrivateCertViaUrl, nil, &apiKey, "test-e", `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL8nKLmviCA2UqLqEwTJ8mPnuN5k1iKw4W1RlxYYtTpvyDDkz5v6
C25MeHrbJqj0tzsBwXNZI2OHrBcSl3qWum0CAwEAAQJAKE4uu2Z8dsRVjNX74CLf
2fp6g+QxtbtjsQwG5kkb73/Um0phhoLDtMzgNg+MncKKgjx2WmVkX45LTP/TN8Pv
iQIhANbGnhTb6PD+ja1lTHuoxb16cKe3YjWRzu5QcqWwziE5AiEA49fKUxVQ5+gn
vYeAzr1zmAnmlvSpgEYiFIx8ENN7xtUCICN0qHYjE6JtM3BTj7u+Ud6EzwIw+OqF
Bpc6+qI1vOGpAiAZg7HBihKMVcAVhlYTUL3gGcO7xdwxZCku2eiOzc//nQIhALuF
N/VC8wvORy4lKBkAeRy0oK+9o6R5mdzP3Fjpz9I4
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICKzANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1lMB4XDTI2MDUwMTIyMTQ0N1oXDTM2MDUwMTIyMTQ0N1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWUwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
rsPDF/cFzPtIHXEiaJf1EGI/SKFk+2fK7LaVqIJqDdECHOxz7e8MtdapREVkSvxV
apaHsWlWwZ0dibcReFR0CQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAEUmUcREcOiU+NtRS+HVTGqMoWfoiOQ4aVPaADeRFUp4
K+zKNM52f8wc8lFE6qlB2zDJICwzauZZGql3PuUuBss=
-----END CERTIFICATE-----
`, nil)

	// Test: correct apikey but api is disabled
	oneTest(t, service.DownloadKeyViaUrl, nil, &apiKey, "test-f", "", output.JsonErrUnauthorized)
}
