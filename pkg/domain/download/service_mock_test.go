package download

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// fake app to create outputter
type fakeOutputterApp struct {
	logger *zap.SugaredLogger
}

func (fa *fakeOutputterApp) GetLogger() *zap.SugaredLogger {
	return fa.logger
}

func makeFakeOutputterApp(l *zap.SugaredLogger) *fakeOutputterApp {
	return &fakeOutputterApp{
		logger: l,
	}
}

// fake storage returns some dummy data for the handlers to work with
type fakeStorage struct {
}

func (fs *fakeStorage) GetOneKeyByName(name string) (private_keys.Key, error) {
	// just get the key from the same name cert
	c, err := fs.GetCertNewestValidOrderByName(name)
	if err != nil {
		return private_keys.Key{}, err
	}
	return *c.FinalizedKey, nil
}

func (fs *fakeStorage) GetCertNewestValidOrderByName(certName string) (order orders.Order, err error) {
	if certName == "test-a" {
		pem := `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBAMLwirxhhFBmtzbKk0+m+MBRBUPcj1CrDmNmvVlkmTTKCzY1RNVk
tOpgN6szMRlX1VRb+v5j7lJ5r2gJZrDNs8kCAwEAAQJACBwGRCSbuCszD1DJZLSM
f+ue7XNydCekN0G3OiMeNdI92AUYEb+Yh8meJIYGob8wcAYCt3pp/WhhoM8Qw8kf
BQIhAMaX+Dhswwehcf2hhO+eS0KNdB8i8demjJGLap+W/eZbAiEA+0osY3/LH24Z
xWboFT6ISJuriZZK24AqbeiS/IsYj6sCIQC+mAEInhE7FI2i/k3n7kKKd9l3PIFg
Fx6XXHcS/MVmOwIhAJi2lwtQ2oybSKYix+BBRGl70V+oKo4C8cYhlVJM5fxJAiBu
N4JkNXHxXM7m8/ItFqWJtKH2DCTDl5SSt64qUnEEbw==
-----END RSA PRIVATE KEY-----`,
				ApiKey:    "k-123",
				ApiKeyNew: "",
			},
			Certificate: certificates.Certificate{
				Name:      "test-a",
				ApiKey:    "c-abc",
				ApiKeyNew: "",
			}}, nil
	}

	if certName == "test-b" {
		pem := `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALfdFbvTuf1r6Mk80ZLTfInivcfu9hF/JcRLnV+EOd4Z4/28zGBD
IwlnYkWDD7gBhuBRJoUPLnyXJ7Rp84+bp10CAwEAAQJAKJHNK6Bse+VdXFgB4zys
kG06VH0fCR3N2soXe728mguq9D3E3PyyFW/OyLUwWgXI3JXFC0+anu7oehFcE3o1
aQIhAOkYh4WJiyP7eBPcuuRNaZUweBsmZMkoW80B3W/RsGn5AiEAye4hhcuWPVuu
CjicvivY0I/y7tJ2nY/vXYfG1JqHoYUCIQDGWMghOpw6vyY7iI1D7heVCsx5Fd+X
SI9tUFP0bbM3SQIgFZuy4KNhh11ZKWTXeQ4uHFtbDq1c3g15+tM9tqB2pRUCIBT8
5URzNI/wCwqQD6D98UNKRJhD4MrDQlBBA9PYqnab
-----END RSA PRIVATE KEY-----`,
				ApiKey:    "",
				ApiKeyNew: "k-123-new",
			},
			Certificate: certificates.Certificate{
				Name:      "test-b",
				ApiKey:    "",
				ApiKeyNew: "c-abc-new",
			}}, nil
	}

	if certName == "test-c" {
		pem := `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAN1VxmTOxWUcQjr7MNUxwkyJT5TuTQJU734/b9f8wKBVy87ikjFe
UGJLzYRJYutBwJBztdXnlhOS/bXRBs1szUsCAwEAAQJACD2RTV+FaeZLcPa5MrbP
jRnvpJPauiN/Zyvldh0q7s0xMQEZHVRmYUpsXZM4fFmSUvq3npBFptA3gNzOv8Hs
AQIhAOFkEtCJ3TyXRb0/pdyY8wQijWaRvOgGgYvKLjy7YPvBAiEA+2SyKDgh15pB
5yPyoGLg68tglMwm4VjVMFaeoiRw/AsCIQCIQVpKbX2senqzfL3FTUVkU4sN3b7I
ud4o5vHqzxBDQQIhANggRwZK09V/Gf90qUf4GjS9wYfLR/XeoFIRdgoh2DznAiEA
j2MbNxUDruMAHCN7tfis2cinNPVYxPGhPxrNWdikuoY=
-----END RSA PRIVATE KEY-----`,
				ApiKey:    "k-123",
				ApiKeyNew: "k-123-new",
			},
			Certificate: certificates.Certificate{
				Name:      "test-c",
				ApiKey:    "c-abc",
				ApiKeyNew: "c-abc-new",
			}}, nil
	}

	if certName == "test-d" {
		pem := `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgIBvDANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1kMB4XDTI2MDUwMTIyMTQxN1oXDTM2MDUwMTIyMTQxN1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
vp1vdm5BzjK9qEKtAdHs9g1LJySrIXPOSviR3I7z9noO0FahmlPt3AnnjzeRb0hJ
GDaGi7muiG8It83B2wZwIQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAAGpYHMJ07rbpJBnv1yPWJ53mOIgjVGiZXgNHBR4zJeg
smEvoom3lC7c5bSutZF6V/bWO23t0hGsyJw4BxXtW5M=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBLDANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZDAeFw0yNjA1MDEyMjE0MTdaFw0zNjA1MDEyMjE0MTdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
vp1vdm5BzjK9qEKtAdHs9g1LJySrIXPOSviR3I7z9noO0FahmlPt3AnnjzeRb0hJ
GDaGi7muiG8It83B2wZwIQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUvvCFAbKpAOuar8f5+IScb4NiWxswDQYJKoZI
hvcNAQELBQADQQAO+Aguu5ZMCH4fg/EQsljnl3kOq5Sgk009fBg0eJ5x4HEIEMLv
9q0+nXdO0m5WIArtZbKvfrVcbLaPsJjIOoRN
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBDANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZDAeFw0yNjA1MDEyMjE0MTdaFw0zNjA1MDEyMjE0MTdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1kMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKTw6Q8Y/ALr
GTkTRas73UuknvNr+sZO5BtiuNu+XQBuHXUOvGOnca80KvCI+UNj+QHSeiqy5jEV
1y8ko316Kw0CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFNfAVAINdRIo/WHyWX9CIkL9225SMA0GCSqGSIb3DQEBCwUA
A0EAM3AWlFRDPeKM4k8QC6xRLUdVBiutj1FYzwq5WkqBh5f+vm2NNsv6v7l43KgT
G/zUBlah2C64Lprd+EZW6xMQ7Q==
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBOQIBAAJBAJ34t82UR4zY3/G0VEVe0iNSTtGmqkWh+amRzRUyxU7aeeTj3mjF
az0gEsw0oiWiWopNSZlszw8MAnxqWsd8HScCAwEAAQJAPXfTJWZGNRMKiMVvaRLN
V7smOkPMy42MVSQLle0Mg3K5J8TmTWvC73axhC4bfZjgiTyN3sWEispogBXgObbA
GQIhAMx6OIPgJM+yCyagxZ7s0wdoAzMTTnU5JqOUW5gf3Y6dAiEAxcak/QJB0UTQ
TrEia3AZmcQVjcMCmbgxDOyr9xd2TZMCIBqYWvlsEd2hvqmLh6igDOKNuLzP6gh9
InVsOm2S13JRAiBhRyqh07lh4FIBUrkWVUYSTtM3LiMaTvG5ZLPUznJ/BwIgHuCc
vtAec+Tog6PclxqDd+Lvj9Q7cfqOhfAD2bvv+xM=
-----END RSA PRIVATE KEY-----`,
				ApiKey:    "",
				ApiKeyNew: "",
			},
			Certificate: certificates.Certificate{
				Name:      "test-d",
				ApiKey:    "",
				ApiKeyNew: "",
			}}, nil
	}

	if certName == "test-e" {
		pem := `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAL8nKLmviCA2UqLqEwTJ8mPnuN5k1iKw4W1RlxYYtTpvyDDkz5v6
C25MeHrbJqj0tzsBwXNZI2OHrBcSl3qWum0CAwEAAQJAKE4uu2Z8dsRVjNX74CLf
2fp6g+QxtbtjsQwG5kkb73/Um0phhoLDtMzgNg+MncKKgjx2WmVkX45LTP/TN8Pv
iQIhANbGnhTb6PD+ja1lTHuoxb16cKe3YjWRzu5QcqWwziE5AiEA49fKUxVQ5+gn
vYeAzr1zmAnmlvSpgEYiFIx8ENN7xtUCICN0qHYjE6JtM3BTj7u+Ud6EzwIw+OqF
Bpc6+qI1vOGpAiAZg7HBihKMVcAVhlYTUL3gGcO7xdwxZCku2eiOzc//nQIhALuF
N/VC8wvORy4lKBkAeRy0oK+9o6R5mdzP3Fjpz9I4
-----END RSA PRIVATE KEY-----`,
				ApiKey:       "k-123",
				ApiKeyNew:    "k-123-new",
				ApiKeyViaUrl: true,
			},
			Certificate: certificates.Certificate{
				Name:         "test-e",
				ApiKey:       "c-abc",
				ApiKeyNew:    "c-abc-new",
				ApiKeyViaUrl: true,
			}}, nil
	}

	if certName == "test-f" {
		pem := `-----BEGIN CERTIFICATE-----
MIIBPDCB56ADAgECAgICmjANBgkqhkiG9w0BAQsFADAeMRwwGgYDVQQKExNJbnRl
cm1lZGlhdGUtVGVzdC1mMB4XDTI2MDUwMTIyMTUxN1oXDTM2MDUwMTIyMTUxN1ow
FjEUMBIGA1UEChMLTGVhZi1UZXN0LWYwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
0dmojebsU+/+a1uvbCy34bjCWucsqsnmuNr/rKyCMqSlRx7uRl2wkvJR+guA8w3V
kNvu6JtnfzwtQZmejKBqRQIDAQABoxcwFTATBgNVHSUEDDAKBggrBgEFBQcDATAN
BgkqhkiG9w0BAQsFAANBAK9Y9EcGQ+DRPwBWNzUQ0zQxP444cXxhzeh//Lh9UaS0
UVf19L8d5L1ZJAlknU8sn1Qx5aMrVqMbpWGZrMKdVp8=
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBZzCCARGgAwIBAgIBQjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZjAeFw0yNjA1MDEyMjE1MTdaFw0zNjA1MDEyMjE1MTdaMB4xHDAaBgNV
BAoTE0ludGVybWVkaWF0ZS1UZXN0LWYwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA
0dmojebsU+/+a1uvbCy34bjCWucsqsnmuNr/rKyCMqSlRx7uRl2wkvJR+guA8w3V
kNvu6JtnfzwtQZmejKBqRQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUMiS73t8F2HgZz70jFcWud0s6bhAwDQYJKoZI
hvcNAQELBQADQQCMn02i1GKmFrJtsVitBvwcRvbkCZ1Jdogxju/Hkx8WhyFzI//9
i/XIxS3WsDlwWfFVkadToNsTvaRcHJBO9WuH
-----END CERTIFICATE-----
 -----BEGIN CERTIFICATE-----
MIIBXzCCAQmgAwIBAgIBBjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtSb290
LVRlc3QtZjAeFw0yNjA1MDEyMjE1MTdaFw0zNjA1MDEyMjE1MTdaMBYxFDASBgNV
BAoTC1Jvb3QtVGVzdC1mMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAJ2F6RYe5VMh
cT/Vk3v+nab2flklZO8O7n+D8xifGna8lexKfJ6+pNbnmddA3Ekx8vS/QYjZ7Q8E
E/JHgFwXj70CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFINOrn3MmHKQROTZg3FtvVAHOQtlMA0GCSqGSIb3DQEBCwUA
A0EAWtSaoEHPq0cseQUod1Bk/vnbidNcexOL6cqTOUivN+k4X+4ig/ClDT/fDm2x
7qe+skSjMhpMSiOiMyw5PgvZ9w==
-----END CERTIFICATE-----`
		return orders.Order{
			Pem: &pem,
			FinalizedKey: &private_keys.Key{
				Pem: `-----BEGIN RSA PRIVATE KEY-----
MIIBOQIBAAJBANbzjgZ/CyBVKbqWxUZJnALVZ0wtLW+Raeaz8a5PriWjYERQMfio
97a9HzlmuWtnWEs90bIS0/SBpigTdHUS3acCAwEAAQJAMTlVLOerBJx8Ed61DoOM
1plILomp/gKu3cYXcnOMzdFQd7h+2H3ENUhhFnwggEUiV4iWRhDwqfl1JwQ6yRVv
AQIhANzVdJKvhNTOQcsLLiDH9fV5fMYc+SrdkKg/LAD1DZJNAiEA+S5LrkePuTxT
pYfJZe/Dw32CCbHoJaDd5vnwJ2aiocMCIAM28D16ZJqcbgTAoulDP+dU32Ya4d2n
4AUy9jcFWi85AiBfA/Y7yHHXcld7ASIcyqZdPth9FeetoX+7+YZHn+1XvQIgQZF1
yVylfB4NSVwqtECZIcO+U/4KupIGF7kvKs7lxjw=
-----END RSA PRIVATE KEY-----`,
				ApiKey:         "k-123",
				ApiKeyNew:      "k-123-new",
				ApiKeyViaUrl:   true,
				ApiKeyDisabled: true,
			},
			Certificate: certificates.Certificate{
				Name:         "test-f",
				ApiKey:       "c-abc",
				ApiKeyNew:    "c-abc-new",
				ApiKeyViaUrl: true,
			}}, nil
	}

	return orders.Order{}, storage.ErrNoRecord
}
func (fs *fakeStorage) PutKeyLastAccess(keyId int, unixLastAccessTime int64) (err error) {
	return errors.New("not implemented")
}
func (fs *fakeStorage) PutCertLastAccess(certId int, unixLastAccessTime int64) (err error) {
	return errors.New("not implemented")
}

