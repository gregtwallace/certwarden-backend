package acme_accounts

import (
	"legocerthub-backend/pkg/domain/private_keys"
)

// a single ACME Account
type Account struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	PrivateKey  *private_keys.Key `json:"private_key"`
	Status      string            `json:"status"`
	Email       string            `json:"email"`
	AcceptedTos bool              `json:"accepted_tos,omitempty"`
	IsStaging   bool              `json:"is_staging"`
	CreatedAt   int               `json:"created_at,omitempty"`
	UpdatedAt   int               `json:"updated_at,omitempty"`
	Kid         string            `json:"kid,omitempty"`
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"available_keys"`
}
