//go:debug x509negativeserial=1
package acme_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/test_helpers"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"testing"
)

// TestACMERenewalInfoIdentifier tests the ARI ID generation function. expected bytes should
// directly be copied from the der certificate (excluding tag and length bytes)
func TestACMERenewalInfoIdentifier(t *testing.T) {
	testCases := []struct {
		pem                 string
		expectedAKIBytes    []byte
		expectedSerialBytes []byte
		expectedErr         error
	}{
		// case: "regular" low first bit values for aki and serial
		{`-----BEGIN CERTIFICATE-----
MIIEQjCCAiqgAwIBAgIQLb8uNBeXJtkCvEmTx49JbDANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTI2MDgwMzIwNTE1M1oXDTI2MDkwMjIwNTE1
M1owEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBAOtoZmSC6k1uukqOol1YRgCRibb2FAjCdnYWszUNPcUszRC61lxBP3qF
QJSe/jp1BvXBkct3DV4qUmPFiXJQLFEFla1D7PaV5fdIZFAJym1nX7fS4xpV2/Ac
3WsKYAtqZP+1yEh0oPdBuHeJPXqEGDzpIYCW3kztjbKmJqLG/4+dUdJa35tCgamb
HM0WN8FsmN52fkKtNREQYEbDWz3o/km1AVVG2cSGbWeLvBwrONedsNGneQQhDzfE
5gt5Ds47/uHi5xIircrH34QpVuet+tUno1cEeQJL2xgqT/EyB7P+T9M5grPdQo4k
cXcqLfvymTup1C2qB7EQcd2cWxDIaBMCAwEAAaOBkzCBkDAOBgNVHQ8BAf8EBAMC
AoQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQU
eBfrLcscDSGN5SQy4Xg64dtKDcwwHwYDVR0jBBgwFoAUJCgXtTPEHSmWxmdGgRXf
fwtrBqYwGwYDVR0RBBQwEoIQZmFrZS5leGFtcGxlLmNvbTANBgkqhkiG9w0BAQsF
AAOCAgEAbF8dpyymM1R0/oE0V4eVEaUj+d5oO0SrDV3vjiIC7Dc1IPzGkcBDCPcJ
ciZ7EaxAsCVnZ7FS4jtx74+jHRYadG6csjhUlxDDyT1k60CPGMiAu8vI0SgdsP/l
n7O92oIyixrmRUU/j8qKcl8B0TQTTxCLmF/Pxoe7i+JMqIuOtAwyPF4TH0b5oqvd
gLfgQpiEmO1ZCdwQQE8hdCUD8cn+sPFiOYNcv1SEwAVp/pG42csWgQY8fNkxkm/K
x/fPMSTHvd5kkS+Nv+d5w05/cYHsnjOJG7+RxRPSqatRa8LoTeJWgLPDDSeiaCIU
8I2JtCAgB61473DQSUAjGbFL3A71VWaNDSWxYMJtiU5J4kNB6ejRhdX1jQHe9lvB
1uF+Mjm07Z0rCXo+vEJBcLzfTt2WSJYifmV73PJk42pcC3dgW5WuCVbEzA7fi1Wu
xs1ExC/W5XbORBw2M899nST5+GN11NXhI6ItIQTktGy5WpL+tHhG+Q5MQmnAg1Ma
u5snPoG/IEXVvcAIq7cP8LMmS+AAtIkB2W2rjp5bPR6inBvrF1ZxGU+iSW5Setp9
7JcpRUz2z/JDPIc6K88HY+X9UMmH20fsROYX6BdhhxGRIupv/e64Y/db/xrlSjvf
2++EBULK9Q476RbnWMJvYZ/rOj+ZiAdb3lOhHG1mEwxuY8na3lo=
-----END CERTIFICATE-----`,
			[]byte{0x24, 0x28, 0x17, 0xB5, 0x33, 0xC4, 0x1D, 0x29, 0x96, 0xC6, 0x67, 0x46, 0x81, 0x15, 0xDF, 0x7F, 0x0B, 0x6B, 0x06, 0xA6},
			[]byte{0x2D, 0xBF, 0x2E, 0x34, 0x17, 0x97, 0x26, 0xD9, 0x02, 0xBC, 0x49, 0x93, 0xC7, 0x8F, 0x49, 0x6C},
			nil,
		},

		// case: serial number absolute value 1st bit is high
		{`-----BEGIN CERTIFICATE-----
MIIEQzCCAiugAwIBAgIRAKr7ULlA5o05qDzkx+7j3KcwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0yNjA4MDMyMDE5MTFaFw0yNjA5MDIyMDE5
MTFaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCYuEHxb7sYX4NLVFYOW0nNi75+NfylutsMt7lu1MaWH+64wlxH+nAf
6cEkIPXW9CFi7rGwvLEmAhfyQdqNseO7cQ462Xo+/cB/0BvnT/39K4LoLkC2u31s
hr/lV+KwuSYOvfJuvaDWS14ds+1ngIAILektt4a6tBerrBO/7eVlk8ja8JKeo434
Q7xqSu4xBtTCjaYqgkW3PjvArFHb/scmUUzZcLf40EnktvaFJk8fLy0rqukEABi9
AiWV8smK6zWPBoumO5/flntldV6qx1RKtSXpumokv82/ciVbllvQAyTsV3LIzWqF
TToy/faIdrZva9SLkFOz/JDfGhrWFVMbAgMBAAGjgZMwgZAwDgYDVR0PAQH/BAQD
AgKEMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYE
FB2D/YdRv3k2wx3DiOyYy9mOuUdWMB8GA1UdIwQYMBaAFMgGObUJJYM2Z1JiamDc
sswKaYLCMBsGA1UdEQQUMBKCEGZha2UuZXhhbXBsZS5jb20wDQYJKoZIhvcNAQEL
BQADggIBAC6LRBMmk3hhQA9cb882s1ENXZCe5vK0JebslcgpmfEOeI6r3GI2oTWr
1CVRsgt5b0K05tJo0NNLP96kjWMuZPBB513IgCKb2cxlennaS2A4ZIkzaC5UY7aq
3javhLzO338yrwiPyugbd8MirFA0vx0qE1PfidFPnXpHpQKD7kUSJyiOsK1c0Xgt
VnGLYhenLxsrN2A6BFhBJJYK3dITX+ZQoIctuaqFlT/fyeRIhHr1jjPwmmdyjmPU
idRh7SROvsrErVji4DG2hOeBaRuZ5lScwBdIdVIxCQBQ8mM3YWwCtQC5y8n59HFS
lDC2HnTjtlnjfgbaCqIZE73H0S2syF+j6lm5Y3C3mGexDN5+0K3xTUrCeT5bQlCV
IROeccH7LWDCc9Lv/S8u2v1Efo2zCfiDqRksS9V5iWdOCRELGthLI3PF/FeDsQ3L
qMGtZhiFMHdyIIzHCIb/yIo3olxV6BfUvMzHqRW0YhldBGLYzVVj2DdnQz4ryobt
BhUi28eO54dH8cD+pA3Ad+w9N0kIbPCtC/EkaU0NLUM6Ay4lNj5yKkE+gVs+5Yos
GOfzAQ9GNRQq3RpNsoJ8VH6hzm5egIlslRk/xV44T3EmI7I6NAZIKjjbdvz8tVAf
CgA+FIWvGNAxAZbisS0a2KOZHFX0RNE+EQouLd1Z6CvS3X92PfPA
-----END CERTIFICATE-----`,
			[]byte{0xC8, 0x06, 0x39, 0xB5, 0x09, 0x25, 0x83, 0x36, 0x67, 0x52, 0x62, 0x6A, 0x60, 0xDC, 0xB2, 0xCC, 0x0A, 0x69, 0x82, 0xC2},
			[]byte{0x00, 0xAA, 0xFB, 0x50, 0xB9, 0x40, 0xE6, 0x8D, 0x39, 0xA8, 0x3C, 0xE4, 0xC7, 0xEE, 0xE3, 0xDC, 0xA7},
			nil,
		},

		// case: serial and aki both have 1st bit high
		{`-----BEGIN CERTIFICATE-----
MIIEQzCCAiugAwIBAgIRAK/KcpoTb7q76DwUm6VmR/0wDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0yNjA4MDMyMDQzNTNaFw0yNjA5MDIyMDQz
NTNaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCflC95WD3hvKLQInlFb9TZ/Gjd2+q7hGjclW0QjPkbpqFgBHiICwiU
JtHESAyWAAQVooPcTIBDKRdWqLicQJfnuwe0UAHzDT9/IpBqX8D3MduMV68xGaVh
aaHeMdiJQXND07H4kjYudOfOWY1/HdDQO/QkbU1nsOi4AIKBBXJf8ObkH3iNoN/Z
/K6K9/yUQjyYdByyuzR9pgW5kJ+EaJ49TWecgdTDuvMH0YNV1OYwfCRlCF4YSPIg
t3Bh0Dxizd/IMtFfaK5orNoPXtJ+/u8UX6I2TsXp9Fv8BCud2Iedt5RbZc46YwMs
kG3ltbW04cD9HeW/CGs4IGKyYPgFBYhDAgMBAAGjgZMwgZAwDgYDVR0PAQH/BAQD
AgKEMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYE
FNUqXdKqXMwTpAYh2Tou3v2WORC/MB8GA1UdIwQYMBaAFMSQZBIARULZmfLHcTPm
hfrmQqn8MBsGA1UdEQQUMBKCEGZha2UuZXhhbXBsZS5jb20wDQYJKoZIhvcNAQEL
BQADggIBAA1bAHRqg6r4PVPb2Bzw1ijsuHBmFEssjiNzw0U5fUrSdsZJJPfzYUaZ
5hynLJwHASIt319Jjsc/TWDa2/5PQZhZ6NqYCKTNiLFB+fWRgvimGgIy4D5jFhKm
vJuYm+s7vSBFaVggBu4sq+DGr6ljBw34Q4UuTg1gH9/5OjWHV1W82gLPm1XWm2cN
z+GAljSkIrQppEZ8kjO0y2ibkXPJIjdxGnztItF5TYmGTLYpX+JUL6Pfhr+4LRun
qHboDxIhAGxiFz+eVkTifk4sg7mBwlanElXFx/Zzbwse/XvWiaC/msXro5mw2m+D
q3Vfa2R7CVFcNo53U6GZSAkpvXuRJDVsJ/0uSA/blMTIeJEJgQkzPg9Q/Pmy3tQz
Gna3rRbHAialX8h9ETPKnqj6RcZgNTbWmBAkny2Qrc4CVDxyinV6MvuVPJZ3NDMl
KwsPAeNTIzJfxzP1nyinxFENZPJuFSfZW5w9hpESNbtueZh1OP30ukqM/T6ZYpDv
0C4Ep81XJTRdvClapOeCd1JnnazgqPwOZPIn2l4gnfRw/HwiiNwMkrAV8aBUt/Ux
h9BxcyRWQXfsiW81NOOeYaIZb0bJqrxiLt+QuqGReeB1I/AFHpO6YWFXJ2cvSMoc
zjSzJ3FWawet91PtfTR1+DYtOyoS+SYxkNNECrSla7+bHvqGgkup
-----END CERTIFICATE-----`,
			[]byte{0xC4, 0x90, 0x64, 0x12, 0x00, 0x45, 0x42, 0xD9, 0x99, 0xF2, 0xC7, 0x71, 0x33, 0xE6, 0x85, 0xFA, 0xE6, 0x42, 0xA9, 0xFC},
			[]byte{0x00, 0xAF, 0xCA, 0x72, 0x9A, 0x13, 0x6F, 0xBA, 0xBB, 0xE8, 0x3C, 0x14, 0x9B, 0xA5, 0x66, 0x47, 0xFD},
			nil,
		},

		// aki leading 0 byte
		{`-----BEGIN CERTIFICATE-----
MIIEQjCCAiqgAwIBAgIQPRr0VCYKE5VWKPpJnsepazANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTI2MDgwMzIwNDgwMloXDTI2MDkwMjIwNDgw
MlowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBALM/PHevlOdfnLjwC4PiZOkndbNlUtaU6ACFRdxURgRgBJhMZa9zJGL8
tGFuTs9C9BpDE7iEpBjLfGMPsNc5jr+yZvjelEPTH03d2TvZOmNCJs9mXFy5uhd7
ZRRKOFGgIaUPd8EPHEsHBDUfxwJyyoDAzzhJi6mxIW323Shqsm+GrYhVQ3NIwmXf
sFC5OsFTPpqHFQ6PA4MyBGR2UEjLYhgf34DRjXj27/FIlJPvzUrCU7qZRVpWWa7W
otemcg3/ixXT6TEXaVvjsDJM653Y/YhDYlLAB5k+w+6LE/yoF6/pvUoLqRuWgRHT
u48nlRX8wMu/QYEv3XqeH5dkNvRNapUCAwEAAaOBkzCBkDAOBgNVHQ8BAf8EBAMC
AoQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQU
8gliLRH4MNJx2HPh4tmU6cInP0IwHwYDVR0jBBgwFoAUAKJ5CUkJdM8ToVWz8fJM
UXKpMy8wGwYDVR0RBBQwEoIQZmFrZS5leGFtcGxlLmNvbTANBgkqhkiG9w0BAQsF
AAOCAgEAehXrv6zIg10qASwuPRIhaaiB9tKO71qMlTAa5bfDLf5B0YLc/FeB12eY
Bh07jItYCI70619bUGgeUNEei6KAgU4q8ClzyLB4gHPOS3605hxDq6NXPkmccoj0
Us7Wo4dUL43mtd2LoBxoBEvU7NtCfgSy00oLtv/0s39SFeFmi7/Lt7bMbj73bvuH
1oO7VvtJ1qtq4cqYKgcSUSmXD9nxI9C23QXfnGg+q48Op7GVa6ZFl/pGvstSR5gx
jkDbMV66hWX7111+9inYdMQUxf8bcunSfa+IUGGJmgHpvhhECBCqCnG/0nqguXex
68VneaxXly+YRz0OfJyRV32Fdtvt9H/wzdi8zvmxKvs+Gd8dXSwv+gxjaSp0+6KS
E7tvKUTDkrcPRu9jyWvwaNsPOqSGasoxHUB3xCr7we44hHn7/29akAuSBEH0guWQ
izogL7KPwrLDV6daZoChBx4StVo7kLm7mmseGXMGvOGF1Iu6O0hGQTwTyNfrX1lf
ccpJwOUr647iSocoVkC4xlvOSlf2Fg7585/pIEBVDqkJQl8i3gRW912PKBMsAyQS
d5d21KeCBbxCesuITAbDocL+pA0PGn6tKFvrVEvH/dkZtabni1EklTWTuplQ30Vq
Fl/0BhMOhYkKLBXIxRwPDsclulMsQM8gyuZRnLjMz1ft9TUBEqo=
-----END CERTIFICATE-----`,
			[]byte{0x00, 0xA2, 0x79, 0x09, 0x49, 0x09, 0x74, 0xCF, 0x13, 0xA1, 0x55, 0xB3, 0xF1, 0xF2, 0x4C, 0x51, 0x72, 0xA9, 0x33, 0x2F},
			[]byte{0x3D, 0x1A, 0xF4, 0x54, 0x26, 0x0A, 0x13, 0x95, 0x56, 0x28, 0xFA, 0x49, 0x9E, 0xC7, 0xA9, 0x6B},
			nil,
		},

		// case: serial number is 0
		{`-----BEGIN CERTIFICATE-----
MIIEMzCCAhugAwIBAgIBADANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdBY21l
IENvMB4XDTI2MDgwMzIxMDgyMVoXDTI2MDkwMjIxMDgyMVowEjEQMA4GA1UEChMH
QWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMBsWoEbss2g
+Vbv4e5YD5/PshhAXeSDBcD/gPZ/fgC9YxfUd5Dp30Dra18rxDNVIffSQDFDTqya
8fufwVNftJ6zRdzRZyoz05UT30pymZUzXO7HoOhTElHfsi1bQjzZR4dm9rD6eU3d
Th7LIJNsH5M9pO8MP+fZjRdiTHbB5aewBVYOraQjeXtsXfS2w7rQNUpl6OYss8E3
e8Rz0a6KOT6bp+zewYvykYtkMyqzsXdbRUWfJ8j4izgng0B7VAbIz1TMAK5YNW4P
WmLReQ/B1jyVDDo5EF+rLn3+WqzG0DGkcMx5Unb/6i1FTdCJMYjsSt7C/xILfS8j
0SZhMfIt0YMCAwEAAaOBkzCBkDAOBgNVHQ8BAf8EBAMCAoQwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUr3DYQbuam6j4G/LO7yER
4yoo+RcwHwYDVR0jBBgwFoAUPCR8fAVDmyeacJN3Fs23cM9hBJowGwYDVR0RBBQw
EoIQZmFrZS5leGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAgEAjvO3f5RcfidY
Q6ugdTwDezMXVjbAf2e48lASkfV8CFjaBuPNcpcdamVC/r0S+yUkgh9Gsrj69JSs
/x4uKSaRK0hVb0eQnZbR3scDfYP+k/DTqb5po8gcxG0oJlpcRWljPK18f0iV4ViK
m6z6okrbzc8MEIQkawn3oJLf1qDJYsa4g6l5E+bZWgwRjIpX+n9U/ebzzEim8dgY
LApQW+Qz/AF9y0cdAmf5FlV6x2r5HnwPqqH+58cj0Q/ThiI1SseZ0XchkGLpK5Sn
Aw2JuiaVfUJqZ4G70W9k8kGlk6EXtXQE2uPNX4FHWglxF3uHH14q7FRUnTyqBZ3R
yNFKDwSgGWy9sc4XB7WQW5BqSeRaJjn9aIAqtQfcdP7BHTOue5jsBNZtbsIgP1II
/9KUvMUY7BpfnDu64K22QaSU6NfDR4k5P+IvaFOQSt8hzMg+qhWTus15yI5wPpTa
Jwt4h+SKTetATmokxZUej/VdnbnpdyoAppc69SgdYYhrlnHjcAIf5pmCiA1xe/Ol
Q09erZ5CJHgf13/q1lJZDLSSasJmRe75AUdb0shi4lrqpzFj+DbAVG8AhBegQ7/f
9SuxAbhvmbtPZqEzvcotIIl5AhQLea+mix721kDrGzByLAcxz6U3Ng84CdYtkhOL
DKJXPZl2cjkiNCZi/6N4qTvzkbcQpa0=
-----END CERTIFICATE-----`,
			[]byte{0x3C, 0x24, 0x7C, 0x7C, 0x05, 0x43, 0x9B, 0x27, 0x9A, 0x70, 0x93, 0x77, 0x16, 0xCD, 0xB7, 0x70, 0xCF, 0x61, 0x04, 0x9A},
			[]byte{0x00},
			nil,
		},

		// case: no aki present
		{`-----BEGIN CERTIFICATE-----
MIIEHzCCAgegAwIBAgIQWOVxeZxI8lMAnxq+bNxMnjANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTI2MDgwMzIwNTU0NFoXDTI2MDkwMjIwNTU0
NFowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBALu9ZfyxA30jWWiDZ9NCVI9Z/dLFs91bcrHz7rcbeRTY/6g7W/oFa4p/
rrh2iqpt03zV7WwOQkVE3rVaiBGyuJuUgVX8kZ3rQ6YCM6SuGlxTSJWvzj3Ev4wZ
lBxVSMrOhg2saTLebn4XESrNtCa0tkS7Jy3zlrCc4WkLWD+zGyABLaNmoSxVmx4g
LIJ1ytL2Sfrk7RJXTfIR9vgnEX2oRUu4HiTCXqjLK9rsX2i6uIrbiT7IE4C8+K+l
hi2OjuV1IwerSCtyn2ITWUmqbJEYvE3SDbZwaG+VKVNbFdP4p5/Wkb/+xNTTXi1i
qR8gulpNFqRe0qolYENgRBOJI+Q/5F0CAwEAAaNxMG8wDgYDVR0PAQH/BAQDAgKE
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDX5
bah/T3tu05n4jgCd9FaoyEzfMBsGA1UdEQQUMBKCEGZha2UuZXhhbXBsZS5jb20w
DQYJKoZIhvcNAQELBQADggIBAI5ywMQr9rEIfhyWXWjIxa4KqXHjXf6aMhpf8k/v
ZAd6IqtrrS2dJOoDrKn7uQaKzODFVoHw9PjdBJgHjrLqiuKG7iVzWJ+Xax09fN2c
hJYNlF7t7hPtskU1FUd9D/8IcvDcd+pQNS2eByxHtD1kWYAenPbNF5CRtWyFb6u5
Zq/rOlwVs7uSW6yrKdQKChv+mdWPDI1vorzJ/mXWrerIfkwYjrtpJX1/Ul9rFPB9
Zx3AR47WdpFenUS8dzK/zlHXPa3u9cDBO6DSh7oBMK+c1daKIXTGl8Y13QwBp08y
VAiEZ2l+MsCJ+J2HiPQxVkD1/ABud5W3UqeQZQJOxlVhMvPP5xkE2ziNQE6SDFqr
KdV2cJYHzdfVdFtgGffsMWFeSrBRmfEuoX/e196CKZpoSfBzgSz/5I0EQUa6ipr0
UguIxcYVwbBksthsxqtnGzCvzXkUzARsvH8fJL7xl+qqIsHF7lX/rOqQ18xt3TjW
T8R+b8ZFjHDgF91M4EyTHHBT+pJjp7tV6VH+RY8NSFXJztZPxI2WM8+EIRNyf+QX
Bo72TequoRph1e0yP2L25am9wqCB3/dlfVlpI12E2WhsV5MbcEzreZniXaGG2diZ
XgrMzu2wy8Nk1l3nvYB0HQbRpiRrNqrWFUzDGoq0XRCccZChNykSP2YPfLjV+KRO
phhj
-----END CERTIFICATE-----`,
			[]byte{},
			[]byte{},
			acme.ErrARIIdentifierFailed,
		},

		// case: no skid present
		// 		{`-----BEGIN CERTIFICATE-----
		// MIIEDzCCAfegAwIBAjANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdBY21l
		// IENvMB4XDTI2MDgwMzIxNDQyNloXDTI2MDkwMjIxNDQyNlowEjEQMA4GA1UE
		// ChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKol
		// ffWOtN0igFINj2rxkwh4jIzBz71dyAp3ucdy0J7qu66Hmx9S1/xSNPqSXUzY
		// IkIOwrdWqmqAJyYA8RNjJo08uafXn2+Ni4dz+o7JwA1IK6+8JiZGbqeoBZ4f
		// 4EZhvZAKRvpN10IwlyvhIKnpIzHXePPCy0Gq7vMqGNanCtyczlBwEThxRMPh
		// SR8k3DWT4m2/16db91pWw9RMTOhcVTW1/JNJqmQhtj6AJkUFbXlYELLKffxW
		// d6Z2sjiKb6MxVfPsppdGThw1aRB7knedhhSE6MvI67MnwOX700+EsWvXKs8X
		// 4QbqOpcNZ4pVghPtdMvHZfYH37mds0XzKf418+8CAwEAAaNzMHEwDgYDVR0P
		// AQH/BAQDAgKEMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAw
		// HwYDVR0jBBgwFoAUM9qHgpaOfmGhdeAqYQMimrjhKCIwGwYDVR0RBBQwEoIQ
		// ZmFrZS5leGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAgEARx8wOtj0iUgP
		// JNFqccikeOLB/gd7Ku5HKsnv7VwC86YKZFUjiRaZS8Vu53KVdl2UgaHc6us6
		// bueBLnc8spR0Eop4DZ0oBCOIQsQqP2r/OyiG3fmslL8ZYCwMJ/fcW1QQHHs4
		// 3ywEMUcBf19IX2does+HOcuJjC8hguBYyvcfKEFNJ7fl6Z2TgpXPRV+Mi/cs
		// msEAx97CSnz6cVOeNO+XwuGhNICi2hNCmOsDcWR9Y3Um4skkNYMyqBxf++gE
		// M7hyHrTWV5HFZKhsBSEtEdp2H+0UAa6ivOu+4fhtsLt9t7uhfePz87SHSkhs
		// 6SU9fPwqYzr+R0fVtq8e9Aem3IZfhn6kPUU2maWdZ4PfkjzrEwHVjv6MANrY
		// xaEfsqbPc/l+4xykq8mVrUmN9VaEGO0EhSsU1JPESIPvayMnjTnvdt41wtaN
		// ++hk0CZFYHp9M09YC+3gDrzI7a8AjZS09rUrTudJP+ucy5hZFDNFLTQ4TMul
		// Mh+pHTWJG3/DxiNfUk3VwXzL1a4egeAD4fhJjbCfO7fhVeYDkacuNjmmFQ6w
		// DrnERmaK4grMakxmPt2WRqWJnJEwTD2eJUhKjhtVWDgmIi4IO4QIn6ewDLNT
		// lxddAwRyhSiwSIq8cY1Vxo3Pnkq9cIZLWYvXaE/lzsEE9gFIXs65jGTjiv30
		// +B82Meb5IAM=
		// -----END CERTIFICATE-----`,
		// 			[]byte{},
		// 			[]byte{},
		// 			acme.ErrARIIdentifierFailed,
		// 		},

		// case: serial number is negative
		{`-----BEGIN CERTIFICATE-----
MIIEMzCCAhugAwIBAgIB/jANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdBY21l
IENvMB4XDTI2MDgwMzIxMDEyM1oXDTI2MDkwMjIxMDEyM1owEjEQMA4GA1UEChMH
QWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL57N0+wMu8G
MSSqLgpc3uc8ndusZRDmI3XXhlT0B8ZIg2wb3a8w1YG3sZi1bEoRe1JsGQvfxAKG
5MyJ6ARiks1mUyfvLNOrZh88V1TlFvnRegisnH0vRxNRTFXznt9EzVs+ZS8B1OCE
+8siwvQPxZOt5i6LlUzb4c8uPzmU3E2HpGHT9uXBYYkLmN8lvg2duYdUc/DPoNtC
IVIsRJ4oGnfUZ7vg/mhTdx0JUODQwpV/LrOYd9LeBRZk2d6YPtjsXWqfX1QbuHOx
5FE80FCrXciouIrZahq+jbLhD9LJd03l1z3+SuFcljHYd+IjKvBjUzEm236NZ55R
uv+zGbcP6wsCAwEAAaOBkzCBkDAOBgNVHQ8BAf8EBAMCAoQwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUgzhxutDZZmsUn5PJa+T7
YokMYrUwHwYDVR0jBBgwFoAUexQKf5THoJrQwSv/+tmvyabYEr0wGwYDVR0RBBQw
EoIQZmFrZS5leGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAgEADIsUzRU0g2tM
zZ/ogi+gSjdzqz4ulAWlgBDDTRU6iCdQy88A0m9jubNhrPAMPe3B+XyLoPoPeo/L
oeiyo6Uw5nEw3+hcsSMLz44xexejXbjTk1O50V4P+xO20dXISeiOT7vOUEvGgO0V
2pH1iX6khX4g4nxXlJUpC8MxVwyfnHb5NHIxQz7FSheT6c39XbMHtN9uW1UrsWpu
8KB/cmtmCAiNMA6LSqan1yhmtyd9h+FpgRnZltR56DZHga9oe71VPkWOhWhlAc/j
im6n6Nc7MxXxfoZgbtEONNzFqCqOccr+6JSnxHtJdENYEjlj5AfO0XzEUcHZMMsG
s1Zofns10iOXxS7x1I4nkrORuY+xbpbL8Q4OQDj+nZMBoQYVYv9ZaR3P9j0F98es
5o5DWMTd1MwI6fe2SWXnEz7ZrNbP2jZFzDwn9Xg3bwzdZtLP0FxURmTXlnfH+7b+
zugRFQFaBB1vjPVn5UYMdsPkUIdgtHUZcQkQhpZ2cYcyC6cjv9al2hoa2zPJAyO+
zMowLiZVvBrlJjIH0ZzEOz+TFmADAD6NXTcJHPuoeCtf+mOhw2XeF+IIczdwg7T6
S8gJYSspubspsmyl9B9VEk9JA+n0MO3c7+g82LXmxRoRN+Ve3jEmDsdkFVijT3mQ
KwPvdQLZ5ziTXL297bkcmIfubfG1pGw=
-----END CERTIFICATE-----`,
			[]byte{},
			[]byte{},
			acme.ErrARIIdentifierFailed,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("indx: %d", i), func(t *testing.T) {
			// parse cert
			b, rest := pem.Decode([]byte(tc.pem))
			if len(rest) > 0 {
				t.Errorf("expected no remaining bytes after PEM decode but got %d", len(rest))
			}

			cert, err := x509.ParseCertificate(b.Bytes)
			if err != nil {
				t.Fatalf("failed to parse certificate: %v", err)
			}

			// expected id
			expectedID := ""
			if tc.expectedErr == nil {
				expectedID = string(base64.RawURLEncoding.EncodeToString(tc.expectedAKIBytes) + "." + base64.RawURLEncoding.EncodeToString(tc.expectedSerialBytes))
			}

			// run func
			actualID, actualErr := acme.ACMERenewalInfoIdentifier(cert)

			// check result
			if actualID != expectedID {
				t.Errorf("expected ari id '%s' but got '%s'", expectedID, actualID)
			}

			// check error
			if !test_helpers.ErrorsIs(actualErr, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(actualErr))
			}
		})
	}
}