// fake app for this package
type fakeApp struct {
	logger    *zap.SugaredLogger
	outputter *output.Service
	storage   Storage
}

func (fa *fakeApp) GetLogger() *zap.SugaredLogger {
	return fa.logger
}

func (fa *fakeApp) GetOutputter() *output.Service {
	return fa.outputter
}

func (fa *fakeApp) GetDownloadStorage() Storage {
	return fa.storage
}

// makeFakeApp returns a *fakeApp that can be used to create a new
// Service specific for testing
func makeFakeApp(t *testing.T) *fakeApp {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)).Sugar() // use fatal to avoid log output
	outputService, err := output.NewService(makeFakeOutputterApp(logger))
	if err != nil {
		t.Fatal(err)
	}

	return &fakeApp{
		logger:    logger,
		outputter: outputService,
		storage:   &fakeStorage{},
	}
}

// function name reflection
func getFunctionName(f interface{}) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()), ".")
	return strs[len(strs)-1]
}

// function to friendly format string pointers
func stringPointerToVal(sp *string) string {
	if sp != nil {
		return *sp
	}
	return "<nil>"
}

// function to run one test
func oneTest(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) *output.JsonError,
	apiKeyHeader *string, apiKeyURL *string, certName string, expectedBody string, expectedJsonErr *output.JsonError) {
	r, err := http.NewRequest("GET", "/certwarden/api/v1/download/certificates", nil)
	if err != nil {
		t.Fatal(err)
	}
	// set cert name like the router would
	ctx := r.Context()
	ps := httprouter.Params{{Key: "name", Value: certName}}
	// add api key to url, if specified
	if apiKeyURL != nil {
		ps = append(ps, httprouter.Param{Key: "apiKey", Value: *apiKeyURL})
	}
	ctx = context.WithValue(ctx, httprouter.ParamsKey, ps)
	r = r.WithContext(ctx)

	// add api key header, if specified
	if apiKeyHeader != nil {
		r.Header.Add("x-api-key", *apiKeyHeader)
	}

	// run the test and check the result
	w := httptest.NewRecorder()
	jsonErr := handler(w, r)

	if !errors.Is(jsonErr, expectedJsonErr) {
		t.Errorf("%s: name '%s' with header api-key '%s' and url api-key '%s' returned error '%s' but expected '%s'", getFunctionName(handler),
			certName, stringPointerToVal(apiKeyHeader), stringPointerToVal(apiKeyURL), jsonErr, expectedJsonErr)
	}

	body := w.Body.String()
	if body != expectedBody {
		t.Errorf("%s: name '%s' with header api-key '%s' and url api-key '%s' returned body '%s' but expected body '%s'", getFunctionName(handler),
			certName, stringPointerToVal(apiKeyHeader), stringPointerToVal(apiKeyURL), body, expectedBody)
	}
}
