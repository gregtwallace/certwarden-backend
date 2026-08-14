package certificates

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"time"
)

// Certificate is a single certificate with all of its fields
type Certificate struct {
	ID                          int
	Name                        string
	Description                 string
	Key                         private_keys.Key
	Account                     acme_accounts.Account
	Subject                     string
	SubjectAltNames             []string
	Organization                string
	OrganizationalUnit          string
	Country                     string
	State                       string
	City                        string
	CSRExtraExtensions          []CertExtension
	PreferredRootCN             string
	LastAccess                  time.Time
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	ApiKey                      string
	ApiKeyNew                   string
	ApiKeyViaUrl                bool
	PostProcessingCommand       string
	PostProcessingEnvironment   []string
	PostProcessingClientAddress string
	PostProcessingClientKeyB64  string
	Profile                     string
}

// certificateSummaryResponse is a JSON response containing only
// fields desired for the summary
type certificateSummaryResponse struct {
	ID              int                               `json:"id"`
	Name            string                            `json:"name"`
	Description     string                            `json:"description"`
	Key             certificateKeySummaryResponse     `json:"private_key"`
	Account         certificateAccountSummaryResponse `json:"acme_account"`
	Subject         string                            `json:"subject"`
	SubjectAltNames []string                          `json:"subject_alts"`
	ApiKeyViaUrl    bool                              `json:"api_key_via_url"`
	LastAccess      int64                             `json:"last_access"`
}

type certificateKeySummaryResponse struct {
	ID        int                  `json:"id"`
	Name      string               `json:"name"`
	Algorithm key_crypto.Algorithm `json:"algorithm"`
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

func (cert *Certificate) summaryResponse() certificateSummaryResponse {
	return certificateSummaryResponse{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		Key: certificateKeySummaryResponse{
			ID:        cert.Key.ID,
			Name:      cert.Key.Name,
			Algorithm: cert.Key.Algorithm,
		},
		Account: certificateAccountSummaryResponse{
			ID:   cert.Account.ID,
			Name: cert.Account.Name,
			CertAccountServer: certificateAccountServerSummaryResponse{
				ID:        cert.Account.AcmeServer.ID,
				Name:      cert.Account.AcmeServer.Name,
				IsStaging: cert.Account.AcmeServer.IsStaging,
			},
		},
		Subject:         cert.Subject,
		SubjectAltNames: cert.SubjectAltNames,
		ApiKeyViaUrl:    cert.ApiKeyViaUrl,
		LastAccess:      cert.LastAccess.Unix(),
	}
}

// certificateDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type certificateDetailedResponse struct {
	certificateSummaryResponse

	Organization                string          `json:"organization"`
	OrganizationalUnit          string          `json:"organizational_unit"`
	Country                     string          `json:"country"`
	State                       string          `json:"state"`
	City                        string          `json:"city"`
	CSRExtraExtensions          []CertExtension `json:"csr_extra_extensions"`
	PreferredRootCN             string          `json:"preferred_root_cn"`
	Profile                     string          `json:"profile"`
	CreatedAt                   int64           `json:"created_at"`
	UpdatedAt                   int64           `json:"updated_at"`
	ApiKey                      string          `json:"api_key"`
	ApiKeyNew                   string          `json:"api_key_new,omitempty"`
	PostProcessingCommand       string          `json:"post_processing_command"`
	PostProcessingEnvironment   []string        `json:"post_processing_environment"`
	PostProcessingClientAddress string          `json:"post_processing_client_address"`
	PostProcessingClientKeyB64  string          `json:"post_processing_client_key"`
}

func (cert *Certificate) detailedResponse() certificateDetailedResponse {
	return certificateDetailedResponse{
		certificateSummaryResponse:  cert.summaryResponse(),
		Organization:                cert.Organization,
		OrganizationalUnit:          cert.OrganizationalUnit,
		Country:                     cert.Country,
		State:                       cert.State,
		City:                        cert.City,
		CSRExtraExtensions:          cert.CSRExtraExtensions,
		PreferredRootCN:             cert.PreferredRootCN,
		Profile:                     cert.Profile,
		CreatedAt:                   cert.CreatedAt.Unix(),
		UpdatedAt:                   cert.UpdatedAt.Unix(),
		ApiKey:                      cert.ApiKey,
		ApiKeyNew:                   cert.ApiKeyNew,
		PostProcessingCommand:       cert.PostProcessingCommand,
		PostProcessingEnvironment:   cert.PostProcessingEnvironment,
		PostProcessingClientAddress: cert.PostProcessingClientAddress,
		PostProcessingClientKeyB64:  cert.PostProcessingClientKeyB64,
	}
}
