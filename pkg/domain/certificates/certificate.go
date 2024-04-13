package certificates

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/domain/private_keys"
)

// Certificate is a single certificate with all of its fields
type Certificate struct {
	ID                         int
	Name                       string
	Description                string
	CertificateKey             private_keys.Key
	CertificateAccount         acme_accounts.Account
	Subject                    string
	SubjectAltNames            []string
	Organization               string
	OrganizationalUnit         string
	Country                    string
	State                      string
	City                       string
	CSRExtraExtensions         []CertExtension
	CreatedAt                  int
	UpdatedAt                  int
	ApiKey                     string
	ApiKeyNew                  string
	ApiKeyViaUrl               bool
	PostProcessingCommand      string
	PostProcessingEnvironment  []string
	PostProcessingClientKeyB64 string
}

// certificateSummaryResponse is a JSON response containing only
// fields desired for the summary
type certificateSummaryResponse struct {
	ID                 int                               `json:"id"`
	Name               string                            `json:"name"`
	Description        string                            `json:"description"`
	CertificateKey     certificateKeySummaryResponse     `json:"private_key"`
	CertificateAccount certificateAccountSummaryResponse `json:"acme_account"`
	Subject            string                            `json:"subject"`
	SubjectAltNames    []string                          `json:"subject_alts"`
	ApiKeyViaUrl       bool                              `json:"api_key_via_url"`
}

type certificateKeySummaryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type certificateAccountSummaryResponse struct {
	ID                int                                     `json:"id"`
	Name              string                                  `json:"name"`
	CertAccountServer certificateAccountServerSummaryResponse `json:"acme_server"`
}

type certificateAccountServerSummaryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsStaging bool   `json:"is_staging"`
}

func (cert Certificate) summaryResponse() certificateSummaryResponse {
	return certificateSummaryResponse{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		CertificateKey: certificateKeySummaryResponse{
			ID:   cert.CertificateKey.ID,
			Name: cert.CertificateKey.Name,
		},
		CertificateAccount: certificateAccountSummaryResponse{
			ID:   cert.CertificateAccount.ID,
			Name: cert.CertificateAccount.Name,
			CertAccountServer: certificateAccountServerSummaryResponse{
				ID:        cert.CertificateAccount.AcmeServer.ID,
				Name:      cert.CertificateAccount.AcmeServer.Name,
				IsStaging: cert.CertificateAccount.AcmeServer.IsStaging,
			},
		},
		Subject:         cert.Subject,
		SubjectAltNames: cert.SubjectAltNames,
		ApiKeyViaUrl:    cert.ApiKeyViaUrl,
	}
}

// certificateDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type certificateDetailedResponse struct {
	certificateSummaryResponse
	Organization               string              `json:"organization"`
	OrganizationalUnit         string              `json:"organizational_unit"`
	Country                    string              `json:"country"`
	State                      string              `json:"state"`
	City                       string              `json:"city"`
	CSRExtraExtensions         []CertExtensionJSON `json:"csr_extra_extensions"`
	CreatedAt                  int                 `json:"created_at"`
	UpdatedAt                  int                 `json:"updated_at"`
	ApiKey                     string              `json:"api_key"`
	ApiKeyNew                  string              `json:"api_key_new,omitempty"`
	PostProcessingCommand      string              `json:"post_processing_command"`
	PostProcessingEnvironment  []string            `json:"post_processing_environment"`
	PostProcessingClientKeyB64 string              `json:"post_processing_client_key"`
}

func (cert Certificate) detailedResponse() certificateDetailedResponse {
	// convert extensions to json output obj
	extraExtensions := []CertExtensionJSON{}
	for i := range cert.CSRExtraExtensions {
		oneExt := cert.CSRExtraExtensions[i].toJSONObj()
		extraExtensions = append(extraExtensions, oneExt)
	}

	return certificateDetailedResponse{
		certificateSummaryResponse: cert.summaryResponse(),
		Organization:               cert.Organization,
		OrganizationalUnit:         cert.OrganizationalUnit,
		Country:                    cert.Country,
		State:                      cert.State,
		City:                       cert.City,
		CSRExtraExtensions:         extraExtensions,
		CreatedAt:                  cert.CreatedAt,
		UpdatedAt:                  cert.UpdatedAt,
		ApiKey:                     cert.ApiKey,
		ApiKeyNew:                  cert.ApiKeyNew,
		PostProcessingCommand:      cert.PostProcessingCommand,
		PostProcessingEnvironment:  cert.PostProcessingEnvironment,
		PostProcessingClientKeyB64: cert.PostProcessingClientKeyB64,
	}
}

// NewOrderPayload creates the appropriate newOrder payload for ACME
func (cert *Certificate) NewOrderPayload() acme.NewOrderPayload {
	var identifiers []acme.Identifier

	// subject is always required and should be first
	// dns is the only supported type and is hardcoded
	identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: cert.Subject})

	// add alt names if they exist
	if cert.SubjectAltNames != nil {
		for _, name := range cert.SubjectAltNames {
			identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: name})
		}
	}

	return acme.NewOrderPayload{
		Identifiers: identifiers,
	}
}
