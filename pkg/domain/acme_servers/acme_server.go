package acme_servers

import "legocerthub-backend/pkg/acme"

// acmeServer contains details about an ACME server as well as a
// service to access the server
type acmeServer struct {
	acmeServerConfig
	service *acme.Service
}

// acmeServerConfig is the struct to hold a configuration before it
// is turned into a full server/service
type acmeServerConfig struct {
	Name    string `yaml:"name"`
	Value   string `yaml:"db_value"`
	Staging bool   `ymal:"staging"`
	DirUri  string `yaml:"directory_uri"`
}

// TODO: Do something to show if dir requires external binding

// Let's Encrypt ACME Servers
var acmeServersLetsEncrypt = []acmeServerConfig{
	{
		Name:    "Let's Encrypt",
		Value:   "lets_encrypt_prod",
		Staging: false,
		DirUri:  "https://acme-v02.api.letsencrypt.org/directory",
	},
	{
		Name:    "Let's Encrypt (Staging)",
		Value:   "lets_encrypt_staging",
		Staging: true,
		DirUri:  "https://acme-staging-v02.api.letsencrypt.org/directory",
	},
}

// Google Cloud ACME Servers
var acmeServersGoogleCloud = []acmeServerConfig{
	{
		Name:    "Google Cloud",
		Value:   "google_cloud_prod",
		Staging: false,
		DirUri:  "https://dv.acme-v02.api.pki.goog/directory",
	},
	{
		Name:    "Google Cloud (Staging)",
		Value:   "google_cloud_staging",
		Staging: true,
		DirUri:  "https://dv.acme-v02.test-api.pki.goog/directory",
	},
}
